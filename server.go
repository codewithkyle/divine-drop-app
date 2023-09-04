package main

import (
    "log"
    "os"
    "strings"
    "net/url"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/template/html/v2"
    "github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

func main() {
    // Load environment variables from file.
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

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

    app.Get("/partials/nav/decks-opened", func(c *fiber.Ctx) error {
        return c.Render("partials/nav/decks-opened", fiber.Map{})
    })
    app.Get("/partials/nav/decks-closed", func(c *fiber.Ctx) error {
        return c.Render("partials/nav/decks-closed", fiber.Map{})
    })

    app.Listen(":3000")
}
