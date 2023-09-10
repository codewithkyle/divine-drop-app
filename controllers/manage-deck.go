package controllers

import (
	"net/url"

	"github.com/gofiber/fiber/v2"

	"app/helpers"
	"app/models"
)

func DeckManagerControllers(app *fiber.App){
    app.Get("/decks/:id", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        decks := models.GetDecks(db, deckId, user.Id)
        deckCards := models.GetDeckCards(db, deckId)
        deckMetadata := models.GetDeckMetadata(db, deckId)

        mythicsCount := models.GetMythicsCount(db, deckId)
        uncommonsCount := models.GetUncommonsCount(db, deckId)
        commonsCount := models.GetCommonsCount(db, deckId)
        raresCount := deckMetadata.CardCount - mythicsCount - uncommonsCount - commonsCount

        landCount := models.GetLandCount(db, deckId)

        containsW, containsU, containsB, containsR, containsG := models.GetDeckColors(db, deckId)

        bannerArt := ""
        if deck.CommanderCardId != "" {
            bannerArt = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + deck.CommanderCardId +  "-art.png"
        } else if len(deckCards) > 0 {
            bannerArt = deckCards[len(deckCards) - 1].Art
        }

        return c.Render("pages/deck-manager/index", fiber.Map{
            "Page": "deck-editor",
            "User": user,
            "Deck": deck,
            "Decks": decks,
            "ActiveDeckId": deckId,
            "SearchPage": 1,
            "DeckCards": deckCards,
            "DeckCardsCount": len(deckCards),
            "DeckMetadata": deckMetadata,
            "ContainsW": containsW,
            "ContainsU": containsU,
            "ContainsB": containsB,
            "ContainsR": containsR,
            "ContainsG": containsG,
            "BannerArtUrl": url.QueryEscape(bannerArt),
            "MythicsCount": mythicsCount,
            "UncommonsCount": uncommonsCount,
            "CommonsCount": commonsCount,
            "RaresCount": raresCount,
            "LandCount": landCount,
        }, "layouts/main")
    }) 
}
