package controllers

import (
	"app/helpers"
    "app/models"

	"github.com/gofiber/fiber/v2"
)

func PlayControllers(app *fiber.App){
    app.Get("/play", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        db := helpers.ConnectDB()

        var decks []models.Deck
        if (c.Cookies("nav_closed", "") != "true") {
            decks = models.GetDecks(db, "", user.Id)
        }

        return c.Render("pages/play/index", fiber.Map{
            "Page": "play",
            "Decks": decks,
        }, "layouts/main")
    })

    app.Get("/vtt", func(c *fiber.Ctx) error {
        return c.Render("pages/play/vtt", fiber.Map{}, "layouts/vtt")
    });

    app.Get("/vtt/:gameId", func(c *fiber.Ctx) error {
        return c.Render("pages/play/vtt", fiber.Map{}, "layouts/vtt")
    });
}
