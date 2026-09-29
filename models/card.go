package models

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
    Price int `gorm:"column:price"`
    FmtPrice string
    Text string
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
    Gamemode string
    DateCreated string `gorm:"column:dateCreated;type:datetime"`
    IsCommander bool
    IsPartner bool
    IsOathbreaker bool
    InSideboard bool `gorm:"column:sideboard;type:tinyint"`
    Print string `gorm:"column:print"`
    Price int `gorm:"column:price"`
    FmtPrice string
    IsGuest bool
    IsLegal bool
    LegalStandard bool `gorm:"column:standard;type:tinyint"`
    LegalFuture bool `gorm:"column:future;type:tinyint"`
    LegalHistoric bool `gorm:"column:historic;type:tinyint"`
    LegalGladiator bool `gorm:"column:gladiator;type:tinyint"`
    LegalPioneer bool `gorm:"column:pioneer;type:tinyint"`
    LegalExplorer bool `gorm:"column:explorer;type:tinyint"`
    LegalModern bool `gorm:"column:modern;type:tinyint"`
    LegalLegacy bool `gorm:"column:legacy;type:tinyint"`
    LegalPauper bool `gorm:"column:pauper;type:tinyint"`
    LegalVintage bool `gorm:"column:vintage;type:tinyint"`
    LegalPenny bool `gorm:"column:penny;type:tinyint"`
    LegalCommander bool `gorm:"column:commander;type:tinyint"`
    LegalOathbreaker bool `gorm:"column:oathbreaker;type:tinyint"`
    LegalBrawl bool `gorm:"column:brawl;type:tinyint"`
    LegalHistoricBrawl bool `gorm:"column:historicbrawl;type:tinyint"`
    LegalAlchemy bool `gorm:"column:alchemy;type:tinyint"`
    LegalPauperCommander bool `gorm:"column:paupercommander;type:tinyint"`
    LegalDuel bool `gorm:"column:duel;type:tinyint"`
    LegalOldSchool bool `gorm:"column:oldschool;type:tinyint"`
    LegalPremodern bool `gorm:"column:premodern;type:tinyint"`
    LegalPredh bool `gorm:"column:predh;type:tinyint"`
}

type DeckCardMetadata struct {
    CardId string `gorm:"column:card_id"`
    Qty    uint8    `gorm:"column:qty"`
    Front string 
    Back string
    Name string
    IsCommander bool
    IsPartner bool
    IsOathbreaker bool
    InSideboard bool `gorm:"column:sideboard;type:tinyint"`
    Print string `gorm:"column:print"`
}

// CardPrint is one distinct look a card can be shown as, not one printing.
// Several printings that render identically collapse into a single CardPrint.
type CardPrint struct {
    CardId string
    // Print is the front_hash, which is what a player's choice is recorded as.
    Print string `gorm:"column:print"`
    DeckId string
    Front string
    Back string
}

// A player's choice lives in Deck_Cards.print as a front_hash; with no choice
// the card's own default look is used. Resolving that in SQL keeps every caller
// from repeating the fallback, and COALESCE also keeps a NULL out of a Go
// string, which the driver will not scan.
const (
    chosenFrontHash = "COALESCE(LOWER(HEX(DC.print)), C.front, '')"
    rawChosenPrint  = "COALESCE(LOWER(HEX(DC.print)), '')"
    // A look determines its own reverse face, so every Card_Prints row sharing
    // a front_hash carries the same back_hash and LIMIT 1 is deterministic
    // rather than an arbitrary pick between differing values.
    chosenBackHash = "COALESCE((SELECT LOWER(HEX(cp.back_hash)) FROM Card_Prints cp " +
        "WHERE cp.card_id = DC.card_id AND cp.front_hash = DC.print AND cp.back_hash IS NOT NULL LIMIT 1), C.back, '')"
)

func SearchCardsByName(db *gorm.DB, name string, offset int, limit int) []Card {
    name = "%" + strings.Trim(name, " ") + "%"
    var cards []Card
    db.Raw("SELECT C.front, HEX(C.id) AS id, C.name FROM Cards AS C WHERE C.name LIKE ? LIMIT ? OFFSET ?", name, limit, offset).Scan(&cards)
    return cards
}

