package controllers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"app/helpers"
	"app/models"
)


func DeckStatsControllers(app *fiber.App){
    app.Get("/decks/:id/stats", func(c *fiber.Ctx) error {
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
        deckCards := models.SearchDeckCards(db, deckId, "", "", "", "", "")
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

        deckGroups := models.GetDeckGroups(db, user.Id)

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
        }


        totalCardTypeCost := models.GetDeckCardManaCountsByType(db, deck.Id)
        totalCardTypeCostJSON, err := json.Marshal(totalCardTypeCost)
        if err != nil {
            totalCardTypeCostJSON = []byte("{}")
        }

        creatureCount := 0
        artifactCount := 0
        enchantmentCount := 0
        sorceryCount := 0
        instantCount := 0
        for i := range totalCardTypeCost {
            creatureCount += totalCardTypeCost[i].CreatureCount
            artifactCount += totalCardTypeCost[i].ArtifactCount
            enchantmentCount += totalCardTypeCost[i].EnchantmentCount
            sorceryCount += totalCardTypeCost[i].SorceryCount
            instantCount += totalCardTypeCost[i].InstantCount
        }

        colorCounts := models.GetDeckCardCountsByColor(db, deck.Id)
        colorCountsArr := []int{colorCounts.WhiteCount, colorCounts.BlueCount, colorCounts.BlackCount, colorCounts.RedCount, colorCounts.GreenCount}
        colorCountsJSON, err := json.Marshal(colorCountsArr)
        if err != nil {
            colorCountsJSON = []byte("[]")
        }

        colorAndTypeCounts := models.GetDeckCardCountsByColorAndType(db, deck.Id)
        creatureColorCounts := []int{colorAndTypeCounts[1].CreatureCount, colorAndTypeCounts[2].CreatureCount, colorAndTypeCounts[3].CreatureCount, colorAndTypeCounts[4].CreatureCount, colorAndTypeCounts[5].CreatureCount}
        creatureColorCountsJSON, err := json.Marshal(creatureColorCounts)
        if err != nil {
            creatureColorCountsJSON = []byte("[]")
        }

        enchantmentColorCounts := []int{colorAndTypeCounts[1].EnchantmentCount, colorAndTypeCounts[2].EnchantmentCount, colorAndTypeCounts[3].EnchantmentCount, colorAndTypeCounts[4].EnchantmentCount, colorAndTypeCounts[5].EnchantmentCount}
        enchantmentColorCountsJSON, err := json.Marshal(enchantmentColorCounts)
        if err != nil {
            enchantmentColorCountsJSON = []byte("[]")
        }

        sorceryColorCounts := []int{colorAndTypeCounts[1].SorceryCount, colorAndTypeCounts[2].SorceryCount, colorAndTypeCounts[3].SorceryCount, colorAndTypeCounts[4].SorceryCount, colorAndTypeCounts[5].SorceryCount}
        sorceryColorCountsJSON, err := json.Marshal(sorceryColorCounts)
        if err != nil {
            sorceryColorCountsJSON = []byte("[]")
        }

        instantColorCounts := []int{colorAndTypeCounts[1].InstantCount, colorAndTypeCounts[2].InstantCount, colorAndTypeCounts[3].InstantCount, colorAndTypeCounts[4].InstantCount, colorAndTypeCounts[5].InstantCount}
        instantColorCountsJSON, err := json.Marshal(instantColorCounts)
        if err != nil {
            instantColorCountsJSON = []byte("[]")
        }

        return c.Render("pages/deck-stats/index", fiber.Map{
            "CardsTotalManaCosts": string(totalCardTypeCostJSON),
            "EnchantmentColorCounts": string(enchantmentColorCountsJSON),
            "SorceryColorCounts": string(sorceryColorCountsJSON),
            "InstantColorCounts": string(instantColorCountsJSON),
            "CreatureColorCounts": string(creatureColorCountsJSON),
            "ColorCounts": string(colorCountsJSON),
            "ArtifactCount": artifactCount,
            "EnchantmentCount": enchantmentCount,
            "SorceryCount": sorceryCount,
            "InstantCount": instantCount,
            "CreatureCount": creatureCount,
            "IsOverBudget": overBudget,
            "Budget": fmt.Sprintf("%.2f", float32(deckMetadata.Budget) / 100),
            "DeckPrice": fmt.Sprintf("%.2f", cost),
            "Page": "deck-editor",
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
            "BannerArtUrl": url.QueryEscape(bannerArt),
            "MythicsCount": mythicsCount,
            "UncommonsCount": uncommonsCount,
            "CommonsCount": commonsCount,
            "RaresCount": raresCount,
            "LandCount": landCount,
            "SideboardCount": sideboardCount,
        }, "layouts/main")
    }) 
}
