package controllers

import (
	"app/helpers"
	"app/models"

	"github.com/gofiber/fiber/v2"
)

func NavControllers(app *fiber.App){
    app.Get("/partials/nav/decks-opened", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.Query("active-deck-id", "")

        db := helpers.ConnectDB()
        decks := models.GetDecks(db, activeDeckId, user.Id)

        c.Cookie(&fiber.Cookie{
            Name: "nav_closed",
            Value: "false",
            Secure: true,
            HTTPOnly: true,
            SameSite: "Strict",
        })

        return c.Render("partials/nav/decks-opened", fiber.Map{
            "Decks": decks,
            "ActiveDeckId": activeDeckId,
        })
    })

    app.Get("/partials/nav/decks-closed", func(c *fiber.Ctx) error {

        activeDeckId := c.Query("active-deck-id")

        c.Cookie(&fiber.Cookie{
            Name: "nav_closed",
            Value: "true",
            Secure: true,
            HTTPOnly: true,
            SameSite: "Strict",
        })

        return c.Render("partials/nav/decks-closed", fiber.Map{
            "ActiveDeckId": activeDeckId,
        })
    })

    app.Get("/partials/nav/decks/:id", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")

        db := helpers.ConnectDB()
        deck := models.GetDeck(db, deckId, user.Id)

        deckMetadata := models.GetDeckMetadata(db, deckId)

        return c.Render("partials/nav/deck-link", fiber.Map{
            "Id": deck.Id,
            "Label": deck.Label,
            "CardCount": deckMetadata.CardCount,
            "Active": deck.Active,
        })
    })
}
