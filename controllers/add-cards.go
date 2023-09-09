package controllers

import (
	"github.com/gofiber/fiber/v2"

    "net/url"

	"app/helpers"
	"app/models"
)

func AllCardsControllers(app *fiber.App) {
    app.Get("/", func(c *fiber.Ctx) error {
        user, _ := helpers.GetUserFromSession(c)

        search := c.Query("search")

        db := helpers.ConnectDB()
        cards := models.SearchCardsByName(db, search, 0, 20)

        decks := models.GetDecks(db, "", user.Id)

        return c.Render("pages/card-browser/index", fiber.Map{
            "Page": "card-browser",
            "Cards": cards,
            "Search": "",
            "NextPage": 1,
            "User": user,
            "Decks": decks,
            "SearchRaw": search,
        }, "layouts/main")
    })
    app.Post("/partials/card-browser", func(c *fiber.Ctx) error {
        search := c.FormValue("search")

        db := helpers.ConnectDB()
        cards := models.SearchCardsByName(db, search, 0, 20)

        c.Response().Header.Set("HX-Replace-Url", "/?search=" + url.QueryEscape(search))

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

        var offset = page * 20

        db := helpers.ConnectDB()
        cards := models.SearchCardsByName(db, search, offset, 20)

        if len(cards) > 0 {
            c.Response().Header.Set("HX-Trigger", "cardBrowserChanged")
        }

        return c.Render("partials/card-browser/card-grid", fiber.Map{
            "Cards": cards,
            "Search": url.QueryEscape(search),
            "NextPage": page + 1,
        })
    })
    app.Get("/partials/card-browser/page-input", func(c *fiber.Ctx) error {
        page := c.QueryInt("page")

        return c.Render("partials/card-browser/page-input", fiber.Map{
            "NextPage": page + 1,
        })
    })
}
