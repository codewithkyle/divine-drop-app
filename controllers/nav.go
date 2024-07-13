package controllers

import (
	"app/helpers"
	"app/models"

        "strings"

	"github.com/gofiber/fiber/v2"

        "github.com/google/uuid"
)

type GroupedDecks struct {
    Id string
    Label string
    Decks []models.Deck
}

func NavControllers(app *fiber.App){

    app.Post("/groups", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.Query("active-deck-id", "")
        label := c.Get("HX-Prompt", "Untitled")
        if strings.Trim(label, " ") == "" {
            label = "Untitled"
        }

        db := helpers.ConnectDB()

        groupId := strings.ReplaceAll(uuid.New().String(), "-", "")

        db.Exec("INSERT INTO Deck_Groups (id, user_id, label) VALUES (UNHEX(?), ?, ?)", groupId, user.Id, label)

        deckGroups := models.GetDeckGroups(db, user.Id)
        decks := models.GetDecks(db, activeDeckId, user.Id)

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

        return c.Render("partials/nav/decks-opened", fiber.Map{
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "ActiveDeckId": activeDeckId,
        })
    })

    app.Delete("/groups/:groupId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        activeDeckId := c.Query("active-deck-id", "")
        groupId := c.Params("groupId")

        db := helpers.ConnectDB()

        group := models.GetDeckGroupByID(db, groupId, user.Id)
        if group.Id == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Failed to delete folder\"}")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET deck_group_id = null WHERE deck_group_id = UNHEX(?) AND user_id = ?", group.Id, user.Id)
        db.Exec("DELETE FROM Deck_Groups WHERE id = UNHEX(?) AND user_id = ?", group.Id, user.Id)

        deckGroups := models.GetDeckGroups(db, user.Id)
        decks := models.GetDecks(db, activeDeckId, user.Id)

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

        return c.Render("partials/nav/decks-opened", fiber.Map{
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "ActiveDeckId": activeDeckId,
        })
    })

    app.Put("/groups/:groupId/:deckId", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        groupId := c.Params("groupId")
        deckId := c.Params("deckId")
        activeDeckId := c.Query("active-deck-id", "")

        db := helpers.ConnectDB()

        group := models.GetDeckGroupByID(db, groupId, user.Id)
        if group.Id == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Failed to move deck into folder\"}")
            return c.SendStatus(404)
        }

        deck := models.GetDeckByID(db, deckId)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Failed to move deck into folder\"}")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET deck_group_id = UNHEX(?) WHERE id = UNHEX(?) AND user_id = ?", group.Id, deck.Id, user.Id)

        deckGroups := models.GetDeckGroups(db, user.Id)
        decks := models.GetDecks(db, activeDeckId, user.Id)

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

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Moved " + deck.Label + " to " + group.Label + "\"}")
        return c.Render("partials/nav/decks-opened", fiber.Map{
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "ActiveDeckId": activeDeckId,
        })
    })

    app.Delete("/decks/:deckId/group", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        deckId := c.Params("deckId")
        activeDeckId := c.Query("active-deck-id", "")

        db := helpers.ConnectDB()

        deck := models.GetDeckByID(db, deckId)
        if deck.Id == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Failed to move deck into folder\"}")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Decks SET deck_group_id = null WHERE id = UNHEX(?) AND user_id = ?", deck.Id, user.Id)

        deckGroups := models.GetDeckGroups(db, user.Id)
        decks := models.GetDecks(db, activeDeckId, user.Id)

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

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Removed " + deck.Label + " from folder\"}")
        return c.Render("partials/nav/decks-opened", fiber.Map{
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
            "ActiveDeckId": activeDeckId,
        })
    })

    app.Patch("/groups/:groupId/label", func(c *fiber.Ctx) error {
        user, err := helpers.GetUserFromSession(c)
        if err != nil {
            c.Response().Header.Add("HX-Redirect", "/sign-in")
            return c.Send(nil)
        }

        groupId := c.Params("groupId")
        label := c.Get("HX-Prompt", "Untitled");
        if strings.Trim(label, " ") == "" {
            label = "Untitled"
        }
        activeDeckId := c.Query("active-deck-id", "")

        db := helpers.ConnectDB()

        group := models.GetDeckGroupByID(db, groupId, user.Id)
        if group.Id == "" {
            c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Failed to rename folder\"}")
            return c.SendStatus(404)
        }

        db.Exec("UPDATE Deck_Groups SET label = ? WHERE id = UNHEX(?) AND user_id = ?", label, group.Id, user.Id)

        deckGroups := models.GetDeckGroups(db, user.Id)
        decks := models.GetDecks(db, activeDeckId, user.Id)

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

        c.Response().Header.Set("Hx-Trigger", "{\"flash:toast\": \"Renamed folder to " + label + "\"}")
        return c.Render("partials/nav/decks-opened", fiber.Map{
            "GroupedDecks": groupedDecks,
            "UngroupedDecks": ungroupedDecks,
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
