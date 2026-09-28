package helpers

import (
    "regexp"
    "strconv"
    "strings"
)

// ImportedCard is one card read out of an exported deck list.
type ImportedCard struct {
    Qty         int
    Name        string
    IsCommander bool
    InSideboard bool
}

// ParsedDeckList is everything a deck list export told us. Skipped and bad
// lines are kept so the review screen can account for every line in the file
// instead of silently dropping some.
type ParsedDeckList struct {
    Cards    []ImportedCard
    Skipped  []string
    BadLines []string
}

var (
    // Archidekt appends user tags to a line as ^Label,#hexcolour^.
    importTagSuffix = regexp.MustCompile(`\s*\^[^^]*\^\s*$`)

    // Archidekt writes categories as [Name{mod}{mod},Other]. Categories are
    // user defined, so only the reserved names carry meaning.
    importCategorySuffix = regexp.MustCompile(`\s*\[([^\]]*)\]\s*$`)

    // Finish markers, e.g. "1x Boros Charm (fdn) 721 *F* [Protection]".
    importFinishSuffix = regexp.MustCompile(`\s*\*[A-Za-z]+\*\s*$`)

    // A printing: "(drc) 58", "(plst) THB-236", "(pone) 23p". Collector
    // numbers are not always numeric, so the trailing token is matched loosely.
    importPrintSuffix = regexp.MustCompile(`\s*\([0-9A-Za-z]{2,10}\)\s+\S+\s*$`)

    // "1 Sol Ring" (Moxfield) and "1x Sol Ring" (Archidekt).
    importQtyPrefix = regexp.MustCompile(`^(\d+)\s*[xX]?\s+`)

    importModifier = regexp.MustCompile(`\{[^}]*\}`)
)

// ParseDeckList reads a Moxfield or Archidekt text export.
//
// Both sites write one card per line as "<qty> <name>", so one parser covers
// them. Archidekt decorates the line with a printing, a finish, user categories
// and user tags. Those are stripped from the right, because a card name can
// contain commas, apostrophes and " // ", which leaves no reliable delimiter to
// split on from the left.
//
// Moxfield does not label its sections: it separates them with a blank line and
// nothing else. commanderBlock decides what a trailing block means, and the
// caller picks that from the chosen gamemode - a commander format means the
// block holds the commander, anything else means it holds the sideboard.
func ParseDeckList(raw string, commanderBlock bool) ParsedDeckList {
    parsed := ParsedDeckList{
        Cards:    []ImportedCard{},
        Skipped:  []string{},
        BadLines: []string{},
    }

    blocks := splitDeckListBlocks(raw)

    for i, block := range blocks {
        isTrailing := len(blocks) > 1 && i == len(blocks)-1

        for _, line := range block {
            card, excluded, ok := parseDeckListLine(line)
            if !ok {
                parsed.BadLines = append(parsed.BadLines, line)
                continue
            }
            if excluded {
                parsed.Skipped = append(parsed.Skipped, card.Name)
                continue
            }

            // An explicit category always wins over the block's position.
            if isTrailing && !card.IsCommander && !card.InSideboard {
                if commanderBlock {
                    card.IsCommander = true
                } else {
                    card.InSideboard = true
                }
            }

            parsed.Cards = append(parsed.Cards, card)
        }
    }

    return parsed
}

// splitDeckListBlocks groups lines into blank line separated blocks, dropping
// the blank lines themselves.
func splitDeckListBlocks(raw string) [][]string {
    blocks := [][]string{}
    block := []string{}

    for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
        line = strings.TrimSpace(line)
        if line == "" {
            if len(block) > 0 {
                blocks = append(blocks, block)
                block = []string{}
            }
            continue
        }
        block = append(block, line)
    }

    if len(block) > 0 {
        blocks = append(blocks, block)
    }

    return blocks
}

