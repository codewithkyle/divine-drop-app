package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Card struct {
    Id            string `gorm:"column:id;primary_key;type:binary(16)"`
    Layout        string `gorm:"column:layout"`
    Front         string `gorm:"column:front"`
    Back          string `gorm:"column:back"`
    Type          string `gorm:"column:type"`
    Toughness     int    `gorm:"column:toughness"`
    Power         int    `gorm:"column:power"`
    TotalManaCost int    `gorm:"column:totalManaCost"`
    Art           string `gorm:"column:art"`
    Standard      bool   `gorm:"column:standard"`
    Future        bool   `gorm:"column:future"`
    Historic      bool   `gorm:"column:historic"`
    Gladiator     bool   `gorm:"column:gladiator"`
    Pioneer       bool   `gorm:"column:pioneer"`
    Explorer      bool   `gorm:"column:explorer"`
    Modern        bool   `gorm:"column:modern"`
    Legacy        bool   `gorm:"column:legacy"`
    Pauper        bool   `gorm:"column:pauper"`
    Vintage       bool   `gorm:"column:vintage"`
    Penny         bool   `gorm:"column:penny"`
    Commander     bool   `gorm:"column:commander"`
    Oathbreaker   bool   `gorm:"column:oathbreaker"`
    Brawl         bool   `gorm:"column:brawl"`
    HistoricBrawl bool   `gorm:"column:historicbrawl"`
    Alchemy       bool   `gorm:"column:alchemy"`
    PauperCommander bool `gorm:"column:paupercommander"`
    Duel          bool   `gorm:"column:duel"`
    OldSchool     bool   `gorm:"column:oldschool"`
    Premodern     bool   `gorm:"column:premodern"`
    Predh         bool   `gorm:"column:predh"`
    Rarity        uint8  `gorm:"column:rarity"`
    ManaCost      string `gorm:"column:manaCost"`
    Name          string
}

type DeckCard struct {
    Id string `gorm:"column:id;primary_key"`
    DeckId string `gorm:"column:deck_id"`
    CardId string `gorm:"column:card_id"`
    Qty    uint8    `gorm:"column:qty"`
    Front string 
    Name string
    Art string
}

func SearchCardsByName(db *gorm.DB, name string, offset int, limit int) []Card {
    var cards []Card
    db.Raw("SELECT C.front, HEX(C.id) AS id, CN.name FROM Cards AS C JOIN Card_Names AS CN ON C.id = CN.card_id WHERE CN.name LIKE ? LIMIT ? OFFSET ?", name, limit, offset).Scan(&cards)
    return cards
}

func GetDeckCards (db *gorm.DB, deckId string) []DeckCard {
    var cards []DeckCard
    db.Raw("SELECT C.art, C.front, HEX(DC.card_id) AS id, CN.name, DC.qty FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id JOIN Card_Names CN ON CN.card_id = DC.card_id WHERE DC.deck_id = UNHEX(?) ORDER BY dateCreated DESC", deckId).Scan(&cards)
    return cards
}

func FilterCards(db *gorm.DB, name string, sort string, mana []string, types []string, offset int, limit int) []Card {
    var cards []Card
    query := "SELECT C.front, C.back, HEX(C.id) AS id, CN.name FROM Cards AS C JOIN Card_Names AS CN ON C.id = CN.card_id WHERE 1=1 "

    sortColumn := "CN.name"
    switch sort {
        case "name":
            sortColumn = "CN.name"
        case "tmc":
            sortColumn = "C.totalManaCost DESC"
        case "power":
            sortColumn = "C.power DESC"
        case "toughness":
            sortColumn = "C.toughness DESC"
    }
    orderBy := "ORDER BY " + sortColumn

    manaCheck := []string{}
    params := map[string]interface{}{
        "limit": limit,
        "offset": offset,
    }
    includeColorless := false
    if len(mana) > 0 {
        for i := 0; i < len(mana); i++ {
            if mana[i] == "C" {
                includeColorless = true
            } else {
                manaCheck = append(manaCheck, "@mana" + fmt.Sprint(i))
                params["mana" + fmt.Sprint(i)] = mana[i]
            }
        }
    }
    colorLogic := "C.id NOT IN (SELECT cc.card_id FROM Card_Colors cc JOIN Colors c ON cc.color_id = c.id WHERE c.color <> " + strings.Join(manaCheck, " AND c.color <> ") + ") "
    if includeColorless == false {
        colorLogic += "AND EXISTS (SELECT 1 FROM Card_Colors cc WHERE cc.card_id = C.id) "
    }
    if len(mana) == 1 && includeColorless {
        colorLogic = "NOT EXISTS (SELECT 1 FROM Card_Colors cc WHERE cc.card_id = C.id) "
    }

    typeCheck := []string{}
    if len(types) > 0 {
        for i := 0; i < len(types); i++ {
            typeCheck = append(typeCheck, "@type" + fmt.Sprint(i))
            params["type" + fmt.Sprint(i)] = types[i]
        }
    }
    typeLogic := "(C.type = " + strings.Join(typeCheck, " OR C.type = ") + ") "

    if name == "%%" {
        if len(mana) > 0 {
            query += "AND " + colorLogic
        }
        if len(types) > 0 {
            query += "AND " + typeLogic
        }
    } else {
        params["name"] = name
        query += "AND CN.name LIKE @name "
        if len(mana) > 0 {
            query += "AND " + colorLogic
        }
        if len(types) > 0 {
            query += "AND " + typeLogic
        }
    }
    query += orderBy + " LIMIT @limit OFFSET @offset"
    db.Raw(query, params).Scan(&cards)
    return cards
}

func GetCardTypes(db *gorm.DB) []string {
    var types []string
    db.Raw("SELECT DISTINCT type FROM Cards").Scan(&types)
    return types
}

func SearchCardTypes(db *gorm.DB, name string) []string {
    name = "%" + strings.Trim(name, " ") + "%"
    var types []string
    db.Raw("SELECT DISTINCT type FROM Cards WHERE type LIKE ?", name).Scan(&types)
    return types
}
