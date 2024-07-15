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
        deckCards := models.SearchDeckCards(db, deckId, "", "", "", "")
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

        creatureCount := models.GetDeckCreaturesCount(db, deck.Id)
        artifactCount := models.GetDeckArtifactCount(db, deck.Id)
        enchantmentCount := models.GetDeckEnchantmentCount(db, deck.Id)
        sorceryCount := models.GetDeckSorceryCount(db, deck.Id)
        instantCount := models.GetDeckInstantCount(db, deck.Id)

        totalManaCost := models.GetDeckManaCosts(db, deck.Id)
        totalManaCostJSON, err := json.Marshal(totalManaCost)
        if err != nil {
            totalManaCostJSON = []byte("[]")
        }
        totalCreatureCost := models.GetDeckCreatureCosts(db, deck.Id)
        totalCreatureCostJSON, err := json.Marshal(totalCreatureCost)
        if err != nil {
            totalCreatureCostJSON = []byte("[]")
        }
        totalArtifactCost := models.GetDeckArtifactCosts(db, deck.Id)
        totalArtifactCostJSON, err := json.Marshal(totalArtifactCost)
        if err != nil {
            totalArtifactCostJSON = []byte("[]")
        }
        totalEnchantmentCost := models.GetDeckEnchantmentCosts(db, deck.Id)
        totalEnchantmentCostJSON, err := json.Marshal(totalEnchantmentCost)
        if err != nil {
            totalEnchantmentCostJSON = []byte("[]")
        }
        totalSorceryCost := models.GetDeckSorceryCosts(db, deck.Id)
        totalSorceryCostJSON, err := json.Marshal(totalSorceryCost)
        if err != nil {
            totalSorceryCostJSON = []byte("[]")
        }
        totalInstantCost := models.GetDeckInstantCosts(db, deck.Id)
        totalInstantCostJSON, err := json.Marshal(totalInstantCost)
        if err != nil {
            totalInstantCostJSON = []byte("[]")
        }

        totalWhiteCards := models.GetDeckWhiteCardCount(db, deck.Id)
        totalBlueCards := models.GetDeckBlueCardCount(db, deck.Id)
        totalBlackCards := models.GetDeckBlackCardCount(db, deck.Id)
        totalRedCards := models.GetDeckRedCardCount(db, deck.Id)
        totalGreenCards := models.GetDeckGreenCardCount(db, deck.Id)
        colorCounts := []int{totalWhiteCards, totalBlueCards, totalBlackCards, totalRedCards, totalGreenCards}
        colorCountsJSON, err := json.Marshal(colorCounts)
        if err != nil {
            colorCountsJSON = []byte("[]")
        }

        totalWhiteCreatureCards := models.GetDeckWhiteCreatureCardCount(db, deck.Id)
        totalBlueCreatureCards := models.GetDeckBlueCreatureCardCount(db, deck.Id)
        totalBlackCreatureCards := models.GetDeckBlackCreatureCardCount(db, deck.Id)
        totalRedCreatureCards := models.GetDeckRedCreatureCardCount(db, deck.Id)
        totalGreenCreatureCards := models.GetDeckGreenCreatureCardCount(db, deck.Id)
        creatureColorCounts := []int{totalWhiteCreatureCards, totalBlueCreatureCards, totalBlackCreatureCards, totalRedCreatureCards, totalGreenCreatureCards}
        creatureColorCountsJSON, err := json.Marshal(creatureColorCounts)
        if err != nil {
            creatureColorCountsJSON = []byte("[]")
        }

        totalWhiteEnchantmentCards := models.GetDeckWhiteEnchantmentCardCount(db, deck.Id)
        totalBlueEnchantmentCards := models.GetDeckBlueEnchantmentCardCount(db, deck.Id)
        totalBlackEnchantmentCards := models.GetDeckBlackEnchantmentCardCount(db, deck.Id)
        totalRedEnchantmentCards := models.GetDeckRedEnchantmentCardCount(db, deck.Id)
        totalGreenEnchantmentCards := models.GetDeckGreenEnchantmentCardCount(db, deck.Id)
        enchantmentColorCounts := []int{totalWhiteEnchantmentCards, totalBlueEnchantmentCards, totalBlackEnchantmentCards, totalRedEnchantmentCards, totalGreenEnchantmentCards}
        enchantmentColorCountsJSON, err := json.Marshal(enchantmentColorCounts)
        if err != nil {
            enchantmentColorCountsJSON = []byte("[]")
        }

        totalWhiteSorceryCards := models.GetDeckWhiteSorceryCardCount(db, deck.Id)
        totalBlueSorceryCards := models.GetDeckBlueSorceryCardCount(db, deck.Id)
        totalBlackSorceryCards := models.GetDeckBlackSorceryCardCount(db, deck.Id)
        totalRedSorceryCards := models.GetDeckRedSorceryCardCount(db, deck.Id)
        totalGreenSorceryCards := models.GetDeckGreenSorceryCardCount(db, deck.Id)
        sorceryColorCounts := []int{totalWhiteSorceryCards, totalBlueSorceryCards, totalBlackSorceryCards, totalRedSorceryCards, totalGreenSorceryCards}
        sorceryColorCountsJSON, err := json.Marshal(sorceryColorCounts)
        if err != nil {
            sorceryColorCountsJSON = []byte("[]")
        }

        totalWhiteInstantCards := models.GetDeckWhiteInstantCardCount(db, deck.Id)
        totalBlueInstantCards := models.GetDeckBlueInstantCardCount(db, deck.Id)
        totalBlackInstantCards := models.GetDeckBlackInstantCardCount(db, deck.Id)
        totalRedInstantCards := models.GetDeckRedInstantCardCount(db, deck.Id)
        totalGreenInstantCards := models.GetDeckGreenInstantCardCount(db, deck.Id)
        instantColorCounts := []int{totalWhiteInstantCards, totalBlueInstantCards, totalBlackInstantCards, totalRedInstantCards, totalGreenInstantCards}
        instantColorCountsJSON, err := json.Marshal(instantColorCounts)
        if err != nil {
            instantColorCountsJSON = []byte("[]")
        }

        return c.Render("pages/deck-stats/index", fiber.Map{
            "EnchantmentColorCounts": string(enchantmentColorCountsJSON),
            "SorceryColorCounts": string(sorceryColorCountsJSON),
            "InstantColorCounts": string(instantColorCountsJSON),
            "CreatureColorCounts": string(creatureColorCountsJSON),
            "ColorCounts": string(colorCountsJSON),
            "TotalArtifactCost": string(totalArtifactCostJSON),
            "TotalEnchantmentCost": string(totalEnchantmentCostJSON),
            "TotalSorceryCost": string(totalSorceryCostJSON),
            "TotalInstantCost": string(totalInstantCostJSON),
            "TotalCreatureCost": string(totalCreatureCostJSON),
            "TotalManaCosts": string(totalManaCostJSON),
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
