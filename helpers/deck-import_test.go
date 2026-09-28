package helpers

import "testing"

func TestParseDeckListArchidektDecoration(t *testing.T) {
    list := ParseDeckList(`1x Academy Ruins (drc) 58 [Land]
1x Boros Charm (fdn) 721 *F* [Protection]
1x Shadowspear (plst) THB-236 [Lifegain]
1x Mondrak, Glory Dominus (pone) 23p [Tokens]
1x Cathars' Crusade (inr) 17 [Counters] ^Getting,#2ccce4^
1x Ghyrson Starn, Kelermorph (40k) 124 [Burn,Creature] ^Proxy,#0500a9^
1x Kellan, the Fae-Blooded // Birthright Boon (woe) 230 [Tutor]
6x Forest (dft) 289 [Land]`, false)

    if len(list.BadLines) != 0 {
        t.Fatalf("expected every line to parse, got bad lines: %v", list.BadLines)
    }

    want := []string{
        "Academy Ruins",
        "Boros Charm",
        "Shadowspear",
        "Mondrak, Glory Dominus",
        "Cathars' Crusade",
        "Ghyrson Starn, Kelermorph",
        "Kellan, the Fae-Blooded // Birthright Boon",
        "Forest",
    }
    if len(list.Cards) != len(want) {
        t.Fatalf("expected %d cards, got %d", len(want), len(list.Cards))
    }
    for i, name := range want {
        if list.Cards[i].Name != name {
            t.Errorf("card %d: expected name %q, got %q", i, name, list.Cards[i].Name)
        }
    }

    if qty := list.Cards[len(list.Cards)-1].Qty; qty != 6 {
        t.Errorf("expected 6x Forest to parse as qty 6, got %d", qty)
    }
}

func TestParseDeckListReservedCategories(t *testing.T) {
    list := ParseDeckList(`1x Saheeli, Radiant Creator (drc) 3 [Commander{top}]
1x Soul's Attendant (plst) ROE-44 [Lifegain,Sideboard]
1x Spectator Seating (cmm) 427 [Maybeboard{noDeck}{noPrice}]
1x Sol Ring (drc) 57 [Artifact]`, false)

    if len(list.Cards) != 3 {
        t.Fatalf("expected 3 imported cards, got %d", len(list.Cards))
    }
    if len(list.Skipped) != 1 || list.Skipped[0] != "Spectator Seating" {
        t.Fatalf("expected the maybeboard card to be skipped, got %v", list.Skipped)
    }
    if !list.Cards[0].IsCommander {
        t.Error("expected [Commander{top}] to mark a commander")
    }
    if !list.Cards[1].InSideboard {
        t.Error("expected Sideboard among several categories to mark a sideboard card")
    }
    if list.Cards[2].IsCommander || list.Cards[2].InSideboard {
        t.Error("expected a plain category to mark neither")
    }
}

// A user can flag any category of their own making as not part of the deck, so
// exclusion is read from the modifier rather than from the category's name.
func TestParseDeckListUserExcludedCategory(t *testing.T) {
    list := ParseDeckList("1x Sol Ring (drc) 57 [Considering{noDeck}]", false)

    if len(list.Cards) != 0 {
        t.Fatalf("expected the card to be skipped, got %d cards", len(list.Cards))
    }
    if len(list.Skipped) != 1 {
        t.Fatalf("expected 1 skipped card, got %v", list.Skipped)
    }
}

// Moxfield writes an unlabelled trailing block. What it holds depends on the
// format the user picked, so the caller decides.
func TestParseDeckListTrailingBlock(t *testing.T) {
    raw := "1 Arcane Signet\n9 Forest\n\n1 Zimone, Mystery Unraveler"

    commander := ParseDeckList(raw, true)
    if len(commander.Cards) != 3 {
        t.Fatalf("expected 3 cards, got %d", len(commander.Cards))
    }
    if !commander.Cards[2].IsCommander {
        t.Error("expected the trailing block to be read as the commander")
    }

    sideboard := ParseDeckList(raw, false)
    if !sideboard.Cards[2].InSideboard {
        t.Error("expected the trailing block to be read as the sideboard")
    }
    if sideboard.Cards[2].IsCommander {
        t.Error("did not expect a commander in a non commander format")
    }
}

func TestParseDeckListBadLines(t *testing.T) {
    list := ParseDeckList("Deck\n1 Sol Ring\nnot a card line", false)

    if len(list.Cards) != 1 {
        t.Fatalf("expected 1 card, got %d", len(list.Cards))
    }
    if len(list.BadLines) != 2 {
        t.Fatalf("expected 2 unparsed lines, got %v", list.BadLines)
    }
}

func TestCardNameLookups(t *testing.T) {
    cases := map[string][]string{
        "Sol Ring":                      {"Sol Ring"},
        "Experimental Lab/Staff Room":   {"Experimental Lab/Staff Room", "Experimental Lab // Staff Room", "Experimental Lab"},
        "Glasspool Mimic // Glasspool Shore": {"Glasspool Mimic // Glasspool Shore", "Glasspool Mimic"},
        "Bontu’s Monument":         {"Bontu's Monument"},
    }

    for name, want := range cases {
        got := CardNameLookups(name)
        if len(got) != len(want) {
            t.Errorf("%q: expected %v, got %v", name, want, got)
            continue
        }
        for i := range want {
            if got[i] != want[i] {
                t.Errorf("%q: expected %v, got %v", name, want, got)
                break
            }
        }
    }
}
