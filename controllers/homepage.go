package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"

	"net/url"

	"app/helpers"
	"app/models"
)

func HomepageControllers(app *fiber.App) {
    app.Get("/", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            log.Error("Failed to get user session", "error", err)
        }

        search := c.Query("search")

        db := helpers.ConnectDB()
        cards := models.SearchCardsByName(db, search, 0, 20)

        deckGroups := models.GetDeckGroups(db, user.Id)
        decks := models.GetDecks(db, "", user.Id)

        groupedDecks := make(map[string]*GroupedDecks)
        ungroupedDecks := []models.Deck{}

        for i := range deckGroups {
            groupedDecks[deckGroups[i].Id] = &GroupedDecks{ Id: deckGroups[i].Id, Label: deckGroups[i].Label, Decks: []models.Deck{} }
        }

        for i := range decks {
            if decks[i].GroupId != "" {
                if value, ok := groupedDecks[decks[i].GroupId]; ok {
                    value.Decks = append(value.Decks, decks[i])
                } 
            } else {
                ungroupedDecks = append(ungroupedDecks, decks[i])
            }
        }

        return c.Render("pages/card-browser/index", fiber.Map{
            "Page": "card-browser",
            "Cards": cards,
            "Search": "",
            "NextPage": 0,
            "User": user,
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "SearchRaw": search,
            "NavClosed": c.Cookies("nav_closed", "") == "true" || len(decks) == 0 || user.Id == "",
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
            "NextPage": 0,
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
