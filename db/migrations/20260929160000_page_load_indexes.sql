-- Indexes for the deck builder, chosen from EXPLAIN against production-sized data.
--
-- Deck_Cards carried nothing but its primary key, so the card-count subquery in
-- GetDecks scanned all of it once per deck in the nav: 41 scans of 10,428 rows
-- on a single page load. The two indexes below turn that into a covering lookup,
-- and let the deck tray read in dateCreated order rather than sorting.
--
-- The Cards indexes only pay off alongside the group key in FilterCards naming
-- the sort column. While the key led with name, a sort on price or mana value
-- had to materialise every card first, and no index on those columns could be
-- reached.

-- migrate:up transaction:false
CREATE INDEX idx_deck_cards_deck_sideboard ON Deck_Cards (deck_id, sideboard, qty);
CREATE INDEX idx_deck_cards_deck_created ON Deck_Cards (deck_id, dateCreated);

-- These scale with the number of users rather than the size of one deck, so they
-- are cheap now and awkward to add later.
CREATE INDEX idx_decks_user ON Decks (user_id);
CREATE INDEX idx_deck_groups_user ON Deck_Groups (user_id);

CREATE INDEX idx_cards_price ON Cards (price);
CREATE INDEX idx_cards_tmc ON Cards (totalManaCost);
CREATE INDEX idx_cards_power ON Cards (power);
CREATE INDEX idx_cards_toughness ON Cards (toughness);
-- name trails set_name so one index serves the filter and the order together.
CREATE INDEX idx_cards_set_name ON Cards (set_name, name);

-- All three lead with id, which is unique, so there is never a second column to
-- seek or order by: they duplicate the primary key and can serve no query it
-- cannot. Dropping them returns index space on Cards and spares the importer
-- three B-tree writes for every row it touches.
DROP INDEX idx_cards_id ON Cards;
DROP INDEX idx_cards_id_power ON Cards;
DROP INDEX idx_cards_id_toughness ON Cards;

-- migrate:down
CREATE INDEX idx_cards_id ON Cards (id);
CREATE INDEX idx_cards_id_power ON Cards (id, power);
CREATE INDEX idx_cards_id_toughness ON Cards (id, toughness);
DROP INDEX idx_cards_set_name ON Cards;
DROP INDEX idx_cards_toughness ON Cards;
DROP INDEX idx_cards_power ON Cards;
DROP INDEX idx_cards_tmc ON Cards;
DROP INDEX idx_cards_price ON Cards;
DROP INDEX idx_deck_groups_user ON Deck_Groups;
DROP INDEX idx_decks_user ON Decks;
DROP INDEX idx_deck_cards_deck_created ON Deck_Cards;
DROP INDEX idx_deck_cards_deck_sideboard ON Deck_Cards;
