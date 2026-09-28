package main

import (
	"app/controllers"
	"app/helpers"
	"app/models"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/google/uuid"
)

func main() {
	if err := helpers.InitDB(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Before anything serves: the schema a request is answered against should
	// be the one this build was written for.
	if err := helpers.RunMigrations(); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// A missing or malformed key leaves client nil. Only the sign in flow uses
	// it, so the rest of the app still runs without one, but say so here
	// rather than letting it surface as a nil dereference on /authorize.
	client, err := clerk.NewClient(os.Getenv("CLERK_API_KEY"))
	if err != nil {
		log.Error("Clerk client unavailable, sign in is disabled", "error", err)
	}

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:             engine,
		BodyLimit:         1024 * 1024 * 100,
		StreamRequestBody: true,
		ReadTimeout:       120 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       180 * time.Second,
	})

	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := helpers.PingDB(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).SendString("database unavailable")
		}
		return c.SendString("ok")
	})

	// Fiber only sends Cache-Control when MaxAge is set, so these were served
	// with no caching directive at all. Build output is not content-hashed, so
	// keep css/js short enough that a deploy is picked up quickly.
	app.Static("/css", "./public/css", fiber.Static{MaxAge: 3600})
	app.Static("/js", "./public/js", fiber.Static{MaxAge: 3600})
	app.Static("/static", "./public/static", fiber.Static{MaxAge: 86400})

	controllers.HomepageControllers(app)
	controllers.DeckEditorControllers(app)
	// Registered before the deck manager so /decks/import is not swallowed by
	// its /decks/:id route, the same reason /decks/new is registered first.
	controllers.ImportDeckControllers(app)
	controllers.DeckManagerControllers(app)
	controllers.NavControllers(app)
	controllers.PlayControllers(app)
	controllers.DeckStatsControllers(app)

	app.Get("/register", func(c *fiber.Ctx) error {
		return c.Render("pages/register/index", fiber.Map{})
	})
	app.Get("/sign-in", func(c *fiber.Ctx) error {
		return c.Render("pages/sign-in/index", fiber.Map{})
	})
	app.Get("/sign-out", func(c *fiber.Ctx) error {
		if sessionId := c.Cookies("session_id", ""); sessionId != "" {
			if err := models.DeleteSession(helpers.ConnectDB(), sessionId); err != nil {
				log.Error("Failed to delete session on sign out", "error", err)
			}
		}
		c.ClearCookie("session_id")
		return c.Render("pages/sign-out/index", fiber.Map{})
	})
	app.Get("/authorize", func(c *fiber.Ctx) error {
		token := c.Cookies("__session", "")
		if token == "" {
			return c.Redirect("/sign-in")
		}
		if client == nil {
			log.Error("Cannot verify session, CLERK_API_KEY is not set")
			return c.Redirect("/sign-in")
		}
		log.Info("Verifying user session")
		sessClaims, err := client.VerifyToken(token)
		if err != nil {
			log.Error("Failed to verify session", "error", err)
			return c.Redirect("/sign-in")
		}
		user, err := client.Users().Read(sessClaims.Claims.Subject)
		if err != nil {
                        log.Error("Failed to get user from Clerk.", "error", err)
			return c.Redirect("/sign-in")
		}

		email := ""
		if len(user.EmailAddresses) > 0 {
			email = user.EmailAddresses[0].EmailAddress
		}

		username := ""
		if user.Username != nil {
			username = *user.Username
		} else {
			username = strings.TrimPrefix(user.ID, "user_")
		}

		customUser := models.User{
			Id:       user.ID,
			Username: username,
			Email:    email,
			Avatar:   user.ProfileImageURL,
		}
		sessionId := uuid.New().String()
		sessionId = strings.ReplaceAll(sessionId, "-", "")
		expires := time.Now().Add(168 * time.Hour)
		blob, err := models.UserToBlob(customUser)
		if err != nil {
			log.Error("Failed to encode user session", "error", err)
			return c.Redirect("/sign-in")
		}

		db := helpers.ConnectDB()
		if res := db.Exec("INSERT INTO Sessions (session_id, user_id, data, expires) VALUES (UNHEX(?), ?, ?, ?)", sessionId, customUser.Id, blob, expires); res.Error != nil {
			log.Error("Failed to store user session", "error", res.Error)
			return c.Redirect("/sign-in")
		}

		c.Cookie(&fiber.Cookie{
			Name:     "session_id",
			Value:    sessionId,
			Expires:  expires,
			Secure:   true,
			HTTPOnly: true,
			SameSite: "Strict",
		})

		postLoginRedirect := c.Cookies("post_login_redirect", "/")
		c.ClearCookie("post_login_redirect")

		return c.Redirect(postLoginRedirect)
	})

	app.Get("/privacy-policy", func(c *fiber.Ctx) error {
		user, _ := helpers.GetUserFromSession(c)
		db := helpers.ConnectDB()

		var decks []models.Deck
		if c.Cookies("nav_closed", "") != "true" {
			decks = models.GetDecks(db, "", user.Id)
		}

		return c.Render("pages/privacy-policy/index", fiber.Map{
			"Page":  "privacy-policy",
			"Decks": decks,
		}, "layouts/main")
	})

	go purgeExpiredSessions()

	// Drain in-flight requests on redeploy instead of cutting them.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Info("Shutting down")
		if err := app.ShutdownWithTimeout(20 * time.Second); err != nil {
			log.Error("Shutdown failed", "error", err)
		}
	}()

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

// purgeExpiredSessions drops rows past their expiry. The lookup already filters
// on expires, so this is purely to stop the table growing without bound.
func purgeExpiredSessions() {
	for {
		if n, err := models.DeleteExpiredSessions(helpers.ConnectDB()); err != nil {
			log.Error("Failed to purge expired sessions", "error", err)
		} else if n > 0 {
			log.Info("Purged expired sessions", "count", n)
		}
		time.Sleep(time.Hour)
	}
}
