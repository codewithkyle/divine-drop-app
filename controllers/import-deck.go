package controllers

import (
    "io"
    "regexp"
    "strconv"
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "app/helpers"
    "app/models"
)

// importMaxBytes caps how much of an uploaded file is read. A hundred card deck
// list is a couple of kilobytes; the server accepts much larger bodies for
// sleeve uploads, and none of that headroom is wanted here.
const importMaxBytes = 512 * 1024

// importMaxCards caps how many lines one import may resolve, so a large paste
// cannot turn into an unbounded lookup.
const importMaxCards = 1000

// deckCardQtyMax matches the unsigned tinyint the qty column is stored in.
const deckCardQtyMax = 255

// importCardValue is the hidden input the review form carries per card, written
// as "<card id>:<qty>:<sideboard>".
var importCardValue = regexp.MustCompile(`^([0-9A-Fa-f]{32}):([0-9]{1,3}):([01])$`)

var importCardId = regexp.MustCompile(`^[0-9A-Fa-f]{32}$`)

// importGamemodes are the values the Decks.gamemode column accepts, matching
// the restrictions dropdown on the deck overview.
var importGamemodes = map[string]bool{
    "standard": true, "future": true, "historic": true, "gladiator": true,
    "pioneer": true, "explorer": true, "modern": true, "legacy": true,
    "pauper": true, "vintage": true, "penny": true, "commander": true,
    "oathbreaker": true, "brawl": true, "historicbrawl": true, "alchemy": true,
    "paupercommander": true, "duel": true, "oldschool": true, "premodern": true,
    "predh": true,
}

// commanderGamemodes are the formats played with a card in the command zone.
// They decide two things: whether the review screen offers a commander picker,
// and what an unlabelled trailing block in a Moxfield export means.
var commanderGamemodes = map[string]bool{
    "commander":       true,
    "paupercommander": true,
    "predh":           true,
    "brawl":           true,
    "historicbrawl":   true,
    "duel":            true,
    "oathbreaker":     true,
}

// ImportCard is a resolved card on the review screen.
type ImportCard struct {
    CardId      string
    Name        string
    Front       string
    Qty         int
    InSideboard bool
    IsCommander bool
}

// ImportMiss is a line that parsed but matched no card.
type ImportMiss struct {
    Name string
    Qty  int
}

