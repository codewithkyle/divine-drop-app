export interface Card {
    name:          string;
    layout:        string;
    colors:        string[];
    legalities:    Legalities;
    rarity:        string;
    keywords:      any[];
    front:         string;
    back:          null;
    type:          string;
    subtypes:      string[];
    texts:         string[];
    manaCosts:     string[];
    totalManaCost: number;
    faceNames:     string[];
    flavorTexts:   string[];
    toughness:     string;
    power:         string;
    id:            string;
}

export interface Legalities {
    standard:        string;
    future:          string;
    historic:        string;
    gladiator:       string;
    pioneer:         string;
    explorer:        string;
    modern:          string;
    legacy:          string;
    pauper:          string;
    vintage:         string;
    penny:           string;
    commander:       string;
    brawl:           string;
    historicbrawl:   string;
    alchemy:         string;
    paupercommander: string;
    duel:            string;
    oldschool:       string;
    premodern:       string;
}
