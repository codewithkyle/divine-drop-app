package models

import "gorm.io/gorm"

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
}

func SearchCardsByName(db *gorm.DB, name string, offset int, limit int) []Card {
    var cards []Card
    db.Raw("SELECT C.front FROM Cards AS C JOIN Card_Names AS CN ON C.id = CN.card_id WHERE CN.name LIKE ? LIMIT ? OFFSET ?", name, limit, offset).Scan(&cards)
    return cards
}
