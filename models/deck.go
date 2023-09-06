package models

import "gorm.io/gorm"

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