func ImportDeckControllers(app *fiber.App) {
    app.Get("/decks/import", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            return c.Redirect("/sign-in")
        }

        db := helpers.ConnectDB()
        groupedDecks, ungroupedDecks, deckCount := importNavDecks(db, user.Id)

        return c.Render("pages/deck-import/index", fiber.Map{
            "Page":           "deck-import",
            "User":           user,
            "GroupedDecks":   groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "ActiveDeckId":   "",
            "Gamemode":       "",
            "NavClosed":      c.Cookies("nav_closed", "") == "true" || deckCount == 0,
        }, "layouts/main")
    })

    app.Get("/partials/deck-import/form", func(c *fiber.Ctx) error {
        if _, err := helpers.GetUserFromSession(c); err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.SendStatus(401)
        }

        return c.Render("partials/deck-import/form", fiber.Map{
            "Gamemode": c.Query("gamemode", ""),
        })
    })

    app.Post("/decks/import/preview", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.SendStatus(401)
        }

        gamemode := c.FormValue("gamemode", "")
        if !importGamemodes[gamemode] {
            gamemode = ""
        }

        raw := readImportedList(c)
        if strings.TrimSpace(raw) == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Pick a file or paste a deck list first\"}")
            return c.SendStatus(422)
        }

        parsed := helpers.ParseDeckList(raw, commanderGamemodes[gamemode])
        truncated := false
        if len(parsed.Cards) > importMaxCards {
            parsed.Cards = parsed.Cards[:importMaxCards]
            truncated = true
        }

        db := helpers.ConnectDB()
        found, missing := resolveImportedCards(db, parsed.Cards)

        mainCount := 0
        sideboardCount := 0
        commanders := []string{}
        for _, card := range found {
            if card.InSideboard {
                sideboardCount += card.Qty
            } else {
                mainCount += card.Qty
            }
            if card.IsCommander {
                commanders = append(commanders, card.CardId)
            }
        }

        commanderId := ""
        partnerId := ""
        if len(commanders) > 0 {
            commanderId = commanders[0]
        }
        if len(commanders) > 1 {
            partnerId = commanders[1]
        }

        return c.Render("partials/deck-import/review", fiber.Map{
            "User":            user,
            "Gamemode":        gamemode,
            "Found":           found,
            "TotalEntries":    len(found) + len(missing),
            "Missing":         missing,
            "MainCount":       mainCount,
            "SideboardCount":  sideboardCount,
            "Skipped":         parsed.Skipped,
            "BadLines":        parsed.BadLines,
            "Truncated":       truncated,
            "ShowCommander":   commanderGamemodes[gamemode],
            "ShowOathbreaker": gamemode == "oathbreaker",
            "CommanderId":     commanderId,
            "PartnerId":       partnerId,
        })
    })

    app.Post("/decks/import", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.SendStatus(401)
        }

        label := strings.TrimSpace(c.FormValue("label", ""))
        if label == "" {
            label = "Untitled"
        }

        gamemode := c.FormValue("gamemode", "")
        if !importGamemodes[gamemode] {
            gamemode = ""
        }

        // The review form's hidden inputs are the client's word for what was
        // found, so every id is checked for shape and then for existence before
        // any of it reaches an insert.
        qtyById := map[string]int{}
        sideboardById := map[string]bool{}
        cardIds := []string{}
        for _, value := range importFormValues(c, "cards") {
            match := importCardValue.FindStringSubmatch(value)
            if match == nil {
                continue
            }

            cardId := strings.ToUpper(match[1])
            qty, err := strconv.Atoi(match[2])
            if err != nil || qty < 1 {
                continue
            }
            if qty > deckCardQtyMax {
                qty = deckCardQtyMax
            }

            if _, seen := qtyById[cardId]; !seen {
                cardIds = append(cardIds, cardId)
            }
            qtyById[cardId] = qty
            sideboardById[cardId] = match[3] == "1"
        }

        if len(cardIds) == 0 {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"There were no cards to import\"}")
            return c.SendStatus(422)
        }
        if len(cardIds) > importMaxCards {
            cardIds = cardIds[:importMaxCards]
        }

        db := helpers.ConnectDB()
        existing := models.FilterExistingCardIds(db, cardIds)

        deckUUID := strings.ReplaceAll(uuid.New().String(), "-", "")

        values := []string{}
        args := []interface{}{}
        imported := 0
        for _, cardId := range cardIds {
            if !existing[cardId] {
                continue
            }

            sideboard := 0
            if sideboardById[cardId] {
                sideboard = 1
            }

            values = append(values, "(UNHEX(?), UNHEX(?), UNHEX(?), ?, ?)")
            args = append(args,
                strings.ReplaceAll(uuid.New().String(), "-", ""),
                deckUUID,
                cardId,
                qtyById[cardId],
                sideboard,
            )
            imported++
        }

        if imported == 0 {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"None of those cards could be found\"}")
            return c.SendStatus(422)
        }

        if gamemode == "" {
            helpers.Exec(db, "INSERT INTO Decks (id, user_id, label) VALUES (UNHEX(?), ?, ?)", deckUUID, user.Id, label)
        } else {
            helpers.Exec(db, "INSERT INTO Decks (id, user_id, label, gamemode) VALUES (UNHEX(?), ?, ?, ?)", deckUUID, user.Id, label, gamemode)
        }

        if err := helpers.Exec(db,
            "INSERT INTO Deck_Cards (id, deck_id, card_id, qty, sideboard) VALUES "+strings.Join(values, ", "),
            args...,
        ); err != nil {
            helpers.Exec(db, "DELETE FROM Decks WHERE id = UNHEX(?) AND user_id = ?", deckUUID, user.Id)
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Failed to import the deck\"}")
            return c.SendStatus(500)
        }

        // Only a card that made it into the deck may hold a special status, and
        // no card may hold two of them.
        assigned := map[string]bool{}
        for _, special := range []struct{ column, field string }{
            {"commander_card_id", "commander"},
            {"partner_card_id", "partner"},
            {"oathbreaker_card_id", "oathbreaker"},
        } {
            cardId := strings.ToUpper(strings.TrimSpace(c.FormValue(special.field, "")))
            if cardId == "" || assigned[cardId] || !importCardId.MatchString(cardId) || !existing[cardId] {
                continue
            }
            if _, ok := qtyById[cardId]; !ok {
                continue
            }

            assigned[cardId] = true
            helpers.Exec(db, "UPDATE Decks SET "+special.column+" = UNHEX(?) WHERE id = UNHEX(?) AND user_id = ?", cardId, deckUUID, user.Id)
        }

        c.Response().Header.Set("HX-Redirect", "/decks/"+deckUUID)
        c.Response().Header.Set("HX-Trigger", "{\"flash:toast\": \"Imported "+helpers.EscapeString(label)+"\"}")
        return c.SendStatus(200)
    })
}

