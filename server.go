package main

import (
	"app/controllers"
	"app/helpers"
	"app/models"
	"log"
	"os"
	"strings"
	"time"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

    client, _ := clerk.NewClient(os.Getenv("CLERK_API_KEY"))

    engine := html.New("./views", ".html")
    app := fiber.New(fiber.Config{
        Views: engine,
    })

    app.Static("/css", "./public/css")
    app.Static("/js", "./public/js")
    app.Static("/static", "./public/static")

    controllers.HomepageControllers(app)
    controllers.DeckEditorControllers(app)
    controllers.NavControllers(app)

    app.Get("/register", func(c *fiber.Ctx) error {
        return c.Render("pages/register/index", fiber.Map{})
    })
    app.Get("/sign-in", func(c *fiber.Ctx) error {
        return c.Render("pages/sign-in/index", fiber.Map{})
    })
    app.Get("/sign-out", func(c *fiber.Ctx) error {
        c.ClearCookie("session_id")
        return c.Render("pages/sign-out/index", fiber.Map{})
    })
    app.Get("/authorize", func(c *fiber.Ctx) error {
        token := c.Cookies("__session", "")
        if token == "" {
            return c.Redirect("/sign-in")
        }
        sessClaims, err := client.VerifyToken(token)
        if err != nil {
            return c.Redirect("/sign-in")
        }
        user, err := client.Users().Read(sessClaims.Claims.Subject)
		if err != nil {
            return c.Redirect("/sign-in")
		}

        email := ""
        if (len(user.EmailAddresses) > 0) {
            email = user.EmailAddresses[0].EmailAddress
        }

        username := ""
        if (user.Username != nil) {
            username = *user.Username
        } else {
            username = strings.Trim(user.ID, "user_")
        }

        customUser := models.User{
            Id: user.ID,
            Username: username,
            Email: email,
            Avatar: user.ProfileImageURL,
        }
        sessionId := uuid.New().String()
        sessionId = strings.ReplaceAll(sessionId, "-", "")
        expires := time.Now().Add(168 * time.Hour)
        blob, _ := models.UserToBlob(customUser)

        db := helpers.ConnectDB()
        db.Exec("INSERT INTO Sessions (session_id, user_id, data, expires) VALUES (UNHEX(?), ?, ?, ?)", sessionId, customUser.Id, blob, expires)

        c.Cookie(&fiber.Cookie{
            Name: "session_id",
            Value: sessionId,
            Expires: expires,
            Secure: true,
            HTTPOnly: true,
            SameSite: "Strict",
        })

        postLoginRedirect := c.Cookies("post_login_redirect", "/")
        c.ClearCookie("post_login_redirect")

        return c.Redirect(postLoginRedirect)
    })

    app.Get("/privacy-policy", func(c *fiber.Ctx) error {
        return c.Render("pages/privacy-policy/index", fiber.Map{}, "layouts/main")
    })

    app.Listen(":3000")
}
