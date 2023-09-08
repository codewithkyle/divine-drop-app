package models

import (
	"gorm.io/gorm"
)

type Deck struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    Label string `gorm:"column:label"`
    CommanderCardId string `gorm:"column:commander_card_id"`
    CardCount int
    Active string
}

type DeckMetadata struct {
    Id string `gorm:"column:id;primary_key"`
    UserId string `gorm:"column:user_id"`
    CardCount int
}

func GetDecks(db *gorm.DB, deckId string, userId string) []Deck {
    var decks []Deck
    db.Raw("SELECT CASE WHEN D.id = UNHEX(?) THEN 'active' ELSE '' END AS Active, HEX(D.id) AS id, HEX(D.commander_card_id) AS commander_card_id, D.label, D.user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE user_id = ?", deckId, userId).Scan(&decks)
    return decks
}

func GetDeck(db *gorm.DB, deckId string, userId string) Deck {
    var deck Deck
    db.Raw("SELECT HEX(D.id) AS id, D.label, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) AND D.user_id = ?", deckId, userId).Scan(&deck)
    deck.Active = "active"
    return deck
}

func GetDeckColors(db *gorm.DB, deckId string) []string {
    var colors []string
    db.Raw("SELECT C.color FROM Colors C WHERE C.id IN (SELECT DISTINCT CC.color_id FROM Deck_Cards DC JOIN Card_Colors CC ON DC.card_id = CC.card_id WHERE DC.deck_id = UNHEX(?));", deckId).Scan(&colors)
    return colors
}

func GetDeckMetadata(db *gorm.DB, deckId string) DeckMetadata {
    var deckMetadata DeckMetadata
    db.Raw("SELECT HEX(D.id) AS id, HEX(D.user_id) AS user_id, (SELECT SUM(DC.qty) FROM Deck_Cards DC WHERE DC.deck_id = D.id) AS CardCount FROM Decks D WHERE D.id = UNHEX(?) GROUP BY D.id, D.user_id", deckId).Scan(&deckMetadata)
    return deckMetadata
}

func GetMythicsCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT COUNT(*) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.deck_id = UNHEX(?) AND R.rarity = 'mythic'", deckId).Scan(&count)
    return count
}

func GetUncommonsCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT COUNT(*) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.deck_id = UNHEX(?) AND R.rarity = 'uncommon'", deckId).Scan(&count)
    return count
}

func GetCommonsCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT COUNT(*) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.deck_id = UNHEX(?) AND R.rarity = 'common'", deckId).Scan(&count)
    return count
}

func GetRaresCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT COUNT(*) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Rarities R ON R.id = C.rarity WHERE DC.deck_id = UNHEX(?) AND R.rarity = 'rare'", deckId).Scan(&count)
    return count
}

func GetLandCount(db *gorm.DB, deckId string) int {
    var count int
    db.Raw("SELECT IFNULL(SUM(DC.qty), 0) FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) AND C.type IN ('Land', 'Basic Land', 'Artifact Land', 'Legendary Land')", deckId).Scan(&count)
    return count
}