// parseDeckListLine reads one line. The second return reports a card the export
// marked as not part of the deck, the third whether the line was a card at all.
func parseDeckListLine(line string) (ImportedCard, bool, bool) {
    card := ImportedCard{Qty: 1}
    excluded := false

    for importTagSuffix.MatchString(line) {
        line = importTagSuffix.ReplaceAllString(line, "")
    }

    if match := importCategorySuffix.FindStringSubmatch(line); match != nil {
        commander, sideboard, notInDeck := parseImportCategories(match[1])
        card.IsCommander = commander
        card.InSideboard = sideboard
        excluded = notInDeck
        line = importCategorySuffix.ReplaceAllString(line, "")
    }

    for importFinishSuffix.MatchString(line) {
        line = importFinishSuffix.ReplaceAllString(line, "")
    }

    line = importPrintSuffix.ReplaceAllString(line, "")

    match := importQtyPrefix.FindStringSubmatch(line)
    if match == nil {
        return card, excluded, false
    }

    qty, err := strconv.Atoi(match[1])
    if err != nil || qty < 1 {
        return card, excluded, false
    }

    card.Qty = qty
    card.Name = strings.TrimSpace(importQtyPrefix.ReplaceAllString(line, ""))

    return card, excluded, card.Name != ""
}

// parseImportCategories reads an Archidekt category list.
//
// Users name and rename categories freely, so a name is only meaningful for the
// three reserved ones. Exclusion is read from the {noDeck} modifier rather than
// from the word "Maybeboard", because a user can set that flag on any category
// of their own making.
func parseImportCategories(raw string) (bool, bool, bool) {
    commander := false
    sideboard := false
    excluded := false

    for _, entry := range strings.Split(raw, ",") {
        entry = strings.TrimSpace(entry)
        if entry == "" {
            continue
        }

        for _, modifier := range importModifier.FindAllString(entry, -1) {
            if strings.EqualFold(strings.Trim(modifier, "{}"), "noDeck") {
                excluded = true
            }
        }

        switch name := strings.TrimSpace(importModifier.ReplaceAllString(entry, "")); {
        case strings.EqualFold(name, "Commander"):
            commander = true
        case strings.EqualFold(name, "Sideboard"):
            sideboard = true
        case strings.EqualFold(name, "Maybeboard"):
            excluded = true
        }
    }

    return commander, sideboard, excluded
}

// NormalizeCardName makes an exported name comparable to the names held in the
// card database. Exports differ in which apostrophe and which ligature they
// use, and a name pasted out of a browser often carries the typographic ones.
func NormalizeCardName(name string) string {
    name = importNameReplacer.Replace(name)
    return strings.Join(strings.Fields(name), " ")
}

var importNameReplacer = strings.NewReplacer(
    "‘", "'",
    "’", "'",
    "“", `"`,
    "”", `"`,
    "Æ", "Ae",
    "æ", "ae",
)

// CardNameLookups returns the names to try for one imported card, most specific
// first. Multi faced cards are written several ways - Moxfield exports a room
// as "Experimental Lab/Staff Room" while Archidekt exports a modal land as
// "Glasspool Mimic // Glasspool Shore" - so the spaced spelling and the front
// face on its own are both worth a try when the combined name is not stored.
func CardNameLookups(name string) []string {
    name = NormalizeCardName(name)
    if name == "" {
        return []string{}
    }

    lookups := []string{name}

    separator := ""
    if strings.Contains(name, "//") {
        separator = "//"
    } else if strings.Contains(name, "/") {
        separator = "/"
    }
    if separator == "" {
        return lookups
    }

    faces := []string{}
    for _, face := range strings.Split(name, separator) {
        if face = strings.TrimSpace(face); face != "" {
            faces = append(faces, face)
        }
    }
    if len(faces) == 0 {
        return lookups
    }

    for _, candidate := range []string{strings.Join(faces, " // "), faces[0]} {
        if candidate != name {
            lookups = append(lookups, candidate)
        }
    }

    return lookups
}
