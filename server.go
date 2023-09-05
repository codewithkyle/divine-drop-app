package main

import (
    "log"
    "os"
    "strings"
    "time"
    "net/url"
    "errors"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/template/html/v2"
    "github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
    "github.com/clerkinc/clerk-sdk-go/clerk"
    "github.com/google/uuid"
    "app/models"
)

func connectDB() *gorm.DB {
    db, err := gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("failed to connect to PlanetScale: %v", err)
	}
    return db;
}

func getUser(sessionId string) (models.User, error) {
    if sessionId == "" {
        return models.User{}, errors.New("Session not found");
    }
    db := connectDB()
    var session models.Session
    db.Raw("SELECT * FROM Sessions WHERE session_id = UNHEX(?) AND expires > ?", sessionId, time.Now()).Scan(&session)
    if session.Id == "" {
        return models.User{}, errors.New("Session not found");
    }
    return models.BlobToUser(session.Data)
}

func main() {
    // Load environment variables from file.
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

    client, _ := clerk.NewClient(os.Getenv("CLERK_API_KEY"))

    // Create a new engine
    engine := html.New("./views", ".html")

    // Pass the engine to the Views
    app := fiber.New(fiber.Config{
        Views: engine,
    })

    app.Static("/css", "./public/css")
    app.Static("/js", "./public/js")
    app.Static("/fonts", "./public/fonts")

    app.Get("/", func(c *fiber.Ctx) error {
        var cards []models.Card
        db := connectDB()
        db.Raw("SELECT C.front FROM Cards AS C JOIN Card_Names AS CN ON C.id = CN.card_id WHERE CN.name LIKE ? LIMIT 20 OFFSET 0", "%%").Scan(&cards)

        return c.Render("pages/card-browser/index", fiber.Map{
            "Page": "card-browser",
            "Cards": cards,
            "Search": "",
            "NextPage": 1,
        }, "layouts/main")
    })
    app.Post("/partials/card-browser", func(c *fiber.Ctx) error {
        search := c.FormValue("search")

        searchQuery := "%" + strings.Trim(search, " ") + "%"

        var cards []models.Card
        db := connectDB()
        db.Raw("SELECT C.front FROM Cards AS C JOIN Card_Names AS CN ON C.id = CN.card_id WHERE CN.name LIKE ? LIMIT 20 OFFSET 0", searchQuery).Scan(&cards)

        return c.Render("pages/card-browser/index", fiber.Map{
            "Cards": cards,
            "Search": url.QueryEscape(search),
            "NextPage": 1,
            "SearchRaw": search,
        })
    })
    app.Get("/partials/card-browser", func(c *fiber.Ctx) error {
        search := c.Query("search")
        page := c.QueryInt("page")

        searchQuery := "%" + strings.Trim(search, " ") + "%"
        var offset = page * 20

        var cards []models.Card
        db := connectDB()
        db.Raw("SELECT C.front FROM Cards AS C JOIN Card_Names AS CN ON C.id = CN.card_id WHERE CN.name LIKE ? LIMIT 20 OFFSET ?", searchQuery, offset).Scan(&cards)

        return c.Render("partials/card-browser", fiber.Map{
            "Cards": cards,
            "Search": url.QueryEscape(search),
            "NextPage": page + 1,
        })
    })

    app.Get("/decks/new", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "") 
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        uuid := uuid.New().String()
        uuid = strings.ReplaceAll(uuid, "-", "")

        db := connectDB()
        db.Exec("INSERT INTO Decks (id, user_id, label) VALUES (UNHEX(?), ?, 'Untitled')", uuid, user.Id)

        return c.Redirect("/decks/" + uuid + "/edit");
    })
    app.Get("/decks/:id/edit", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        db := connectDB()
        var deck models.Deck
        db.Raw("SELECT HEX(D.id) AS id, HEX(D.commander_card_id) AS commander_card_id, D.label, D.user_id, (SELECT COUNT(*) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) AND D.user_id = ?", deckId, user.Id).Scan(&deck)

        if deck.Id == "" {
            return c.Redirect("/")
        }

        return c.Render("pages/deck-builder/index", fiber.Map{
            "Page": "deck-builder",
            "User": user,
            "Deck": deck,
        }, "layouts/main")
    })
    app.Patch("/decks/:id", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        db := connectDB()
        label := c.FormValue("label")
        db.Exec("UPDATE Decks SET label = ? WHERE id = UNHEX(?) AND user_id = ?", label, deckId, user.Id)

        deck := models.Deck{}
        db.Raw("SELECT HEX(id) AS id, label, HEX(commander_card_id) AS commander_card_id, user_id FROM Decks WHERE id = UNHEX(?) AND user_id = ?", deckId, user.Id).Scan(&deck)

        return c.Render("partials/deck-builder/label-input", fiber.Map{
            "Deck": deck,
        })
    })

    app.Get("/partials/nav/decks-opened", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }
        db := connectDB()
        var decks []models.Deck
        db.Raw("SELECT HEX(D.id) AS id, HEX(D.commander_card_id) AS commander_card_id, D.label, D.user_id, (SELECT COUNT(*) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE user_id = ?", user.Id).Scan(&decks)

        return c.Render("partials/nav/decks-opened", fiber.Map{
            "Decks": decks,
        })
    })
    app.Get("/partials/nav/decks-closed", func(c *fiber.Ctx) error {
        return c.Render("partials/nav/decks-closed", fiber.Map{})
    })

    app.Get("/register", func(c *fiber.Ctx) error {
        return c.Render("pages/register/index", fiber.Map{})
    })
    app.Get("/sign-in", func(c *fiber.Ctx) error {
        return c.Render("pages/sign-in/index", fiber.Map{})
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

        // TODO: insert into DB
        db := connectDB()
        db.Exec("INSERT INTO Sessions (session_id, user_id, data, expires) VALUES (UNHEX(?), ?, ?, ?)", sessionId, customUser.Id, blob, expires)

        c.Cookie(&fiber.Cookie{
            Name: "session_id",
            Value: sessionId,
            Expires: expires,
            Secure: true,
            HTTPOnly: true,
            SameSite: "Strict",
        })

        return c.Redirect("/")
    })

    app.Listen(":3000")
}
