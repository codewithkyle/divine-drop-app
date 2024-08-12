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

        return c.Render("pages/play/index", fiber.Map{
            "Page": "play",
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
        }, "layouts/main")
    })
}
