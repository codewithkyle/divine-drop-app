package controllers

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"app/helpers"
	"app/models"
)

type PublicMetadata struct {
    Name string
    Cost string
}

func DeckManagerControllers(app *fiber.App){
    app.Get("/decks/:id", func(c *fiber.Ctx) error {
        user, _ := helpers.GetUserFromSession(c)

        deckId := c.Params("id")

        db := helpers.ConnectDB()

        deck := models.GetDeckByID(db, deckId)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        isGuest := true
        if user.Id != "" && deck.UserId == user.Id {
            isGuest = false
        }

        search := c.Query("search")
        sort := c.Query("sort")
        filter := c.Query("filter")
        rarity := c.Query("rarity")
        color := c.Query("color")

        decks := []models.Deck{}
        deckCards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity, color)
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

            if deck.Gamemode != "" {
                switch deck.Gamemode {
                    case "standard":
                        deckCards[i].IsLegal = deckCards[i].LegalStandard
                    case "future":
                        deckCards[i].IsLegal = deckCards[i].LegalFuture
                    case "historic":
                        deckCards[i].IsLegal = deckCards[i].LegalHistoric
                    case "gladiator":
                        deckCards[i].IsLegal = deckCards[i].LegalGladiator
                    case "pioneer":
                        deckCards[i].IsLegal = deckCards[i].LegalPioneer
                    case "explorer":
                        deckCards[i].IsLegal = deckCards[i].LegalExplorer
                    case "modern":
                        deckCards[i].IsLegal = deckCards[i].LegalModern
                    case "legacy":
                        deckCards[i].IsLegal = deckCards[i].LegalLegacy
                    case "pauper":
                        deckCards[i].IsLegal = deckCards[i].LegalPauper
                    case "vintage":
                        deckCards[i].IsLegal = deckCards[i].LegalVintage
                    case "commander":
                        deckCards[i].IsLegal = deckCards[i].LegalCommander
                    case "oathbreaker":
                        deckCards[i].IsLegal = deckCards[i].LegalOathbreaker
                    case "brawl":
                        deckCards[i].IsLegal = deckCards[i].LegalBrawl
                    case "historicbrawl":
                        deckCards[i].IsLegal = deckCards[i].LegalHistoricBrawl
                    case "alchemy":
                        deckCards[i].IsLegal = deckCards[i].LegalAlchemy
                    case "paupercommander":
                        deckCards[i].IsLegal = deckCards[i].LegalPauperCommander
                    case "duel":
                        deckCards[i].IsLegal = deckCards[i].LegalDuel
                    case "oldschool":
                        deckCards[i].IsLegal = deckCards[i].LegalOldSchool
                    case "premodern":
                        deckCards[i].IsLegal = deckCards[i].LegalPremodern
                    case "predh":
                        deckCards[i].IsLegal = deckCards[i].LegalPredh
                }
            } else {
                deckCards[i].IsLegal = true
            }
        }

        deckGroups := []models.DeckGroup{}
        if user.Id != "" {
            deckGroups = models.GetDeckGroups(db, user.Id)
            decks = models.GetDecks(db, deckId, user.Id)
        }

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

        cost := models.GetDeckCost(db, deck.Id);

        overBudget := false
        if deckMetadata.Budget > 0 && int(cost * 100) > deckMetadata.Budget {
            overBudget = true
        }

        for i := range deckCards {
            deckCards[i].FmtPrice = fmt.Sprintf("%.2f", float32(deckCards[i].Price * int(deckCards[i].Qty)) / 100)
            deckCards[i].IsGuest = isGuest
        }

        return c.Render("pages/deck-manager/index", fiber.Map{
            "Legality": deck.Gamemode,
            "IsGuest": isGuest,
            "IsOverBudget": overBudget,
            "Budget": fmt.Sprintf("%.2f", float32(deckMetadata.Budget) / 100),
            "DeckPrice": fmt.Sprintf("%.2f", cost),
            "Page": "deck-manager",
            "User": user,
            "Deck": deck,
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "Cards": deckCards,
            "ActiveDeckId": deckId,
            "DeckCardsCount": len(deckCards),
            "DeckMetadata": deckMetadata,
            "ContainsW": containsW,
            "ContainsU": containsU,
            "ContainsB": containsB,
            "ContainsR": containsR,
            "ContainsG": containsG,
            "EncodedBannerArtUrl": url.QueryEscape(bannerArt),
            "BannerArtUrl": bannerArt,
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
        user, _ := helpers.GetUserFromSession(c)

        deckId := c.Params("id")
        search := c.Query("search")
        sort := c.Query("sort")
        filter := c.Query("filter")
        rarity := c.Query("rarity")
        color := c.Query("color")

        db := helpers.ConnectDB()

        deck := models.GetDeckByID(db, deckId)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        isGuest := true
        if user.Id != "" && deck.UserId == user.Id {
            isGuest = false
        }

        cards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity, color)

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
            cards[i].IsGuest = isGuest

            if deck.Gamemode != "" {
                switch deck.Gamemode {
                    case "standard":
                        cards[i].IsLegal = cards[i].LegalStandard
                    case "future":
                        cards[i].IsLegal = cards[i].LegalFuture
                    case "historic":
                        cards[i].IsLegal = cards[i].LegalHistoric
                    case "gladiator":
                        cards[i].IsLegal = cards[i].LegalGladiator
                    case "pioneer":
                        cards[i].IsLegal = cards[i].LegalPioneer
                    case "explorer":
                        cards[i].IsLegal = cards[i].LegalExplorer
                    case "modern":
                        cards[i].IsLegal = cards[i].LegalModern
                    case "legacy":
                        cards[i].IsLegal = cards[i].LegalLegacy
                    case "pauper":
                        cards[i].IsLegal = cards[i].LegalPauper
                    case "vintage":
                        cards[i].IsLegal = cards[i].LegalVintage
                    case "commander":
                        cards[i].IsLegal = cards[i].LegalCommander
                    case "oathbreaker":
                        cards[i].IsLegal = cards[i].LegalOathbreaker
                    case "brawl":
                        cards[i].IsLegal = cards[i].LegalBrawl
                    case "historicbrawl":
                        cards[i].IsLegal = cards[i].LegalHistoricBrawl
                    case "alchemy":
                        cards[i].IsLegal = cards[i].LegalAlchemy
                    case "paupercommander":
                        cards[i].IsLegal = cards[i].LegalPauperCommander
                    case "duel":
                        cards[i].IsLegal = cards[i].LegalDuel
                    case "oldschool":
                        cards[i].IsLegal = cards[i].LegalOldSchool
                    case "premodern":
                        cards[i].IsLegal = cards[i].LegalPremodern
                    case "predh":
                        cards[i].IsLegal = cards[i].LegalPredh
                }
            } else {
                cards[i].IsLegal = true
            }
        }

        url := "/decks/" + deckId + "?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&filter=" + url.QueryEscape(filter) + "&rarity=" + url.QueryEscape(rarity) + "&color=" + url.QueryEscape(color)
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
                if deck.SleeveImage != "" {
                    cards[i].Back = deck.SleeveImage
                } else {
                    cards[i].Back = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/back.png"
                }
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

    app.Get("/api/v1/decks/:id/metadata", func(c *fiber.Ctx) error {
        deckId := c.Params("id")
        if deckId == "" {
            return c.SendStatus(400)
        }
        db := helpers.ConnectDB()
        deck := models.GetDeckMetadata(db, deckId)
        cost := models.GetDeckCost(db, deck.Id)

        metadata := PublicMetadata{}
        metadata.Name = deck.Label
        metadata.Cost = fmt.Sprintf("%.2f", cost)

        return c.JSON(metadata)
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
        user, _ := helpers.GetUserFromSession(c)

        deckId := c.Params("id")
        search := c.Query("search")
        sort := c.Query("sort")
        filter := c.Query("filter")
        rarity := c.Query("rarity")
        color := c.Query("color")

        db := helpers.ConnectDB()

        deck := models.GetDeckByID(db, deckId)
        if deck.Id == "" {
            return c.Redirect("/")
        }

        isGuest := true
        if user.Id != "" && deck.UserId == user.Id {
            isGuest = false
        }

        cards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity, color)

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
            cards[i].IsGuest = isGuest
        }

        url := "/decks/" + deckId + "?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&filter=" + url.QueryEscape(filter) + "&rarity=" + url.QueryEscape(rarity) + "&color=" + url.QueryEscape(color)
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
            "DeckId": deck.Id,
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
        if deckId == "" {
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to upload file."}`)
            return c.SendStatus(400)
        }

        src, err := file.Open()
        if err != nil {
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to upload file."}`)
            return c.SendStatus(400)
        }
        defer src.Close()

        id := uuid.New().String()
        id = strings.ReplaceAll(id, "-", "")
        mimeType := file.Header.Get("Content-Type")
        isVideo := false
        switch mimeType {
        case "image/jpeg":
            break
        case "image/png":
            break
        case "image/jpg":
            break
        case "video/webm":
            isVideo = true
            break
        case "video/mp4":
            isVideo = true
            break
        case "video/mov":
            isVideo = true
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

        db.Exec("INSERT INTO Sleeves (id, user_id, image_url, is_video) VALUES (UNHEX(?), ?, ?, ?)", id, user.Id, fileUrl, isVideo)

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

    app.Delete("/sleeves/image/:id", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        sleeveId := c.Params("id")
        deckId := c.Query("deckId")

        db := helpers.ConnectDB()

        sleeve := models.GetSleeve(db, user.Id, sleeveId)
        if sleeve.Id == "" {
            sleeves := models.GetSleeves(db, user.Id)
            return c.Render("partials/deck-manager/sleeves", fiber.Map{
                "Sleeves": sleeves,
            })
        }

        s3Client := CreateSpacesClient()

        object := s3.DeleteObjectInput{
            Bucket:      aws.String("divinedrop"),
            Key:         aws.String("users/" + user.Id + "/" + strings.ToLower(sleeve.Id)),
        }
        _, err = s3Client.DeleteObject(&object)
        if err != nil {
            c.Response().Header.Set("HX-Trigger", `{"flash:toast": "Failed to delete file."}`)
            return c.SendStatus(500)
        }

        db.Exec("UPDATE Decks SET sleeve_id = null WHERE sleeve_id = UNHEX(?) AND user_id = ?", sleeve.Id, user.Id)

        db.Exec("DELETE FROM Sleeves WHERE id = UNHEX(?) AND user_id = ?", sleeve.Id, user.Id)

        sleeves := models.GetSleeves(db, user.Id)

        for i := range sleeves {
            sleeves[i].DeckId = deckId
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

    app.Post("/decks/:deckId/budget", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        budget, err := strconv.ParseFloat(strings.Trim(c.Get("HX-Prompt", "0"), "$"), 32)
        if err != nil {
            budget = 0;
        }
        budgetInt := int(budget * 100)

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        if budget > 0 {
            db.Exec("UPDATE Decks SET budget = ? WHERE id = UNHEX(?) AND user_id = ?", budgetInt, deck.Id, user.Id)
        } else {
            db.Exec("UPDATE Decks SET budget = null WHERE id = UNHEX(?) AND user_id = ?", deck.Id, user.Id)
        }

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Deck budget updated to $" + fmt.Sprintf("%.2f", budget) + "\", \"deckUpdated\": \"" + deckId + "\"}")
        return c.Render("partials/deck-builder/budget", fiber.Map{
            "Budget": fmt.Sprintf("%.2f", budget),
        })
    })

    app.Post("/decks/:deckId/gamemode", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("deckId")
        gamemode := c.FormValue("gamemode", "")

        db := helpers.ConnectDB()

        deck := models.GetDeck(db, deckId, user.Id)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Redirect", "/")
            return c.SendStatus(404)
        }

        if gamemode == "" {
            db.Exec("UPDATE Decks SET gamemode = null WHERE id = UNHEX(?) AND user_id = ?", deck.Id, user.Id)
        } else {
            db.Exec("UPDATE Decks SET gamemode = ? WHERE id = UNHEX(?) AND user_id = ?", gamemode, deck.Id, user.Id)
        }

        search := c.FormValue("search")
        sort := c.FormValue("sort")
        filter := c.FormValue("filter")
        rarity := c.FormValue("rarity")
        color := c.FormValue("color")

        cards := models.SearchDeckCards(db, deckId, search, sort, filter, rarity, color)

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

            if gamemode != "" {
                switch gamemode {
                    case "standard":
                        cards[i].IsLegal = cards[i].LegalStandard
                    case "pioneer":
                        cards[i].IsLegal = cards[i].LegalPioneer
                    case "modern":
                        cards[i].IsLegal = cards[i].LegalModern
                    case "legacy":
                        cards[i].IsLegal = cards[i].LegalLegacy
                    case "pauper":
                        cards[i].IsLegal = cards[i].LegalPauper
                    case "vintage":
                        cards[i].IsLegal = cards[i].LegalVintage
                    case "commander":
                        cards[i].IsLegal = cards[i].LegalCommander
                    case "oathbreaker":
                        cards[i].IsLegal = cards[i].LegalOathbreaker
                    case "brawl":
                        cards[i].IsLegal = cards[i].LegalBrawl
                    case "historicbrawl":
                        cards[i].IsLegal = cards[i].LegalHistoricBrawl
                    case "alchemy":
                        cards[i].IsLegal = cards[i].LegalAlchemy
                    case "paupercommander":
                        cards[i].IsLegal = cards[i].LegalPauperCommander
                    case "duel":
                        cards[i].IsLegal = cards[i].LegalDuel
                    case "oldschool":
                        cards[i].IsLegal = cards[i].LegalOldSchool
                    case "premodern":
                        cards[i].IsLegal = cards[i].LegalPremodern
                    case "predh":
                        cards[i].IsLegal = cards[i].LegalPredh
                }
            } else {
                cards[i].IsLegal = true
            }
        }

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\":\"Deck gamemode changed to " + gamemode + "\", \"deckUpdated\": \"" + deckId + "\"}")
        return c.Render("partials/deck-manager/card-grid", fiber.Map{
            "Cards": cards,
        })
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
