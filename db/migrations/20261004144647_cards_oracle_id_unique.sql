-- The one column the card importer looks every card up by.
--
-- import.js resolves a card's existing row by oracle_id before it writes
-- anything, so that a re-import updates the row a player's decks already point
-- at instead of creating a second card. Cards carried no index on that column,
-- so each of those lookups was a full scan of the table. Measured against a
-- production sized copy, 35,968 cards: 5.3ms a lookup before this index, 0.006ms
-- after, and EXPLAIN goes from type: ALL to type: const. The importer does one
-- per card in the manifest, so that is three minutes of server CPU a run, for
-- a query that should not touch disk at all.
--
-- Nothing in this application reads oracle_id, which is why the column went
-- unnoticed while the page-load indexes were being chosen: no page was ever
-- slow because of it.
--
-- UNIQUE because that is already what the importer assumes. It takes the first
-- row the lookup returns and updates it, so a second row carrying the same
-- oracle id would be a card that silently never gets refreshed again. There are
-- none today - 35,968 cards, 35,968 distinct oracle ids, none NULL - and the
-- constraint is what stops one arriving.
--
-- The algorithm and lock are named, unlike the page-load indexes, because this
-- one is expected to run with the site up and an import waiting behind it. Both
-- are what MySQL 8 picks anyway for a secondary index, and the build took 225ms
-- on that copy; naming them means the server refuses outright rather than
-- quietly holding the table if that ever stops being true. CREATE INDEX takes
-- them space separated - the comma that ALTER TABLE wants is a syntax error
-- here.

-- migrate:up transaction:false
CREATE UNIQUE INDEX idx_cards_oracle_id ON Cards (oracle_id) ALGORITHM=INPLACE LOCK=NONE;

-- migrate:down
DROP INDEX idx_cards_oracle_id ON Cards;
