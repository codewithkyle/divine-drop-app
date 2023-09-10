package controllers

import (
	"net/url"
	"strconv"

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

        search := c.Query("search")
        sort := c.Query("sort")

        decks := models.GetDecks(db, deckId, user.Id)
        deckCards := models.SearchDeckCards(db, deckId, search, sort)
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

        for i := range deckCards {
            deckCards[i].IsCommander = deckCards[i].CardId == deck.CommanderCardId
            deckCards[i].IsOathbreaker = deckCards[i].CardId == deck.OathbreakerCardId
        }

        return c.Render("pages/deck-manager/index", fiber.Map{
            "Page": "deck-editor",
            "User": user,
            "Deck": deck,
            "Decks": decks,
            "Cards": deckCards,
            "ActiveDeckId": deckId,
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
            "Sort": sort,
            "Search": search,
        }, "layouts/main")
    }) 

    app.Get("/partials/deck-manager/card-grid/:id", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        search := c.Query("search")
        sort := c.Query("sort")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        cards := models.SearchDeckCards(db, deckId, search, sort)

        url := "/decks/" + deckId + "?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort)
        c.Response().Header.Set("Hx-Replace-Url", url)

        return c.Render("partials/deck-manager/card-grid", fiber.Map{
            "Cards": cards,
        })
    })

    app.Post("/decks/:id/commander/:cardId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        cardId := c.Params("cardId")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        if deck.CommanderCardId == cardId {
            db.Exec("UPDATE Decks SET commander_card_id = NULL WHERE id = UNHEX(?)", deckId)
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Commander removed\"}")
            return c.SendStatus(200)
        }

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET commander_card_id = UNHEX(?) WHERE id = UNHEX(?)", cardId, deckId)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + helpers.EscapeString(card.Name) + " is now the Commander\"}")

        return c.SendStatus(200)
    })

    app.Post("/decks/:id/oathbreaker/:cardId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        cardId := c.Params("cardId")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        if deck.OathbreakerCardId == cardId {
            db.Exec("UPDATE Decks SET oathbreaker_card_id = NULL WHERE id = UNHEX(?)", deckId)
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Oathbreaker removed\"}")
            return c.SendStatus(200)
        }

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET oathbreaker_card_id = UNHEX(?) WHERE id = UNHEX(?)", cardId, deckId)
        
        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + helpers.EscapeString(card.Name) + " is now the Oathbreaker\"}")

        return c.SendStatus(200)
    })

    app.Patch("/decks/:deckId/cards/:cardId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        cardId := c.Params("cardId")
        qty := c.FormValue("qty")
        newQty, err := strconv.Atoi(qty)
        if err != nil {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Invalid quantity.\"}")
            return c.SendStatus(400)
        }
        if newQty < 1 {
            newQty = 1
        }

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Deck_Cards SET qty = ? WHERE deck_id = UNHEX(?) AND card_id = UNHEX(?)", newQty, deckId, cardId)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + card.Name + " quantity updated\"}")

        return c.SendStatus(200)
    })
}