func GetDeckCards (db *gorm.DB, deckId string) []DeckCard {
    var cards []DeckCard
    db.Raw("SELECT C.standard, C.future, C.historic, C.gladiator, C.pioneer, C.explorer, C.modern, C.legacy, C.pauper, C.vintage, C.penny, C.commander, C.oathbreaker, C.brawl, C.historicbrawl, C.alchemy, C.paupercommander, C.duel, C.oldschool, C.premodern, C.predh, C.price, " + rawChosenPrint + " AS print, DC.sideboard, DC.dateCreated, C.art, " + chosenFrontHash + " AS front, " + chosenBackHash + " AS back, HEX(DC.id) AS id, HEX(DC.card_id) AS card_id, (SELECT c.name FROM Card_Names c WHERE C.id = c.card_id LIMIT 1) AS name, DC.qty FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) ORDER BY dateCreated DESC", deckId).Scan(&cards)
    return cards
}

func GetDeckCardsMetadata (db *gorm.DB, deckId string) []DeckCardMetadata {
    var cards []DeckCardMetadata
    db.Raw("SELECT " + rawChosenPrint + " AS print, DC.sideboard, HEX(DC.card_id) AS card_id, " + chosenFrontHash + " AS front, " + chosenBackHash + " AS back, (SELECT c.name FROM Card_Names c WHERE C.id = c.card_id LIMIT 1) AS name, DC.qty FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) ORDER BY dateCreated DESC", deckId).Scan(&cards)
    return cards
}

func SearchDeckCards(db *gorm.DB, deckId string, name string, sort string, filter string, rarity string, color string) []DeckCard {
    query := "SELECT C.price, " + rawChosenPrint + " AS print, DC.sideboard, HEX(DC.deck_id) AS deck_id, HEX(C.id) as card_id, DC.dateCreated, C.art, " + chosenFrontHash + " AS front, " + chosenBackHash + " AS back, HEX(DC.id) AS id, DC.qty, C.name, "
    query += "C.standard, C.future, C.historic, C.gladiator, C.pioneer, C.explorer, C.modern, C.legacy, C.pauper, C.vintage, C.penny, C.commander, C.oathbreaker, C.brawl, C.historicbrawl, C.alchemy, C.paupercommander, C.duel, C.oldschool, C.premodern, C.predh "
    query += "FROM Deck_Cards DC JOIN Cards C ON C.id = DC.card_id "
    params := map[string]interface{}{
        "deck": deckId,
        "name": "%" + strings.Trim(name, " ") + "%",
    }
    
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
        case "price":
            sortColumn = "C.price DESC"
    }
    filterLogic := "AND 1=1"
    switch filter {
        case "creatures":
            filterLogic = "AND C.type LIKE '%Creature%'"
        case "enchantments":
            filterLogic = "AND C.type LIKE '%Enchantment%'"
        case "artifacts":
            filterLogic = "AND C.type LIKE '%Artifact%'"
        case "lands":
            filterLogic = "AND C.type LIKE '%Land%'"
        case "instants":
            filterLogic = "AND C.type LIKE '%Instant%'"
        case "sorceries":
            filterLogic = "AND C.type LIKE '%Sorcery%'"
    }

    switch color {
        case "W":
            query += "JOIN Card_Colors cc ON C.id = cc.card_id JOIN Colors c ON cc.color_id = c.id AND c.color = 'W' "
        case "U":
            query += "JOIN Card_Colors cc ON C.id = cc.card_id JOIN Colors c ON cc.color_id = c.id AND c.color = 'U' "
        case "B":
            query += "JOIN Card_Colors cc ON C.id = cc.card_id JOIN Colors c ON cc.color_id = c.id AND c.color = 'B' "
        case "R":
            query += "JOIN Card_Colors cc ON C.id = cc.card_id JOIN Colors c ON cc.color_id = c.id AND c.color = 'R' "
        case "G":
            query += "JOIN Card_Colors cc ON C.id = cc.card_id JOIN Colors c ON cc.color_id = c.id AND c.color = 'G' "
    }

    if rarity != "any" && rarity != "" {
        params["rarity"] = rarity
        query += "JOIN Rarities R ON C.rarity = R.id AND R.rarity = @rarity "
    }

    var cards []DeckCard
    db.Raw(query + "WHERE DC.deck_id = UNHEX(@deck) AND C.name LIKE @name " + filterLogic + " GROUP BY DC.id ORDER BY " + sortColumn, params).Scan(&cards)
    return cards
}

