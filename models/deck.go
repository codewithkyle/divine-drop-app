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
    CardCount int
    Active string
    SleeveImage string
}

type DeckMetadata struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    CardCount int
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

func GetDecks(db *gorm.DB, deckId string, userId string) []Deck {
    var decks []Deck
    db.Raw("SELECT CASE WHEN D.id = UNHEX(?) THEN 'active' ELSE '' END AS Active, HEX(D.id) AS id, HEX(D.commander_card_id) AS commander_card_id, D.label, D.user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id AND DC.sideboard = 0) AS CardCount FROM Decks D WHERE user_id = ?", deckId, userId).Scan(&decks)
    return decks
}

func GetDeck(db *gorm.DB, deckId string, userId string) Deck {
    var deck Deck
    db.Raw("SELECT HEX(D.sleeve_id) as sleeve_id, HEX(D.commander_card_id) AS commander_card_id, HEX(D.oathbreaker_card_id) AS oathbreaker_card_id, HEX(D.id) AS id, D.label, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) AND D.user_id = ?", deckId, userId).Scan(&deck)
    deck.Active = "active"
    return deck
}

func GetDeckByID(db *gorm.DB, deckId string) Deck {
    var deck Deck
    db.Raw("SELECT S.image_url AS SleeveImage, HEX(D.commander_card_id) AS commander_card_id, HEX(D.oathbreaker_card_id) AS oathbreaker_card_id, HEX(D.id) AS id, D.label, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D JOIN Sleeves S ON S.id = D.sleeve_id WHERE D.id = UNHEX(?)", deckId).Scan(&deck)
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
    db.Raw("SELECT HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id AND DC.sideboard = 0) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) GROUP BY D.id, D.user_id", deckId).Scan(&deckMetadata)
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
