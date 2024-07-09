package controllers

import (
	"math/rand"
	"net/url"
	"strconv"
	"strings"
        "os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

        "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
        "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

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
        filter := c.Query("filter")
        rarity := c.Query("rarity")

        decks := models.GetDecks(db, deckId, user.Id)
        deckCards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity)
        deckMetadata := models.GetDeckMetadata(db, deckId)

        mythicsCount := models.GetMythicsCount(db, deckId)
        uncommonsCount := models.GetUncommonsCount(db, deckId)
        commonsCount := models.GetCommonsCount(db, deckId)
        raresCount := deckMetadata.CardCount - mythicsCount - uncommonsCount - commonsCount

        landCount := models.GetLandCount(db, deckId)
        sideboardCount := models.GetSideboardCount(db, deckId)

        containsW, containsU, containsB, containsR, containsG := models.GetDeckColors(db, deckId)

        bannerArt := ""
        if deck.CommanderCardId != "" {
            bannerArt = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(deck.CommanderCardId) +  "-art.png"
        } else if len(deckCards) > 0 {
            bannerArt = deckCards[len(deckCards) - 1].Art
        }

        for i := range deckCards {
            deckCards[i].IsCommander = deckCards[i].CardId == deck.CommanderCardId
            deckCards[i].IsOathbreaker = deckCards[i].CardId == deck.OathbreakerCardId
            if deckCards[i].Print != 0 {
                printDate := strconv.Itoa(deckCards[i].Print)
                deckCards[i].Front = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(deckCards[i].CardId) + "-" + printDate +  "-front.png"
                if deckCards[i].Back != "" {
                    deckCards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(deckCards[i].CardId) + "-" + printDate +  "-back.png"
                }
            }
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
            "SideboardCount": sideboardCount,
            "Sort": sort,
            "Search": search,
            "Filter": filter,
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
        filter := c.Query("filter")
        rarity := c.Query("rarity")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        cards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity)

        for i := range(cards) {
            if cards[i].CardId == deck.CommanderCardId {
                cards[i].IsCommander = true
            }
            if cards[i].CardId == deck.OathbreakerCardId {
                cards[i].IsOathbreaker = true
            }

            if cards[i].Print != 0 {
                printDate := strconv.Itoa(cards[i].Print)
                cards[i].Front = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-front.png"
                if cards[i].Back != "" {
                    cards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-back.png"
                }
            }
        }

        url := "/decks/" + deckId + "?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&filter=" + url.QueryEscape(filter)
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
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Commander removed\", \"deckUpdated\": \"" + deckId + "\"}")
            return c.SendStatus(200)
        }

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET commander_card_id = UNHEX(?) WHERE id = UNHEX(?)", cardId, deckId)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + helpers.EscapeString(card.Name) + " is now the Commander\", \"bannerArtUpdate\": \"" + card.Art + "\", \"deckUpdated\": \"" + deckId + "\"}")

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
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Oathbreaker removed\", \"deckUpdated\": \"" + deckId + "\"}")
            return c.SendStatus(200)
        }

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET oathbreaker_card_id = UNHEX(?) WHERE id = UNHEX(?)", cardId, deckId)
        
        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + helpers.EscapeString(card.Name) + " is now the Oathbreaker\", \"deckUpdated\": \"" + deckId + "\"}")

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

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + card.Name + " quantity updated\", \"deckUpdated\": \"" + deckId + "\"}")

        return c.SendStatus(200)
    })

    app.Delete("/decks/:deckId/cards/:cardId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        cardId := c.Params("cardId")

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

        db.Exec("DELETE FROM Deck_Cards WHERE deck_id = UNHEX(?) AND card_id = UNHEX(?)", deckId, cardId)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + card.Name + " removed from deck\", \"deckUpdated\": \"" + deckId + "\"}")

        return c.SendStatus(200)
    })

    app.Get("/partials/deck-manager/:deckId/quick-grid", func(c *fiber.Ctx) error {

        name := c.Query("quick-search")
        deckId := c.Params("deckId")
        pageStr := c.Query("page")
        page, _ := strconv.Atoi(pageStr)
        page += 1

        db := helpers.ConnectDB()

        cards := models.SearchCardsByName(db, name, page, 50)

        for i := range cards {
            cards[i].ActiveDeckId = deckId
        }

        return c.Render("partials/deck-manager/quick-card-grid", fiber.Map{
            "Cards": cards,
            "ActiveDeckId": deckId,
        })
    })

    app.Put("/decks/:deckId/cards/:cardId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        cardId := c.Params("cardId")

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

        deckCard := models.GetDeckCard(db, deckId, cardId)
        if deckCard.Id == "" {
            newCardId := uuid.New().String()
            newCardId = strings.ReplaceAll(newCardId, "-", "")
            db.Exec("INSERT INTO Deck_Cards (id, deck_id, card_id, qty) VALUES (UNHEX(?), UNHEX(?), UNHEX(?), 1)", newCardId, deckId, cardId)
        } else {
            db.Exec("UPDATE Deck_Cards SET qty = qty + 1 WHERE deck_id = UNHEX(?) AND id = UNHEX(?)", deckId, deckCard.Id)
        }

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + card.Name + " added to deck\", \"deckUpdated\": \"" + deckId + "\", \"addedCard\": \"\"}")

        return c.SendStatus(200)
    })

    app.Get("/partials/deck-manager/simulate-draw", func(c *fiber.Ctx) error {
        _, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Query("id")
        db := helpers.ConnectDB()
        cards := models.GetDeckCards(db, deckId)

        deckCards := []models.DeckCard{}
        for i := range cards {
            if !cards[i].InSideboard {
                if cards[i].Print != 0 {
                    printDate := strconv.Itoa(cards[i].Print)
                    cards[i].Front = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-front.png"
                    if cards[i].Back != "" {
                        cards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-back.png"
                    }
                }
                for j := uint8(0); j < cards[i].Qty; j++ {
                    deckCards = append(deckCards, cards[i])
                }
            }
        }

        if len(deckCards) >= 7 {
            // Bogo sort ftw
            for i := range deckCards {
                j := rand.Intn(i + 1)
                deckCards[i], deckCards[j] = deckCards[j], deckCards[i]
            }
            
            hand := []models.DeckCard{}
            for i := 0; i < 7; i++ {
                hand = append(hand, deckCards[i])
            }
            deckCards = hand
        }

        return c.Render("partials/deck-manager/simulate-draw", fiber.Map{
            "Cards": deckCards,
        })
    })


   app.Get("/api/v1/decks/:id", func(c *fiber.Ctx) error {
        deckId := c.Params("id")
        if deckId == "" {
            return c.SendStatus(400)
        }
        db := helpers.ConnectDB()
        cards := models.GetDeckCardsMetadata(db, deckId)
        deck := models.GetDeckByID(db, deckId)

        deckCards := []models.DeckCardMetadata{}
        for i := range cards {
            if cards[i].Print != 0 {
                printDate := strconv.Itoa(cards[i].Print)
                cards[i].Front = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-front.png"
                if cards[i].Back != "" {
                    cards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-back.png"
                }
            }
            if cards[i].Back == "" {
                cards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/back.png"
            }
            if cards[i].CardId == deck.CommanderCardId {
                cards[i].IsCommander = true
            }
            if cards[i].CardId == deck.OathbreakerCardId {
                cards[i].IsOathbreaker = true
            }
            deckCards = append(deckCards, cards[i])
        }

        return c.JSON(deckCards)
    })

    app.Put("/decks/:id/sideboard/:cardId", func(c *fiber.Ctx) error {
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

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Deck_Cards SET sideboard = 1 WHERE card_id = UNHEX(?) AND deck_id = UNHEX(?)", card.Id, deck.Id)

        if deck.CommanderCardId == cardId {
            db.Exec("UPDATE Decks SET commander_card_id = null WHERE id = UNHEX(?)", deckId)
        }
        if deck.OathbreakerCardId == cardId {
            db.Exec("UPDATE Decks SET oathbreaker_card_id = null WHERE id = UNHEX(?)", deckId)
        }

        sideboardCount := models.GetSideboardCount(db, deckId)
        
        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + helpers.EscapeString(card.Name) + " is now in the sideboard\", \"deckUpdated\": \"" + deckId + "\", \"sideboardUpdated\": " + strconv.Itoa(sideboardCount) + "}")

        return c.SendStatus(200)
    })

    app.Delete("/decks/:id/sideboard/:cardId", func(c *fiber.Ctx) error {
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

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Deck_Cards SET sideboard = 0 WHERE card_id = UNHEX(?) AND deck_id = UNHEX(?)", card.Id, deck.Id)

        sideboardCount := models.GetSideboardCount(db, deckId)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"" + card.Name + " added to deck\", \"deckUpdated\": \"" + deckId + "\", \"addedCard\": \"\", \"sideboardUpdated\": " + strconv.Itoa(sideboardCount) + "}")

        return c.SendStatus(200)
    })

    app.Get("/partials/deck-manager/sideboard-card-grid/:id", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        search := c.Query("search")
        sort := c.Query("sort")
        filter := c.Query("filter")
        rarity := c.Query("rarity")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        cards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity)

        for i := range(cards) {
            if cards[i].CardId == deck.CommanderCardId {
                cards[i].IsCommander = true
            }
            if cards[i].CardId == deck.OathbreakerCardId {
                cards[i].IsOathbreaker = true
            }
            if cards[i].Print != 0 {
                printDate := strconv.Itoa(cards[i].Print)
                cards[i].Front = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-front.png"
                if cards[i].Back != "" {
                    cards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(cards[i].CardId) + "-" + printDate +  "-back.png"
                }
            }
        }

        url := "/decks/" + deckId + "?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&filter=" + url.QueryEscape(filter)
        c.Response().Header.Set("Hx-Replace-Url", url)

        return c.Render("partials/deck-manager/sideboard-card-grid", fiber.Map{
            "Cards": cards,
        })
    })

    app.Get("/partials/deck-manager/:deckId/prints", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        cardId := c.Query("cardId")

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

        prints := models.GetPrints(db, cardId)

        for i := range prints {
            prints[i].DeckId = deck.Id
            prints[i].Front = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(prints[i].CardId) + "-" + strconv.Itoa(prints[i].Print) +  "-front.png"
            if prints[i].Back != "" {
                prints[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(prints[i].CardId) + "-" + strconv.Itoa(prints[i].Print) +  "-back.png"
            }
        }

        return c.Render("partials/deck-manager/card-prints", fiber.Map{
            "Prints": prints,
            "DeckId": deck.Id,
            "CardId": card.Id,
        })
    })

    app.Delete("/decks/:id/prints/:cardId", func(c *fiber.Ctx) error {
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

        card := models.GetCard(db, cardId)
        if card.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Deck_Cards SET print = null WHERE card_id = UNHEX(?) AND deck_id = UNHEX(?)", card.Id, deck.Id)

        return c.Render("partials/deck-manager/card-image", fiber.Map{
            "Front": card.Front,
            "Back": card.Back,
        })
    })

    app.Patch("/decks/:id/prints/:cardId/:print", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        cardId := c.Params("cardId")
        printId := c.Params("print")

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

        db.Exec("UPDATE Deck_Cards SET print = ? WHERE card_id = UNHEX(?) AND deck_id = UNHEX(?)", printId, card.Id, deck.Id)

        front := "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(card.Id) + "-" + printId +  "-front.png"
        back := ""
        if card.Back != "" {
            back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + strings.ToUpper(card.Id) + "-" + printId +  "-back.png"
        }

        return c.Render("partials/deck-manager/card-image", fiber.Map{
            "Front": front,
            "Back": back,
        })
    })

    app.Get("/partials/deck-manager/sleeves", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Query("id")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        sleeves := models.GetSleeves(db, user.Id)

        for i := range sleeves {
            sleeves[i].DeckId = deckId
            if deck.SleeveId == sleeves[i].Id {
                sleeves[i].Selected = true
            }
        }

        return c.Render("partials/deck-manager/sleeves", fiber.Map{
            "Sleeves": sleeves,
        })
    })

    app.Post("/sleeves/image", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        file, err := c.FormFile("file")
        if err != nil {
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to upload file."}`)
            return c.SendStatus(400)
        }

        deckId := c.FormValue("deckId")

        src, err := file.Open()
        if err != nil {
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to upload file."}`)
            return c.SendStatus(400)
        }
        defer src.Close()

        id := uuid.New().String()
        id = strings.ReplaceAll(id, "-", "")
        mimeType := file.Header.Get("Content-Type")
        switch mimeType {
        case "image/jpeg":
            break
        case "image/png":
            break
        case "image/jpg":
            break
        case "image/webm":
            break
        case "image/gif":
            break
        default:
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to upload file."}`)
            return c.SendStatus(400)
        }

        s3Client := CreateSpacesClient()

        object := s3.PutObjectInput{
            Bucket:      aws.String("divinedrop"),
            Key:         aws.String("users/" + user.Id + "/" + id),
            Body:        src,
            ACL:         aws.String("public-read"),
            ContentType: aws.String(mimeType),
        }
        _, err = s3Client.PutObject(&object)
        if err != nil {
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to upload file."}`)
            return c.SendStatus(500)
        }

        fileUrl := "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/users/" + user.Id + "/" + id

        db := helpers.ConnectDB()

        db.Exec("INSERT INTO Sleeves (id, user_id, image_url) VALUES (UNHEX(?), ?, ?)", id, user.Id, fileUrl)

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        sleeves := models.GetSleeves(db, user.Id)

        for i := range sleeves {
            sleeves[i].DeckId = deckId
            if deck.SleeveId == sleeves[i].Id {
                sleeves[i].Selected = true
            }
        }

        return c.Render("partials/deck-manager/sleeves", fiber.Map{
            "Sleeves": sleeves,
        })
    })

    app.Put("/decks/:deckId/sleeves/:sleeveId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        sleeveId := c.Params("sleeveId")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        sleeve := models.GetSleeve(db, user.Id, sleeveId)
        if sleeve.Id == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Failed to add sleeves to " + deck.Label + "\"}")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET sleeve_id = UNHEX(?) WHERE id = UNHEX(?) AND user_id = ?", sleeve.Id, deck.Id, user.Id)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Added sleeves to " + deck.Label + "\"}")
        return c.SendStatus(200)
    })

    app.Delete("/decks/:deckId/sleeves", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET sleeve_id = null WHERE id = UNHEX(?) AND user_id = ?", deck.Id, user.Id)

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Removed sleeves from " + deck.Label + "\"}")
        return c.SendStatus(200)
    })
}

func CreateSpacesClient() *s3.S3 {
    key := os.Getenv("SPACES_KEY")
    secret := os.Getenv("SPACES_SECRET")

    s3Config := &aws.Config{
        Credentials: credentials.NewStaticCredentials(key, secret, ""),
        Endpoint:    aws.String("https://nyc3.digitaloceanspaces.com"),
        Region:      aws.String("us-east-1"),
        S3ForcePathStyle: aws.Bool(false),
    }

    newSession := session.New(s3Config)
    s3Client := s3.New(newSession)
    return s3Client
}
