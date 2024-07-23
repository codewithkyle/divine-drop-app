package models

import (
	"gorm.io/gorm"
)

type Deck struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    Label string `gorm:"column:label"`
    CommanderCardId string `gorm:"column:commander_card_id"`
    OathbreakerCardId string `gorm:"column:oathbreaker_card_id"`
    SleeveId string `gorm:"column:sleeve_id"`
    GroupId string `gorm:"column:group_id"`
    CardCount int
    Active string
    SleeveImage string
    Gamemode string
    IsLegal bool
}

type DeckMetadata struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    CardCount int
    Budget int 
    Label string
    Gamemode string
}

type Sleeve struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    Image string `gorm:"column:image_url"`
    IsVideo bool `gorm:"column:is_video;type:tinyint"`
    DeckId string
    Selected bool
}

type DeckSleeve struct {
    Id string `gorm:"column:id;primary_key"`
    DeckId string `gorm:"column:deck_id"`
    SleeveId string `gorm:"column:sleeve_id"`
    IsVideo bool `gorm:"column:is_video;type:tinyint"`
    Image string
}

type DeckGroup struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    Label string
}

type TypeCost struct {
    TMC int
    Count int
}

type DeckCardTypeCount struct {
    TMC int
    Count int
    CardType string
}

type DeckCardTMCCount struct {
    TMC int
    Count int
    InstantCount int
    SorceryCount int
    EnchantmentCount int
    CreatureCount int
    ArtifactCount int
}

type DeckCardColorAndTypeCount struct {
    ColorId int
    InstantCount int
    SorceryCount int
    EnchantmentCount int
    CreatureCount int
}

type DeckCardColorCount struct {
    WhiteCount int
    BlueCount int
    BlackCount int
    RedCount int
    GreenCount int
}

func GetDeckGroups(db *gorm.DB, userId string) []DeckGroup {
    var groups []DeckGroup
    db.Raw("SELECT HEX(id) as id, user_id, label FROM Deck_Groups WHERE user_id = ?", userId).Scan(&groups)
    return groups
}

func GetDeckGroupByID(db *gorm.DB, groupId string, userId string) DeckGroup {
    var groups DeckGroup
    db.Raw("SELECT HEX(id) as id, user_id, label FROM Deck_Groups WHERE user_id = ? AND id = UNHEX(?) LIMIT 1", userId, groupId).Scan(&groups)
    return groups
}

func GetDecks(db *gorm.DB, deckId string, userId string) []Deck {
    var decks []Deck
    db.Raw("SELECT CASE WHEN D.id = UNHEX(?) THEN 'active' ELSE '' END AS Active, HEX(D.deck_group_id) as group_id, HEX(D.id) AS id, HEX(D.commander_card_id) AS commander_card_id, D.label, D.user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id AND DC.sideboard = 0) AS CardCount FROM Decks D WHERE user_id = ?", deckId, userId).Scan(&decks)
    return decks
}

func GetDeck(db *gorm.DB, deckId string, userId string) Deck {
    var deck Deck
    db.Raw("SELECT D.gamemode, HEX(D.sleeve_id) as sleeve_id, HEX(D.commander_card_id) AS commander_card_id, HEX(D.oathbreaker_card_id) AS oathbreaker_card_id, HEX(D.id) AS id, D.label, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) AND D.user_id = ?", deckId, userId).Scan(&deck)
    deck.Active = "active"
    return deck
}

func GetDeckByID(db *gorm.DB, deckId string) Deck {
    var deck Deck
    db.Raw("SELECT D.gamemode, D.user_id, S.image_url AS SleeveImage, HEX(D.commander_card_id) AS commander_card_id, HEX(D.oathbreaker_card_id) AS oathbreaker_card_id, HEX(D.id) AS id, D.label, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D LEFT JOIN Sleeves S ON S.id = D.sleeve_id WHERE D.id = UNHEX(?)", deckId).Scan(&deck)
    deck.Active = "active"
    return deck
}

