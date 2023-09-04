package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/template/html/v2"
)

func main() {
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
        // Render index within layouts/main
        return c.Render("index", fiber.Map{
            "Title": "Hello, World!",
        })
    })

    app.Listen(":3000")
}
