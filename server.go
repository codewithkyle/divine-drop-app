package main

import (
	"app/models"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectDB() *gorm.DB {
    db, err := gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("failed to connect to PlanetScale: %v", err)
	}
    return db;
}

func getUser(sessionId string) (models.User, error) {
    if sessionId == "" {
        return models.User{}, errors.New("Session not found");
    }
    db := connectDB()
    var session models.Session
    db.Raw("SELECT * FROM Sessions WHERE session_id = UNHEX(?) AND expires > ?", sessionId, time.Now()).Scan(&session)
    if session.Id == "" {
        return models.User{}, errors.New("Session not found");
    }
    return models.BlobToUser(session.Data)
}

func main() {
    // Load environment variables from file.
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

    client, _ := clerk.NewClient(os.Getenv("CLERK_API_KEY"))

    // Create a new engine
    engine := html.New("./views", ".html")

    // Pass the engine to the Views
    app := fiber.New(fiber.Config{
        Views: engine,
    })

    app.Static("/css", "./public/css")
    app.Static("/js", "./public/js")
    app.Static("/fonts", "./public/fonts")

    app.Get("/", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, _ := getUser(sessionId)

        search := c.Query("search")
        searchQuery := "%" + strings.Trim(search, " ") + "%"

        db := connectDB()
        cards := models.SearchCardsByName(db, searchQuery, 0, 20)

        var decks []models.Deck
        if user.Id != "" {
            db.Raw("SELECT HEX(D.id) AS id, HEX(D.commander_card_id) AS commander_card_id, D.label, D.user_id, (SELECT COUNT(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE user_id = ?", user.Id).Scan(&decks)
        }

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
        searchQuery := "%" + strings.Trim(search, " ") + "%"

        db := connectDB()
        cards := models.SearchCardsByName(db, searchQuery, 0, 20)

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

        searchQuery := "%" + strings.Trim(search, " ") + "%"
        var offset = page * 20

        db := connectDB()
        cards := models.SearchCardsByName(db, searchQuery, offset, 20)

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

    app.Get("/decks/new", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "") 
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        uuid := uuid.New().String()
        uuid = strings.ReplaceAll(uuid, "-", "")

        db := connectDB()
        db.Exec("INSERT INTO Decks (id, user_id, label) VALUES (UNHEX(?), ?, 'Untitled')", uuid, user.Id)

        return c.Redirect("/decks/" + uuid + "/edit");
    })
    app.Get("/decks/:id/edit", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")

        search := c.Query("search")
        searchQuery := "%" + strings.Trim(search, " ") + "%"
        sort := c.Query("sort")

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

        db := connectDB()
        deck := models.GetDeck(db, deckId, user.Id)
        
        if deck.Id == "" {
            return c.Redirect("/")
        }

        decks := models.GetDecks(db, deckId, user.Id)
        cards := models.FilterCards(db, searchQuery, sort, mana, types, subtypes, keywords, 0, 20)
        deckCards := models.GetDeckCards(db, deckId)
        deckMetadata := models.GetDeckMetadata(db, deckId)
        cardTypes := models.GetCardTypes(db)
        cardSubtypes := models.GetCardSubtypes(db)
        cardKeywords := models.GetCardKeywords(db)

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
        }, "layouts/main")
    })
    app.Patch("/decks/:id", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        db := connectDB()
        label := c.FormValue("label")
        db.Exec("UPDATE Decks SET label = ? WHERE id = UNHEX(?) AND user_id = ?", label, deckId, user.Id)

        deck := models.Deck{}
        db.Raw("SELECT HEX(id) AS id, label, HEX(commander_card_id) AS commander_card_id, user_id FROM Decks WHERE id = UNHEX(?) AND user_id = ?", deckId, user.Id).Scan(&deck)

        c.Response().Header.Set("HX-Trigger-After-Swap", "{\"deckUpdated\": \"" + deck.Id + "\"}")

        return c.Render("partials/deck-builder/label-input", fiber.Map{
            "Deck": deck,
        })
    })
    app.Post("/partials/deck-builder/card-grid", func(c *fiber.Ctx) error {

        form, err := c.MultipartForm()
        if err == nil {
            search := form.Value["search"][0]
            searchQuery := "%" + strings.Trim(search, " ") + "%"
            sort := form.Value["sort"][0]
            mana := form.Value["mana[]"]
            types := form.Value["types[]"]
            subtypes := form.Value["subtypes[]"]
            keywords := form.Value["keywords[]"]
            deckId := form.Value["deck-id"][0]
            page := form.Value["page"][0]
            var pageInt int
            fmt.Sscan(page, &pageInt)
            offset := pageInt * 20

            db := connectDB()
            cards := models.FilterCards(db, searchQuery, sort, mana, types, subtypes, keywords, offset, 20)

            if len(cards) > 0 {
                c.Response().Header.Set("HX-Trigger-After-Swap", "cardGridUpdated")
            }

            c.Response().Header.Set("HX-Replace-Url", "/decks/" + deckId + "/edit?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&mana=" + url.QueryEscape(strings.Join(mana, ",")) + "&types=" + url.QueryEscape(strings.Join(types, ",")) + "&subtypes=" + url.QueryEscape(strings.Join(subtypes, ",")) + "&keywords=" + url.QueryEscape(strings.Join(keywords, ",")))

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
            searchQuery := "%" + strings.Trim(search, " ") + "%"
            sort := form.Value["sort"][0]
            deckId := form.Value["deck-id"][0]
            mana := form.Value["mana[]"]
            types := form.Value["types[]"]
            subtypes := form.Value["subtypes[]"]
            keywords := form.Value["keywords[]"]

            db := connectDB()
            cards := models.FilterCards(db, searchQuery, sort, mana, types, subtypes, keywords, 0, 20)

            c.Response().Header.Set("HX-Replace-Url", "/decks/" + deckId + "/edit?search=" + url.QueryEscape(search) + "&sort=" + url.QueryEscape(sort) + "&mana=" + url.QueryEscape(strings.Join(mana, ",")) + "&types=" + url.QueryEscape(strings.Join(types, ",")) + "&subtypes=" + url.QueryEscape(strings.Join(subtypes, ",")) + "&keywords=" + url.QueryEscape(strings.Join(keywords, ",")))
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
        sessionId := c.Cookies("session_id", "")
        _, err := getUser(sessionId)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.FormValue("active-deck-id", "")
        cardId := c.Params("id")

        db := connectDB()
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
        })
    })
    app.Delete("/partials/deck-tray/card/:id", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        _, err := getUser(sessionId)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.FormValue("active-deck-id", "")
        cardId := c.Params("id")

        db := connectDB()
        db.Exec("DELETE FROM Deck_Cards WHERE deck_id = UNHEX(?) AND card_id = UNHEX(?)", activeDeckId, cardId)
        deckCards := models.GetDeckCards(db, activeDeckId)

        c.Response().Header.Set("HX-Trigger-After-Swap", "{\"deckUpdated\": \"" + activeDeckId + "\"}")

        return c.Render("partials/deck-builder/deck-tray", fiber.Map{
            "DeckCards": deckCards,
        })
    })
    app.Get("/partials/deck-builder/card-count/:id", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        _, err := getUser(sessionId)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        deckId := c.Params("id")

        db := connectDB()
        deckMetadata := models.DeckMetadata{}
        db.Raw("SELECT HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?)", deckId).Scan(&deckMetadata)

        return c.Render("partials/deck-builder/card-count", fiber.Map{
            "DeckMetadata": deckMetadata,
        })
    })
    app.Get("/partials/deck-builder/mana-types/:deckId", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        _, err := getUser(sessionId)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        deckId := c.Params("deckId")

        db := connectDB()
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
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        deckId := c.Query("active-deck-id")

        db := connectDB()
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
        db := connectDB()
        cardTypes := models.SearchCardTypes(db, typeStr)
        return c.Render("partials/deck-builder/card-types", fiber.Map{
            "CardTypes": cardTypes,
        })
    })
    app.Post("/partials/deck-builder/card-type-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            types := form.Value["types[]"]
            newType := form.Value["type"][0]
            types = append(types, newType)

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
        db := connectDB()
        subtypes := models.SearchCardSubtypes(db, subtypeStr)
        return c.Render("partials/deck-builder/card-subtypes", fiber.Map{
            "CardSubtypes": subtypes,
        })
    })
    app.Post("/partials/deck-builder/card-subtype-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            subtypes := form.Value["subtypes[]"]
            newSubtype := form.Value["subtype"][0]
            subtypes = append(subtypes, newSubtype)

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
        db := connectDB()
        keywords := models.SearchCardKeywords(db, keywordStr)
        return c.Render("partials/deck-builder/card-keywords", fiber.Map{
            "CardKeywords": keywords,
        })
    })
    app.Post("/partials/deck-builder/card-keyword-chips", func(c *fiber.Ctx) error {
        form, err := c.MultipartForm()
        if err == nil {
            keywords := form.Value["keywords[]"]
            newKeyword := form.Value["keyword"][0]
            keywords = append(keywords, newKeyword)

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

    app.Get("/partials/nav/decks-opened", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }
        activeDeckId := c.Query("active-deck-id", "")
        db := connectDB()
        decks := models.GetDecks(db, activeDeckId, user.Id)

        return c.Render("partials/nav/decks-opened", fiber.Map{
            "Decks": decks,
            "ActiveDeckId": activeDeckId,
        })
    })
    app.Get("/partials/nav/decks-closed", func(c *fiber.Ctx) error {
        activeDeckId := c.Query("active-deck-id")
        return c.Render("partials/nav/decks-closed", fiber.Map{
            "ActiveDeckId": activeDeckId,
        })
    })
    app.Get("/partials/nav/decks/:id", func(c *fiber.Ctx) error {
        sessionId := c.Cookies("session_id", "")
        user, err := getUser(sessionId)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        deckId := c.Params("id")
        db := connectDB()
        deck := models.GetDeck(db, deckId, user.Id)

        deckMetadata := models.DeckMetadata{}
        db.Raw("SELECT HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?)", deck.Id).Scan(&deckMetadata)

        return c.Render("partials/nav/deck-link", fiber.Map{
            "Id": deck.Id,
            "Label": deck.Label,
            "CardCount": deckMetadata.CardCount,
            "Active": deck.Active,
        })
    })

    app.Get("/register", func(c *fiber.Ctx) error {
        return c.Render("pages/register/index", fiber.Map{})
    })
    app.Get("/sign-in", func(c *fiber.Ctx) error {
        return c.Render("pages/sign-in/index", fiber.Map{})
    })
    app.Get("/authorize", func(c *fiber.Ctx) error {
        token := c.Cookies("__session", "")
        if token == "" {
            return c.Redirect("/sign-in")
        }
        sessClaims, err := client.VerifyToken(token)
        if err != nil {
            return c.Redirect("/sign-in")
        }
        user, err := client.Users().Read(sessClaims.Claims.Subject)
		if err != nil {
            return c.Redirect("/sign-in")
		}

        email := ""
        if (len(user.EmailAddresses) > 0) {
            email = user.EmailAddresses[0].EmailAddress
        }

        username := ""
        if (user.Username != nil) {
            username = *user.Username
        } else {
            username = strings.Trim(user.ID, "user_")
        }

        customUser := models.User{
            Id: user.ID,
            Username: username,
            Email: email,
            Avatar: user.ProfileImageURL,
        }
        sessionId := uuid.New().String()
        sessionId = strings.ReplaceAll(sessionId, "-", "")
        expires := time.Now().Add(168 * time.Hour)
        blob, _ := models.UserToBlob(customUser)

        // TODO: insert into DB
        db := connectDB()
        db.Exec("INSERT INTO Sessions (session_id, user_id, data, expires) VALUES (UNHEX(?), ?, ?, ?)", sessionId, customUser.Id, blob, expires)

        c.Cookie(&fiber.Cookie{
            Name: "session_id",
            Value: sessionId,
            Expires: expires,
            Secure: true,
            HTTPOnly: true,
            SameSite: "Strict",
        })

        return c.Redirect("/")
    })

    app.Listen(":3000")
}
