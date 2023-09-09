package controllers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
    "net/url"

	"app/helpers"
	"app/models"
)

func DeckEditorControllers(app *fiber.App){
    app.Get("/decks/new", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        uuid := uuid.New().String()
        uuid = strings.ReplaceAll(uuid, "-", "")

        db := helpers.ConnectDB()
        db.Exec("INSERT INTO Decks (id, user_id, label) VALUES (UNHEX(?), ?, 'Untitled')", uuid, user.Id)

        return c.Redirect("/decks/" + uuid + "/edit")
    })

    app.Get("/decks/:id/edit", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")

        search := c.Query("search")
        sort := c.Query("sort")
        rarity := c.Query("rarity")
        legality := c.Query("legality")

        manaStr := c.Query("mana")
        mana := []string{}
        if manaStr != "" {
            mana = strings.Split(manaStr, ",")
        }

        typesStr := c.Query("types")
        types := []string{}
        if typesStr != "" {
            types = strings.Split(typesStr, ",")
        }

        subtypesStr := c.Query("subtypes")
        subtypes := []string{}
        if subtypesStr != "" {
            subtypes = strings.Split(subtypesStr, ",")
        }

        keywordStr := c.Query("keywords")
        keywords := []string{}
        if keywordStr != "" {
            keywords = strings.Split(keywordStr, ",")
        }

        db := helpers.ConnectDB()
        deck := models.GetDeck(db, deckId, user.Id)
        
        if deck.Id == "" {
            return c.Redirect("/")
        }

        decks := models.GetDecks(db, deckId, user.Id)
        cards := models.FilterCards(db, search, sort, mana, types, subtypes, keywords, rarity, legality, 0, 20)
        deckCards := models.GetDeckCards(db, deckId)
        deckMetadata := models.GetDeckMetadata(db, deckId)
        cardTypes := models.GetCardTypes(db)
        cardSubtypes := models.GetCardSubtypes(db)
        cardKeywords := models.GetCardKeywords(db)

        mythicsCount := models.GetMythicsCount(db, deckId)
        uncommonsCount := models.GetUncommonsCount(db, deckId)
        commonsCount := models.GetCommonsCount(db, deckId)
        raresCount := deckMetadata.CardCount - mythicsCount - uncommonsCount - commonsCount

        landCount := models.GetLandCount(db, deckId)

        deckColors := models.GetDeckColors(db, deckId)
        containsW := false
        containsU := false
        containsB := false
        containsR := false
        containsG := false
        for _, color := range deckColors {
            switch color {
                case "W":
                    containsW = true
                case "U":
                    containsU = true
                case "B":
                    containsB = true
                case "R":
                    containsR = true
                case "G":
                    containsG = true
            }
        }

        manaFilterW := false
        manaFilterU := false
        manaFilterB := false
        manaFilterR := false
        manaFilterG := false
        manaFilterC := false
        for _, color := range mana {
            switch color {
                case "W":
                    manaFilterW = true
                case "U":
                    manaFilterU = true
                case "B":
                    manaFilterB = true
                case "R":
                    manaFilterR = true
                case "G":
                    manaFilterG = true
                case "C":
                    manaFilterC = true
            }
        }

        bannerArt := ""
        if deck.CommanderCardId != "" {
            bannerArt = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + deck.CommanderCardId +  "-art.png"
        } else if len(deckCards) > 0 {
            bannerArt = deckCards[len(deckCards) - 1].Art
        }

        return c.Render("pages/deck-builder/index", fiber.Map{
            "Page": "deck-editor",
            "User": user,
            "Deck": deck,
            "Decks": decks,
            "ActiveDeckId": deckId,
            "Cards": cards,
            "SearchPage": 1,
            "DeckCards": deckCards,
            "DeckCardsCount": len(deckCards),
            "DeckMetadata": deckMetadata,
            "ContainsW": containsW,
            "ContainsU": containsU,
            "ContainsB": containsB,
            "ContainsR": containsR,
            "ContainsG": containsG,
            "BannerArtUrl": bannerArt,
            "SearchRaw": search,
            "Sort": sort,
            "ManaFilterW": manaFilterW,
            "ManaFilterU": manaFilterU,
            "ManaFilterB": manaFilterB,
            "ManaFilterR": manaFilterR,
            "ManaFilterG": manaFilterG,
            "ManaFilterC": manaFilterC,
            "CardTypes": cardTypes,
            "TypeChips": types,
            "CardSubtypes": cardSubtypes,
            "SubtypeChips": subtypes,
            "KeywordChips": keywords,
            "CardKeywords": cardKeywords,
            "Rarity": rarity,
            "Legality": legality,
            "MythicsCount": mythicsCount,
            "UncommonsCount": uncommonsCount,
            "CommonsCount": commonsCount,
            "RaresCount": raresCount,
            "LandCount": landCount,
        }, "layouts/main")
    })

    app.Patch("/decks/:id", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        db := helpers.ConnectDB()
        label := c.FormValue("label")
        db.Exec("UPDATE Decks SET label = ? WHERE id = UNHEX(?) AND user_id = ?", label, deckId, user.Id)

        deck := models.Deck{}
        db.Raw("SELECT HEX(id) AS id, label, HEX(commander_card_id) AS commander_card_id, user_id FROM Decks WHERE id = UNHEX(?) AND user_id = ?", deckId, user.Id).Scan(&deck)

        c.Response().Header.Set("HX-Trigger-After-Swap", "{\"deckUpdated\": \"" + deck.Id + "\"}")

        return c.Render("partials/deck-builder/label-input", fiber.Map{
            "Deck": deck,
        })
    })

    app.Get("/partials/deck-builder/rarity-counts", func(c *fiber.Ctx) error {
        db := helpers.ConnectDB()

        deckId := c.Query("active-deck-id")
        mythicsCount := models.GetMythicsCount(db, deckId)
        uncommonsCount := models.GetUncommonsCount(db, deckId)
        commonsCount := models.GetCommonsCount(db, deckId)
        deckMetadata := models.GetDeckMetadata(db, deckId)
        raresCount := deckMetadata.CardCount - mythicsCount - uncommonsCount - commonsCount

        return c.Render("partials/deck-builder/rarity-counts", fiber.Map{
            "MythicsCount": mythicsCount,
            "UncommonsCount": uncommonsCount,
            "CommonsCount": commonsCount,
            "RaresCount": raresCount,
        })
    })

    app.Post("/partials/deck-builder/card-grid", func(c *fiber.Ctx) error {

        form, err := c.MultipartForm()
        if err == nil {
            search := form.Value["search"][0]
            sort := form.Value["sort"][0]
            mana := form.Value["mana[]"]
            types := form.Value["types[]"]
            subtypes := form.Value["subtypes[]"]
            keywords := form.Value["keywords[]"]
            deckId := form.Value["deck-id"][0]
            rarity := form.Value["rarity"][0]
            legality := form.Value["legality"][0]
            page := form.Value["page"][0]
            var pageInt int
            fmt.Sscan(page, &pageInt)
            offset := pageInt * 20

            db := helpers.ConnectDB()
            cards := models.FilterCards(db, search, sort, mana, types, subtypes, keywords, rarity, legality, offset, 20)

            if len(cards) > 0 {
                c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdated")
            }

            c.Response().Header.Set("HX-Replace-Url", "/decks/" + deckId + "/edit?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&mana=" + url.QueryEscape(strings.Join(mana, ",")) + "&types=" + url.QueryEscape(strings.Join(types, ",")) + "&subtypes=" + url.QueryEscape(strings.Join(subtypes, ",")) + "&keywords=" + url.QueryEscape(strings.Join(keywords, ",")) + "&rarity=" + url.QueryEscape(rarity) + "&legality=" + url.QueryEscape(legality))

            return c.Render("partials/deck-builder/card-grid", fiber.Map{
                "Cards": cards,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Post("/partials/deck-builder/card-grid-filters", func(c *fiber.Ctx) error {

        form, err := c.MultipartForm()
        if err == nil {
            search := form.Value["search"][0]
            sort := form.Value["sort"][0]
            deckId := form.Value["deck-id"][0]
            mana := form.Value["mana[]"]
            types := form.Value["types[]"]
            subtypes := form.Value["subtypes[]"]
            keywords := form.Value["keywords[]"]
            rarity := form.Value["rarity"][0]
            legality := form.Value["legality"][0]

            db := helpers.ConnectDB()
            cards := models.FilterCards(db, search, sort, mana, types, subtypes, keywords, rarity, legality, 0, 20)

            c.Response().Header.Set("HX-Replace-Url", "/decks/" + deckId + "/edit?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&mana=" + url.QueryEscape(strings.Join(mana, ",")) + "&types=" + url.QueryEscape(strings.Join(types, ",")) + "&subtypes=" + url.QueryEscape(strings.Join(subtypes, ",")) + "&keywords=" + url.QueryEscape(strings.Join(keywords, ",")) + "&rarity=" + url.QueryEscape(rarity) + "&legality=" + url.QueryEscape(legality))
            c.Response().Header.Set("HX-Trigger", "cardGridReset")

            return c.Render("partials/deck-builder/card-grid", fiber.Map{
                "Cards": cards,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Get("/partials/deck-builder/card-grid-loader" , func(c *fiber.Ctx) error {
        page := c.QueryInt("page")
        page = page + 1
        return c.Render("partials/deck-builder/card-grid-loader", fiber.Map{
            "ActiveDeckId": c.Query("active-deck-id"),
            "SearchPage": page,
        })
    })

    app.Get("/partials/deck-builder/card-grid-settings" , func(c *fiber.Ctx) error {
        deckId := c.Query("active-deck-id")
        return c.Render("partials/deck-builder/card-grid-settings", fiber.Map{
            "SearchPage": 1,
            "ActiveDeckId": deckId,
        })
    })

    app.Put("/partials/deck-tray/card/:id", func(c *fiber.Ctx) error {
        _, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.FormValue("active-deck-id", "")
        cardId := c.Params("id")

        db := helpers.ConnectDB()

        deckCard := models.DeckCard{}
        db.Raw("SELECT HEX(deck_id) AS deck_id, HEX(card_id) AS card_id, HEX(id) AS id, qty FROM Deck_Cards WHERE deck_id = UNHEX(?) AND card_id = UNHEX(?)", activeDeckId, cardId).Scan(&deckCard)
        if deckCard.Id != "" {
            if (deckCard.Qty < 255) {
                db.Exec("UPDATE Deck_Cards SET qty = ? WHERE id = UNHEX(?)", deckCard.Qty + 1, deckCard.Id)
            }
        } else {
            uuid := uuid.New().String()
            uuid = strings.ReplaceAll(uuid, "-", "")
            db.Exec("INSERT INTO Deck_Cards (id, deck_id, card_id) VALUES (UNHEX(?), UNHEX(?), UNHEX(?))", uuid, activeDeckId, cardId)
        }

        c.Response().Header.Set("HX-Trigger-After-Swap", "{\"deckUpdated\": \"" + activeDeckId + "\"}")

        deckCards := models.GetDeckCards(db, activeDeckId)
        return c.Render("partials/deck-builder/deck-tray", fiber.Map{
            "DeckCards": deckCards,
            "DeckCardsCount": len(deckCards),
        })
    })

    app.Delete("/partials/deck-tray/card/:id", func(c *fiber.Ctx) error {
        _, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.FormValue("active-deck-id", "")
        cardId := c.Params("id")

        db := helpers.ConnectDB()

        db.Exec("DELETE FROM Deck_Cards WHERE deck_id = UNHEX(?) AND card_id = UNHEX(?)", activeDeckId, cardId)
        deckCards := models.GetDeckCards(db, activeDeckId)

        c.Response().Header.Set("HX-Trigger-After-Swap", "{\"deckUpdated\": \"" + activeDeckId + "\"}")

        return c.Render("partials/deck-builder/deck-tray", fiber.Map{
            "DeckCards": deckCards,
            "DeckCardsCount": len(deckCards),
        })
    })

    app.Get("/partials/deck-builder/card-count/:id", func(c *fiber.Ctx) error {
        db := helpers.ConnectDB()

        deckId := c.Params("id")

        deckMetadata := models.DeckMetadata{}
        db.Raw("SELECT HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?)", deckId).Scan(&deckMetadata)

        return c.Render("partials/deck-builder/card-count", fiber.Map{
            "DeckMetadata": deckMetadata,
        })
    })

    app.Get("/partials/deck-builder/land-count/:id", func(c *fiber.Ctx) error {
        db := helpers.ConnectDB()

        deckId := c.Params("id")
        landCount := models.GetLandCount(db, deckId)

        return c.Render("partials/deck-builder/land-count", fiber.Map{
            "LandCount": landCount,
            "ActiveDeckId": deckId,
        })
    })

    app.Get("/partials/deck-builder/mana-types/:deckId", func(c *fiber.Ctx) error {
        _, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        deckId := c.Params("deckId")

        db := helpers.ConnectDB()
        deckColors := models.GetDeckColors(db, deckId)
        containsW := false
        containsU := false
        containsB := false
        containsR := false
        containsG := false
        for _, color := range deckColors {
            switch color {
                case "W":
                    containsW = true
                case "U":
                    containsU = true
                case "B":
                    containsB = true
                case "R":
                    containsR = true
                case "G":
                    containsG = true
            }
        }

        deckMetadata := models.DeckMetadata{}
        db.Raw("SELECT HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) GROUP BY D.id, D.user_id", deckId).Scan(&deckMetadata)

        return c.Render("partials/deck-builder/mana-types", fiber.Map{
            "ContainsW": containsW,
            "ContainsU": containsU,
            "ContainsB": containsB,
            "ContainsR": containsR,
            "ContainsG": containsG,
            "DeckMetadata": deckMetadata,
        })
    })

    app.Get("/partials/deck-builder/banner-art", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        deckId := c.Query("active-deck-id")

        db := helpers.ConnectDB()
        deck := models.GetDeck(db, deckId, user.Id)

        bannerArt := ""
        if deck.CommanderCardId != "" {
            bannerArt = "https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards/" + deck.CommanderCardId +  "-art.png"
        } else {
            deckCards := models.GetDeckCards(db, deckId)
            if len(deckCards) > 0 {
                bannerArt = deckCards[len(deckCards) - 1].Art
            }
        }

        return c.Render("partials/deck-builder/banner-art", fiber.Map{
            "BannerArtUrl": bannerArt,
        })
    })

    app.Get("/partials/deck-builder/card-types", func(c *fiber.Ctx) error {
        typeStr := c.Query("types")

        db := helpers.ConnectDB()
        cardTypes := models.SearchCardTypes(db, typeStr)

        return c.Render("partials/deck-builder/card-types", fiber.Map{
            "CardTypes": cardTypes,
        })
    })

    app.Post("/partials/deck-builder/card-type-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            types := form.Value["types[]"]
            if len(form.Value["type"]) > 0 {
                newType := form.Value["type"][0]
                types = append(types, newType)
            }
            types = helpers.RemoveDuplicateStrings(types)

            c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdate")

            return c.Render("partials/deck-builder/card-type-chips", fiber.Map{
                "TypeChips": types,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Delete("/partials/deck-builder/card-type-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            types := form.Value["types[]"]
            typeToDelete := form.Value["type"][0]
            newTypes := []string{}
            for _, t := range types {
                if t != typeToDelete {
                    newTypes = append(newTypes, t)
                }
            }

            c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdate")

            return c.Render("partials/deck-builder/card-type-chips", fiber.Map{
                "TypeChips": newTypes,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Get("/partials/deck-builder/card-subtypes", func(c *fiber.Ctx) error {
        subtypeStr := c.Query("subtypes")

        db := helpers.ConnectDB()
        subtypes := models.SearchCardSubtypes(db, subtypeStr)

        return c.Render("partials/deck-builder/card-subtypes", fiber.Map{
            "CardSubtypes": subtypes,
        })
    })

    app.Post("/partials/deck-builder/card-subtype-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            subtypes := form.Value["subtypes[]"]
            if len(form.Value["subtype"]) > 0 {
                newSubtype := form.Value["subtype"][0]
                subtypes = append(subtypes, newSubtype)
            }
            subtypes = helpers.RemoveDuplicateStrings(subtypes)

            c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdate")

            return c.Render("partials/deck-builder/card-subtype-chips", fiber.Map{
                "SubtypeChips": subtypes,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Delete("/partials/deck-builder/card-subtype-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            subtypes := form.Value["subtypes[]"]
            subtypeToDelete := form.Value["subtype"][0]
            newSubtypes := []string{}
            for _, t := range subtypes {
                if t != subtypeToDelete {
                    newSubtypes = append(newSubtypes, t)
                }
            }

            c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdate")

            return c.Render("partials/deck-builder/card-subtype-chips", fiber.Map{
                "SubtypeChips": newSubtypes,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Get("/partials/deck-builder/card-keywords", func(c *fiber.Ctx) error {
        keywordStr := c.Query("keywords")

        db := helpers.ConnectDB()
        keywords := models.SearchCardKeywords(db, keywordStr)

        return c.Render("partials/deck-builder/card-keywords", fiber.Map{
            "CardKeywords": keywords,
        })
    })

    app.Post("/partials/deck-builder/card-keyword-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            keywords := form.Value["keywords[]"]
            if len(form.Value["keyword"]) > 0 {
                newKeyword := form.Value["keyword"][0]
                keywords = append(keywords, newKeyword)
            }
            keywords = helpers.RemoveDuplicateStrings(keywords)

            c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdate")

            return c.Render("partials/deck-builder/card-keyword-chips", fiber.Map{
                "KeywordChips": keywords,
            })
        } else {
            return c.Send(nil)
        }
    })

    app.Delete("/partials/deck-builder/card-keyword-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            keywords := form.Value["keywords[]"]
            keywordToDelete := form.Value["keyword"][0]
            newKeywords := []string{}
            for _, t := range keywords {
                if t != keywordToDelete {
                    newKeywords = append(newKeywords, t)
                }
            }

            c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdate")

            return c.Render("partials/deck-builder/card-keyword-chips", fiber.Map{
                "KeywordChips": newKeywords,
            })
        } else {
            return c.Send(nil)
        }
    })
}