func GetDeckColors(db *gorm.DB, deckId string) (bool, bool, bool, bool, bool) {
    var colors []string
    db.Raw("SELECT C.color FROM Colors C WHERE C.id IN (SELECT DISTINCT CC.color_id FROM Deck_Cards DC JOIN Card_Colors CC ON DC.card_id = CC.card_id WHERE DC.sideboard = 0 AND DC.deck_id = UNHEX(?));", deckId).Scan(&colors)

    containsW := false
    containsU := false
    containsB := false
    containsR := false
    containsG := false
    for _, color := range colors {
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

    return containsW, containsU, containsB, containsR, containsG
}

func GetDeckMetadata(db *gorm.DB, deckId string) DeckMetadata {
    var deckMetadata DeckMetadata
    db.Raw("SELECT D.gamemode, D.label, IFNULL(D.budget, 0) AS budget, HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id AND DC.sideboard = 0) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) GROUP BY D.id, D.user_id", deckId).Scan(&deckMetadata)
    return deckMetadata
}

func GetMythicsCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.sideboard = 0 AND DC.deck_id = UNHEX(?) AND R.rarity = 'mythic'", deckId).Scan(&count)
    return count
}

func GetUncommonsCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.sideboard = 0 AND DC.deck_id = UNHEX(?) AND R.rarity = 'uncommon'", deckId).Scan(&count)
    return count
}

func GetCommonsCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.sideboard = 0 AND DC.deck_id = UNHEX(?) AND R.rarity = 'common'", deckId).Scan(&count)
    return count
}

func GetRaresCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.sideboard = 0 AND DC.deck_id = UNHEX(?) AND R.rarity = 'rare'", deckId).Scan(&count)
    return count
}

func GetLandCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.sideboard = 0 AND DC.deck_id = UNHEX(?) AND C.type IN ('Land', 'Basic Land', 'Artifact Land', 'Legendary Land')", deckId).Scan(&count)
    return count
}

func GetSideboardCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC WHERE DC.deck_id = UNHEX(?) AND DC.sideboard = 1", deckId).Scan(&count)
    return count
}

func GetSleeves(db *gorm.DB, userId string) []Sleeve {
    var sleeves []Sleeve
    db.Raw("SELECT user_id, HEX(id) as Id, image_url, is_video FROM Sleeves WHERE user_id = ?", userId).Scan(&sleeves)
    return sleeves
}

func GetSleeve(db *gorm.DB, userId string, sleeveId string) Sleeve {
    var sleeve Sleeve
    db.Raw("SELECT user_id, HEX(id) as id, image_url FROM Sleeves WHERE user_id = ? AND id = UNHEX(?) LIMIT 1", userId, sleeveId).Scan(&sleeve)
    return sleeve
}

func GetDeckCost(db *gorm.DB, deckId string) float32 {
    cost := 0
    db.Raw("SELECT IFNULL(SUM(C.price * DC.qty), 0) AS total_price FROM Deck_Cards DC LEFT JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) AND C.name NOT IN ('forest', 'plains', 'mountain', 'island') AND DC.sideboard = 0", deckId).Scan(&cost)
    return float32(cost) / 100
}