// readImportedList prefers an uploaded file and falls back to the textarea, so
// a user who does both still gets the file they picked.
func readImportedList(c *fiber.Ctx) string {
    fileHeader, err := c.FormFile("file")
    if err != nil || fileHeader == nil {
        return c.FormValue("list", "")
    }

    file, err := fileHeader.Open()
    if err != nil {
        return c.FormValue("list", "")
    }
    defer file.Close()

    contents, err := io.ReadAll(io.LimitReader(file, importMaxBytes))
    if err != nil || len(strings.TrimSpace(string(contents))) == 0 {
        return c.FormValue("list", "")
    }

    return string(contents)
}

// resolveImportedCards looks every parsed name up in one batch and splits the
// list into what was found and what was not.
//
// Copies that resolve to the same card are merged, which is what happens when a
// user files a card under several categories. Deck_Cards holds one row per deck
// and card with a single sideboard flag, so a card is only treated as a
// sideboard card when every copy of it was one.
func resolveImportedCards(db *gorm.DB, cards []helpers.ImportedCard) ([]ImportCard, []ImportMiss) {
    found := []ImportCard{}
    missing := []ImportMiss{}

    lookups := []string{}
    seen := map[string]bool{}
    for _, card := range cards {
        for _, name := range helpers.CardNameLookups(card.Name) {
            if key := strings.ToLower(name); !seen[key] {
                seen[key] = true
                lookups = append(lookups, name)
            }
        }
    }

    matches := models.FindCardsByNames(db, lookups)

    position := map[string]int{}
    for _, card := range cards {
        match := models.CardMatch{}
        resolved := false
        for _, name := range helpers.CardNameLookups(card.Name) {
            if hit, ok := matches[strings.ToLower(name)]; ok {
                match = hit
                resolved = true
                break
            }
        }

        if !resolved {
            missing = append(missing, ImportMiss{Name: card.Name, Qty: card.Qty})
            continue
        }

        if index, ok := position[match.Id]; ok {
            found[index].Qty += card.Qty
            if found[index].Qty > deckCardQtyMax {
                found[index].Qty = deckCardQtyMax
            }
            if card.IsCommander {
                found[index].IsCommander = true
            }
            if !card.InSideboard {
                found[index].InSideboard = false
            }
            continue
        }

        qty := card.Qty
        if qty > deckCardQtyMax {
            qty = deckCardQtyMax
        }

        position[match.Id] = len(found)
        found = append(found, ImportCard{
            CardId:      strings.ToUpper(match.Id),
            Name:        match.Name,
            Front:       match.Front,
            Qty:         qty,
            InSideboard: card.InSideboard,
            IsCommander: card.IsCommander,
        })
    }

    return found, missing
}

// importFormValues reads every value submitted under one name. Fiber's
// FormValue returns only the first, and the review form submits one entry per
// card.
func importFormValues(c *fiber.Ctx, key string) []string {
    if form, err := c.MultipartForm(); err == nil && form != nil {
        if values, ok := form.Value[key]; ok {
            return values
        }
    }

    values := []string{}
    for _, value := range c.Request().PostArgs().PeekMulti(key) {
        values = append(values, string(value))
    }

    return values
}

// importNavDecks builds the sidebar's deck list, matching what the other full
// page routes pass to the layout.
func importNavDecks(db *gorm.DB, userId string) (map[string]*GroupedDecks, []models.Deck, int) {
    deckGroups := models.GetDeckGroups(db, userId)
    decks := models.GetDecks(db, "", userId)

    groupedDecks := make(map[string]*GroupedDecks)
    ungroupedDecks := []models.Deck{}

    for i := range deckGroups {
        groupedDecks[deckGroups[i].Id] = &GroupedDecks{Id: deckGroups[i].Id, Label: deckGroups[i].Label, Decks: []models.Deck{}}
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

    return groupedDecks, ungroupedDecks, len(decks)
}
