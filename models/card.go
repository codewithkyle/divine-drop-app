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
    ActiveDeckId string
}

type DeckCard struct {
    Id string `gorm:"column:id;primary_key"`
    DeckId string `gorm:"column:deck_id"`
    CardId string `gorm:"column:card_id"`
    Qty    uint8    `gorm:"column:qty"`
    Front string 
    Back string
    Name string
    Art string
    DateCreated string `gorm:"column:dateCreated;type:datetime"`
    IsCommander bool
    IsOathbreaker bool
    InSideboard bool `gorm:"column:sideboard;type:tinyint"`
    Print int32 `gorm:"column:print;type:mediumint"`
}

type DeckCardMetadata struct {
    CardId string `gorm:"column:card_id"`
    Qty    uint8    `gorm:"column:qty"`
    Front string 
    Back string
    Name string
    IsCommander bool
    IsOathbreaker bool
    InSideboard bool `gorm:"column:sideboard;type:tinyint"`
    Print int32 `gorm:"column:print;type:mediumint"`
}

func SearchCardsByName(db *gorm.DB, name string, offset int, limit int) []Card {
    name = "%" + strings.Trim(name, " ") + "%"
    var cards []Card
    db.Raw("SELECT C.front, HEX(C.id) AS id, C.name FROM Cards AS C WHERE C.name LIKE ? LIMIT ? OFFSET ?", name, limit, offset).Scan(&cards)
    return cards
}

func GetDeckCards (db *gorm.DB, deckId string) []DeckCard {
    var cards []DeckCard
    db.Raw("SELECT DC.print, DC.sideboard, DC.dateCreated, C.art, C.front, C.back, HEX(DC.id) AS id, HEX(DC.card_id) AS card_id, (SELECT c.name FROM Card_Names c WHERE C.id = c.card_id LIMIT 1) AS name, DC.qty FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) ORDER BY dateCreated DESC", deckId).Scan(&cards)
    return cards
}

func GetDeckCardsMetadata (db *gorm.DB, deckId string) []DeckCardMetadata {
    var cards []DeckCardMetadata
    db.Raw("SELECT DC.print, DC.sideboard, HEX(DC.card_id) AS card_id, C.front, C.back, (SELECT c.name FROM Card_Names c WHERE C.id = c.card_id LIMIT 1) AS name, DC.qty FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) ORDER BY dateCreated DESC", deckId).Scan(&cards)
    return cards
}

func SearchDeckCards(db *gorm.DB, deckId string, name string, sort string, filter string) []DeckCard {
    name = "%" + strings.Trim(name, " ") + "%"
    sortColumn := "name"
    switch sort {
        case "name":
            sortColumn = "C.name"
        case "tmc":
            sortColumn = "C.totalManaCost DESC"
        case "power":
            sortColumn = "C.power DESC"
        case "toughness":
            sortColumn = "C.toughness DESC"
    }
    filterLogic := "AND 1=1"
    switch filter {
        case "creatures":
            filterLogic = "AND C.type = 'Creature'"
        case "enchantments":
            filterLogic = "AND C.type = 'Enchantment'"
        case "artifacts":
            filterLogic = "AND C.type = 'Artifact'"
        case "lands":
            filterLogic = "AND C.type LIKE '%Land%'"
        case "instants":
            filterLogic = "AND C.type = 'Instant'"
        case "sorceries":
            filterLogic = "AND C.type = 'Sorcery'"
    }
    var cards []DeckCard
    db.Raw("SELECT DC.print, DC.sideboard, HEX(DC.deck_id) AS deck_id, HEX(C.id) as card_id, DC.dateCreated, C.art, C.front, C.back, HEX(DC.id) AS id, DC.qty, C.name FROM Deck_Cards DC JOIN Cards C ON C.id = DC.card_id WHERE DC.deck_id = UNHEX(?) AND C.name LIKE ? " + filterLogic + " GROUP BY DC.id ORDER BY " + sortColumn, deckId, name).Scan(&cards)
    return cards
}