func GetDeckCardManaCountsByType(db *gorm.DB, deckId string) map[int]*DeckCardTMCCount {
    raw := []DeckCardTypeCount{}
    db.Raw(`
        SELECT 
            C.totalManaCost as TMC,
            SUM(DC.qty) as Count, 
            CASE 
                WHEN C.type NOT LIKE '%Creature%' AND C.type NOT LIKE '%Artifact%' AND C.type NOT LIKE '%Enchantment%' AND C.type NOT LIKE '%Sorcery%' AND C.type NOT LIKE '%Instant%' THEN 'Other'
                WHEN C.type LIKE '%Creature%' THEN 'Creature'
                WHEN C.type LIKE '%Artifact%' THEN 'Artifact'
                WHEN C.type LIKE '%Enchantment%' THEN 'Enchantment'
                WHEN C.type LIKE '%Sorcery%' THEN 'Sorcery'
                WHEN C.type LIKE '%Instant%' THEN 'Instant'
            END as CardType
        FROM 
            Deck_Cards DC 
        JOIN 
            Cards C 
        ON 
            C.id = DC.card_id 
        WHERE 
            DC.deck_id = UNHEX(?) 
            AND DC.sideboard = 0 
            AND C.type NOT LIKE '%Land%'
        GROUP BY 
            C.totalManaCost, CardType 
        ORDER BY 
            C.totalManaCost ASC
    `, deckId).Scan(&raw)
    counts := make(map[int]*DeckCardTMCCount)
    for i := 0; i < 11; i++ {
        counts[i] = &DeckCardTMCCount{
            TMC: i,
            Count: 0,
            InstantCount: 0,
            SorceryCount: 0,
            EnchantmentCount: 0,
            CreatureCount: 0,
            ArtifactCount: 0,
        }
    }
    for _, count := range raw {
        if _, ok := counts[count.TMC]; ok {
            counts[count.TMC].Count += count.Count
            switch count.CardType {
                case "Instant":
                    counts[count.TMC].InstantCount += count.Count
                case "Sorcery":
                    counts[count.TMC].SorceryCount += count.Count
                case "Enchantment":
                    counts[count.TMC].EnchantmentCount += count.Count
                case "Creature":
                    counts[count.TMC].CreatureCount += count.Count
                case "Artifact":
                    counts[count.TMC].ArtifactCount += count.Count
            }
        }
    }
    return counts
}

func GetDeckCardCountsByColor(db *gorm.DB, deckId string) DeckCardColorCount {
    count := DeckCardColorCount{}
    db.Raw(`
        SELECT
            SUM(CASE WHEN CLR.color_id LIKE 1 THEN DC.qty ELSE 0 END) AS WhiteCount,
            SUM(CASE WHEN CLR.color_id LIKE 2 THEN DC.qty ELSE 0 END) AS BlueCount,
            SUM(CASE WHEN CLR.color_id LIKE 3 THEN DC.qty ELSE 0 END) AS BlackCount,
            SUM(CASE WHEN CLR.color_id LIKE 4 THEN DC.qty ELSE 0 END) AS RedCount,
            SUM(CASE WHEN CLR.color_id LIKE 5 THEN DC.qty ELSE 0 END) AS GreenCount
        FROM Deck_Cards DC 
        JOIN Cards C ON C.id = DC.card_id 
        JOIN Card_Colors CLR ON CLR.card_id = C.id 
        WHERE DC.deck_id = UNHEX(?) AND DC.sideboard = 0 
    `, deckId).Scan(&count)
    return count
}

func GetDeckCardCountsByColorAndType(db *gorm.DB, deckId string) map[int]DeckCardColorAndTypeCount {
    raw := []DeckCardColorAndTypeCount{}
    db.Raw(`
        SELECT CLR.color_id AS ColorId,
            SUM(CASE WHEN C.type LIKE '%Instant%' THEN DC.qty ELSE 0 END) AS InstantCount,
            SUM(CASE WHEN C.type LIKE '%Sorcery%' THEN DC.qty ELSE 0 END) AS SorceryCount,
            SUM(CASE WHEN C.type LIKE '%Enchantment%' THEN DC.qty ELSE 0 END) AS EnchantmentCount,
            SUM(CASE WHEN C.type LIKE '%Creature%' THEN DC.qty ELSE 0 END) AS CreatureCount
        FROM Deck_Cards DC
        JOIN Cards C ON C.id = DC.card_id
        JOIN Card_Colors CLR ON CLR.card_id = C.id
        WHERE DC.deck_id = UNHEX(?)
        AND DC.sideboard = 0
        GROUP BY CLR.color_id
        ORDER BY CLR.color_id ASC
    `, deckId).Scan(&raw)
    counts := make(map[int]DeckCardColorAndTypeCount)
    for _, count := range raw {
        counts[count.ColorId] = count
    }
    return counts
}