func FilterCards(db *gorm.DB, name string, searchText bool, sort string, mana []string, types []string, subtypes []string, keywords []string, rarity string, legality string, set string, price int, offset int, limit int) []Card {
    var cards []Card
    query := "SELECT GROUP_CONCAT(CT.text) as text, C.name, C.price, C.front, C.back, HEX(C.id) AS id, C.name FROM Cards AS C LEFT JOIN Card_Texts CT ON C.id = CT.card_id "
    name = strings.Trim(name, " ")

    manaCheck := []string{}
    params := map[string]interface{}{
        "limit": limit,
        "offset": offset,
    }
    includeColorless := false
    includeColor := false
    if len(mana) > 0 {
        for i := 0; i < len(mana); i++ {
            if mana[i] == "C" {
                includeColorless = true
            } else {
                includeColor = true
                manaCheck = append(manaCheck, "@mana" + fmt.Sprint(i))
                params["mana" + fmt.Sprint(i)] = mana[i]
            }
        }
    }
    colorLogic := ""
    if includeColor && !includeColorless {
        colorLogic += "ExcludeColors.card_id IS NULL "
        query += "LEFT JOIN ( SELECT cc.card_id FROM Card_Colors cc JOIN Colors c ON cc.color_id = c.id WHERE c.color <> " + strings.Join(manaCheck, " AND c.color <> ") + ") AS ExcludeColors ON C.id = ExcludeColors.card_id INNER JOIN Card_Colors cc ON C.id = cc.card_id "
    } else if includeColorless && !includeColor {
        query += "LEFT JOIN Card_Colors cc ON C.id = cc.card_id "
        colorLogic += "cc.card_id IS NULL "
    } else if includeColorless && includeColor {
        colorLogic += "ExcludeColors.card_id IS NULL "
        query += "LEFT JOIN ( SELECT cc.card_id FROM Card_Colors cc JOIN Colors c ON cc.color_id = c.id WHERE c.color <> " + strings.Join(manaCheck, " AND c.color <> ") + ") AS ExcludeColors ON C.id = ExcludeColors.card_id LEFT JOIN Card_Colors cc ON C.id = cc.card_id "
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
        if searchText {
            query += "JOIN (SELECT DISTINCT card_id FROM Card_Texts CT WHERE MATCH(`text`) AGAINST (@fts IN NATURAL LANGUAGE MODE)) AS ftx_card ON ftx_card.card_id = C.id ";
            params["fts"] = name
            // Without this, the filter clauses below append to the join's ON
            // instead of a WHERE. Equivalent for an inner join, but only by
            // accident.
            query += "WHERE 1=1 "
        } else {
            query += "WHERE C.name LIKE @name "
            params["name"] = "%" + strings.Trim(name, " ") + "%"
        }
    } else {
        query += "WHERE 1=1 "
    }
    
    if colorLogic != "" {
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

    if set != "" {
        query += "AND C.set_name = @set "
        params["set"] = set
    }

    if price > 0 {
        query += "AND C.price <= @price "
        params["price"] = price
    }

    sortColumn := "C.name"
    switch sort {
        case "name":
            sortColumn = "C.name"
        case "tmc":
            sortColumn = "C.totalManaCost DESC"
        case "lmc":
            sortColumn = "C.totalManaCost ASC"
        case "power":
            sortColumn = "C.power DESC"
        case "toughness":
            sortColumn = "C.toughness DESC"
        case "priceHL":
            sortColumn = "C.price DESC"
        case "priceLH":
            sortColumn = "C.price ASC"
        case "edhRank":
            sortColumn = "C.edh_rank"
    }
    query += "GROUP BY C.name, C.front, C.back, C.id ORDER BY " + sortColumn + " LIMIT @limit OFFSET @offset"
    db.Raw(query, params).Scan(&cards)
    return cards
}

func GetCardTypes(db *gorm.DB) []string {
    return typesList.get(db, "SELECT DISTINCT type FROM Cards ORDER BY type")
}

// SearchCardTypes filters the cached type list in memory. The query it replaced
// ran LIKE '%name%' over all of Cards on every keystroke, which cannot use an
// index. There are only a few hundred distinct types, so the whole list is
// cached and filtered here instead.
func SearchCardTypes(db *gorm.DB, name string) []string {
    return filterContains(GetCardTypes(db), name)
}

func GetCardSubtypes(db *gorm.DB) []string {
    return subtypesList.get(db, "SELECT DISTINCT subtype FROM Card_Subtypes ORDER BY subtype")
}

func SearchCardSubtypes(db *gorm.DB, name string) []string {
    return filterContains(GetCardSubtypes(db), name)
}

func GetCardKeywords(db *gorm.DB) []string {
    return keywordsList.get(db, "SELECT DISTINCT keyword FROM Card_Keywords ORDER BY keyword")
}

func SearchCardKeywords(db *gorm.DB, name string) []string {
    return filterContains(GetCardKeywords(db), name)
}

// filterContains mirrors MySQL's LIKE '%name%', which is case insensitive under
// the default collation.
func filterContains(values []string, name string) []string {
    needle := strings.ToLower(strings.Trim(name, " "))
    matches := []string{}
    for _, value := range values {
        if strings.Contains(strings.ToLower(value), needle) {
            matches = append(matches, value)
        }
    }
    return matches
}

func GetDeckCard(db *gorm.DB, activeDeckId string, cardId string) DeckCard {
    deckCard := DeckCard{}
    db.Raw("SELECT " + rawChosenPrint + " AS print, DC.sideboard, HEX(DC.deck_id) AS deck_id, HEX(DC.card_id) AS card_id, HEX(DC.id) AS id, DC.qty, C.name, " + chosenFrontHash + " AS front, C.art FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) AND DC.card_id = UNHEX(?) LIMIT 1", activeDeckId, cardId).Scan(&deckCard)
    return deckCard
}
func GetDeckCardById(db *gorm.DB, activeDeckId string, deckCardId string) DeckCard {
    deckCard := DeckCard{}
    db.Raw("SELECT " + rawChosenPrint + " AS print, DC.sideboard, HEX(DC.deck_id) AS deck_id, HEX(DC.card_id) AS card_id, HEX(DC.id) AS id, DC.qty, C.name, " + chosenFrontHash + " AS front, C.art FROM Deck_Cards DC JOIN Cards C ON DC.card_id = C.id WHERE DC.deck_id = UNHEX(?) AND DC.id = UNHEX(?) LIMIT 1", activeDeckId, deckCardId).Scan(&deckCard)
    return deckCard
}

func GetCard(db *gorm.DB, cardId string) Card {
    card := Card{}
    db.Raw("SELECT C.name, HEX(C.id) AS id, C.front, C.back, C.type, C.toughness, C.power, C.totalManaCost, C.art, C.standard, C.future, C.historic, C.gladiator, C.pioneer, C.explorer, C.modern, C.legacy, C.pauper, C.vintage, C.penny, C.commander, C.oathbreaker, C.brawl, C.historicbrawl, C.alchemy, C.paupercommander, C.duel, C.oldschool, C.premodern, C.predh, C.rarity, C.manaCost FROM Cards C WHERE C.id = UNHEX(?) LIMIT 1", cardId).Scan(&card)
    return card
}

// GetPrints lists the distinct looks a card can be shown as, oldest first.
//
// Grouping is on the hash rather than on the release date, which is what makes
// this correct: several printings share one look when they reuse an artwork and
// treatment, and one date covers several looks when a set ships alternate art.
// front_hash IS NOT NULL skips printings the processor has not reached yet, so
// rows still carrying the old scheme are simply absent rather than broken.
func GetPrints(db *gorm.DB, cardId string) []CardPrint {
    prints := []CardPrint{}
    db.Raw(`SELECT LOWER(HEX(CP.front_hash)) AS print,
                   COALESCE(LOWER(HEX(CP.back_hash)), '') AS back
            FROM Card_Prints CP
            WHERE CP.card_id = UNHEX(?) AND CP.front_hash IS NOT NULL
            GROUP BY CP.front_hash, CP.back_hash
            ORDER BY MIN(CP.released)`, cardId).Scan(&prints)
    return prints
}

// GetPrintLook reports the reverse face of a look and whether that look exists
// for the card at all.
//
// UNHEX of a value that is not hex is NULL, which matches no row, so this also
// rejects a malformed hash rather than letting it reach Deck_Cards.print.
func GetPrintLook(db *gorm.DB, cardId string, frontHash string) (string, bool) {
    var found []string
    db.Raw(`SELECT COALESCE(LOWER(HEX(CP.back_hash)), '') AS back
            FROM Card_Prints CP
            WHERE CP.card_id = UNHEX(?) AND CP.front_hash = UNHEX(?)
            LIMIT 1`, cardId, frontHash).Scan(&found)
    if len(found) == 0 {
        return "", false
    }
    return found[0], true
}

// referenceTTL is how long a reference list is reused. These only change when
// the card database is reimported.
const referenceTTL = time.Hour

var (
    setsList     cachedList
    typesList    cachedList
    subtypesList cachedList
    keywordsList cachedList
)

// cachedList holds a DISTINCT lookup that is identical for every user. Each of
// these was a full scan plus a sort, run per request.
type cachedList struct {
    mu      sync.RWMutex
    values  []string
    fetched time.Time
}

func (c *cachedList) get(db *gorm.DB, query string) []string {
    c.mu.RLock()
    if c.values != nil && time.Since(c.fetched) < referenceTTL {
        cached := c.values
        c.mu.RUnlock()
        return cached
    }
    c.mu.RUnlock()

    values := []string{}
    db.Raw(query).Scan(&values)
    if len(values) == 0 {
        // A failed query scans into an empty slice; don't cache that.
        return values
    }

    c.mu.Lock()
    c.values = values
    c.fetched = time.Now()
    c.mu.Unlock()

    return values
}

func GetSets(db *gorm.DB) []string {
    return setsList.get(db, "SELECT DISTINCT set_name FROM Cards ORDER BY set_name")
}

// CardMatch is a card resolved from a name in an imported deck list.
type CardMatch struct {
    Id        string `gorm:"column:id"`
    Name      string `gorm:"column:name"`
    Front     string `gorm:"column:front"`
    MatchName string `gorm:"column:match_name"`
}

// FindCardsByNames resolves a batch of imported names in one round trip. An
// import looks up a hundred or so names at once, which is a hundred queries if
// each is resolved on its own.
//
// Names are matched against Card_Names as well as Cards.name so that the faces
// of split, modal and room cards resolve. The result is keyed by the lower
// cased name that matched, which is what the caller has in hand.
func FindCardsByNames(db *gorm.DB, names []string) map[string]CardMatch {
    matches := map[string]CardMatch{}
    if len(names) == 0 {
        return matches
    }

    placeholders := strings.TrimSuffix(strings.Repeat("?,", len(names)), ",")
    args := []interface{}{}
    for i := 0; i < 2; i++ {
        for _, name := range names {
            args = append(args, name)
        }
    }

    rows := []CardMatch{}
    db.Raw(
        "SELECT HEX(C.id) AS id, C.name, C.front, LOWER(C.name) AS match_name FROM Cards C WHERE C.name IN ("+placeholders+") "+
            "UNION "+
            "SELECT HEX(C.id) AS id, IFNULL(C.name, N.name) AS name, C.front, LOWER(N.name) AS match_name FROM Card_Names N JOIN Cards C ON C.id = N.card_id WHERE N.name IN ("+placeholders+")",
        args...,
    ).Scan(&rows)

    // A name can resolve to more than one row when a card has several printings
    // in Cards. Keep the first so an import is repeatable; the print can be
    // changed per card from the deck overview afterwards.
    for _, row := range rows {
        if _, seen := matches[row.MatchName]; !seen {
            matches[row.MatchName] = row
        }
    }

    return matches
}

// FilterExistingCardIds returns the subset of ids that exist. The import review
// form carries card ids in hidden inputs, so they are the client's word until
// they are checked against the database.
func FilterExistingCardIds(db *gorm.DB, ids []string) map[string]bool {
    existing := map[string]bool{}
    if len(ids) == 0 {
        return existing
    }

    placeholders := strings.TrimSuffix(strings.Repeat("UNHEX(?),", len(ids)), ",")
    args := []interface{}{}
    for _, id := range ids {
        args = append(args, id)
    }

    found := []string{}
    db.Raw("SELECT HEX(id) FROM Cards WHERE id IN ("+placeholders+")", args...).Scan(&found)

    for _, id := range found {
        existing[strings.ToUpper(id)] = true
    }

    return existing
}