func FilterCards(db *gorm.DB, name string, sort string, mana []string, types []string, subtypes []string, keywords []string, rarity string, legality string, offset int, limit int) []Card {
    var cards []Card
    query := "SELECT C.front, C.back, HEX(C.id) AS id, C.name FROM Cards AS C "
    name = strings.Trim(name, " ")

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

    subtypeCheck := []string{}
    if len(subtypes) > 0 {
        for i := 0; i < len(subtypes); i++ {
            subtypeCheck = append(subtypeCheck, "@subtype" + fmt.Sprint(i))
            params["subtype" + fmt.Sprint(i)] = subtypes[i]
        }
    }
    subtypeLogic := "C.id IN (SELECT cs.card_id FROM Card_Subtypes cs WHERE cs.subtype = " + strings.Join(subtypeCheck, " OR cs.subtype = ") + ") "

    keywordCheck := []string{}
    if len(keywords) > 0 {
        for i := 0; i < len(keywords); i++ {
            keywordCheck = append(keywordCheck, "@keyword" + fmt.Sprint(i))
            params["keyword" + fmt.Sprint(i)] = keywords[i]
        }
    }
    keywordLogic := "C.id IN (SELECT ck.card_id FROM Card_Keywords ck WHERE ck.keyword = " + strings.Join(keywordCheck, " OR ck.keyword = ") + ") "

    if rarity != "any" && rarity != "" {
        params["rarity"] = rarity
        query += "JOIN Rarities R ON C.rarity = R.id AND R.rarity = @rarity "
    }

    if name != "" {
        query += "LEFT JOIN (SELECT DISTINCT card_id FROM Card_Texts CT WHERE MATCH(`text`) AGAINST (@fts IN NATURAL LANGUAGE MODE)) AS ftx_card ON ftx_card.card_id = C.id ";
        query += "WHERE C.name LIKE @name "
        params["name"] = "%" + strings.Trim(name, " ") + "%"
        params["fts"] = name
    } else {
        query += "WHERE 1=1 "
    }
    
    if len(mana) > 0 {
        query += "AND " + colorLogic
    }
    if len(types) > 0 {
        query += "AND " + typeLogic
    }
    if len(subtypes) > 0 {
        query += "AND " + subtypeLogic
    }
    if len(keywords) > 0 {
        query += "AND " + keywordLogic
    }

    if legality != "any" && legality != "" {
        switch legality {
            case "standard":
                query += "AND C.standard = 1 "
            case "future":
                query += "AND C.future = 1 "
            case "historic":
                query += "AND C.historic = 1 "
            case "gladiator":
                query += "AND C.gladiator = 1 "
            case "pioneer":
                query += "AND C.pioneer = 1 "
            case "explorer":
                query += "AND C.explorer = 1 "
            case "modern":
                query += "AND C.modern = 1 "
            case "legacy":
                query += "AND C.legacy = 1 "
            case "pauper":
                query += "AND C.pauper = 1 "
            case "vintage":
                query += "AND C.vintage = 1 "
            case "penny":
                query += "AND C.penny = 1 "
            case "commander":
                query += "AND C.commander = 1 "
            case "oathbreaker":
                query += "AND C.oathbreaker = 1 "
            case "brawl":
                query += "AND C.brawl = 1 "
            case "historicbrawl":
                query += "AND C.historicbrawl = 1 "
            case "alchemy":
                query += "AND C.alchemy = 1 "
            case "paupercommander":
                query += "AND C.paupercommander = 1 "
            case "duel":
                query += "AND C.duel = 1 "
            case "oldschool":
                query += "AND C.oldschool = 1 "
            case "premodern":
                query += "AND C.premodern = 1 "
            case "predh":
                query += "AND C.predh = 1 "
        }
    }

    sortColumn := "C.name"
    switch sort {
        case "name":
            sortColumn = "C.name"
        case "tmc":
            sortColumn = "C.totalManaCost DESC"
        case "power":
            sortColumn = "C.power DESC"
        case "toughness":
            sortColumn = "C.toughness DESC"
    }
    query += "GROUP BY C.id, C.name ORDER BY " + sortColumn + " LIMIT @limit OFFSET @offset"
    db.Raw(query, params).Scan(&cards)
    return cards
}

func GetCardTypes(db *gorm.DB) []string {
    var types []string
    db.Raw("SELECT DISTINCT type FROM Cards ORDER BY type").Scan(&types)
    return types
}

func SearchCardTypes(db *gorm.DB, name string) []string {
    name = "%" + strings.Trim(name, " ") + "%"
    var types []string
    db.Raw("SELECT DISTINCT type FROM Cards WHERE type LIKE ? ORDER BY type", name).Scan(&types)
    return types
}

func GetCardSubtypes(db *gorm.DB) []string {
    var subtypes []string
    db.Raw("SELECT DISTINCT subtype FROM Card_Subtypes ORDER BY subtype").Scan(&subtypes)
    return subtypes
}

func SearchCardSubtypes(db *gorm.DB, name string) []string {
    name = "%" + strings.Trim(name, " ") + "%"
    var subtypes []string
    db.Raw("SELECT DISTINCT subtype FROM Card_Subtypes WHERE subtype LIKE ? ORDER BY subtype", name).Scan(&subtypes)
    return subtypes
}

func GetCardKeywords(db *gorm.DB) []string {
    var keywords []string
    db.Raw("SELECT DISTINCT keyword FROM Card_Keywords ORDER BY keyword").Scan(&keywords)
    return keywords
}

func SearchCardKeywords(db *gorm.DB, name string) []string {
    name = "%" + strings.Trim(name, " ") + "%"
    var keywords []string
    db.Raw("SELECT DISTINCT keyword FROM Card_Keywords WHERE keyword LIKE ? ORDER BY keyword", name).Scan(&keywords)
    return keywords
}

func GetDeckCard(db *gorm.DB, activeDeckId string, cardId string) DeckCard {
    deckCard := DeckCard{}
    db.Raw("SELECT DC.print, DC.sideboard, HEX(DC.deck_id) AS deck_id, HEX(DC.card_id) AS card_id, HEX(DC.id) AS id, DC.qty, C.name, C.front, C.art FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) AND DC.card_id = UNHEX(?) LIMIT 1", activeDeckId, cardId).Scan(&deckCard)
    return deckCard
}
func GetDeckCardById(db *gorm.DB, activeDeckId string, deckCardId string) DeckCard {
    deckCard := DeckCard{}
    db.Raw("SELECT DC.print, DC.sideboard, HEX(DC.deck_id) AS deck_id, HEX(DC.card_id) AS card_id, HEX(DC.id) AS id, DC.qty, C.name, C.front, C.art FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) AND DC.id = UNHEX(?) LIMIT 1", activeDeckId, deckCardId).Scan(&deckCard)
    return deckCard
}

func GetCard(db *gorm.DB, cardId string) Card {
    card := Card{}
    db.Raw("SELECT C.name, HEX(C.id) AS id, C.front, C.back, C.type, C.toughness, C.power, C.totalManaCost, C.art, C.standard, C.future, C.historic, C.gladiator, C.pioneer, C.explorer, C.modern, C.legacy, C.pauper, C.vintage, C.penny, C.commander, C.oathbreaker, C.brawl, C.historicbrawl, C.alchemy, C.paupercommander, C.duel, C.oldschool, C.premodern, C.predh, C.rarity, C.manaCost FROM Cards C WHERE C.id = UNHEX(?) LIMIT 1", cardId).Scan(&card)
    return card
}
