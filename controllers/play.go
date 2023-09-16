package controllers

import (
	"app/helpers"

	"github.com/gofiber/fiber/v2"
)

func PlayControllers(app *fiber.App){
    app.Get("/play", func(c *fiber.Ctx) error {
        _, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        return c.Render("pages/play/index", fiber.Map{
            "Page": "play",
        }, "layouts/main")
    })

    app.Get("/vtt", func(c *fiber.Ctx) error {
        return c.Render("pages/play/vtt", fiber.Map{}, "layouts/vtt")
    });

    app.Get("/vtt/:gameId", func(c *fiber.Ctx) error {
        return c.Render("pages/play/vtt", fiber.Map{}, "layouts/vtt")
    });
}
