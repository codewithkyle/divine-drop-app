-- Deck_Cards.print becomes the look a player chose.
--
-- It held a release date, which never identified a look: one date can cover
-- several alternate arts, and one look can span several dates. A front_hash does
-- identify one, and unlike a print id it survives a Scryfall refresh that
-- renumbers printings.
--
-- This translates existing choices through Card_Prints.front_hash, so it depends
-- on the card processor having populated that column. Run first, it would turn
-- every recorded choice into NULL without complaining, so it refuses instead:
-- boot fails loudly rather than quietly discarding what players picked.

-- migrate:up transaction:false
DROP PROCEDURE IF EXISTS dd_assert_choices_resolvable;
CREATE PROCEDURE dd_assert_choices_resolvable()
BEGIN
    DECLARE unresolved INT;
    -- Exactly the condition the backfill below relies on. Checking that some
    -- hashes exist is not enough: the processor leaves untouched rows NULL until
    -- --prune, so a half finished run would resolve the choices it reached and
    -- silently empty the rest.
    SELECT COUNT(*) INTO unresolved
    FROM Deck_Cards dc
    WHERE dc.print IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM Card_Prints cp
          WHERE cp.card_id = dc.card_id AND cp.released = dc.print
            AND cp.front_hash IS NOT NULL);
    IF unresolved > 0 THEN
        SET @dd_msg = CONCAT(unresolved, ' recorded print choices have no hashed printing yet; let the card processor finish');
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = @dd_msg;
    END IF;
END;
CALL dd_assert_choices_resolvable();
DROP PROCEDURE dd_assert_choices_resolvable;

ALTER TABLE Deck_Cards ADD COLUMN print_hash binary(16) DEFAULT NULL;

-- Where a date covers several looks the choice is genuinely ambiguous, so order
-- by the print id to land on the same row every time rather than whichever the
-- optimiser happens to reach first.
UPDATE Deck_Cards dc SET print_hash = (
    SELECT cp.front_hash FROM Card_Prints cp
    WHERE cp.card_id = dc.card_id AND cp.released = dc.print
      AND cp.front_hash IS NOT NULL
    ORDER BY cp.id LIMIT 1
) WHERE dc.print IS NOT NULL;

ALTER TABLE Deck_Cards DROP COLUMN print;
ALTER TABLE Deck_Cards CHANGE COLUMN print_hash print binary(16) DEFAULT NULL;

-- migrate:down
-- Restores the shape, not the contents: the release dates are gone once the
-- column is dropped, so a rollback leaves every choice empty.
ALTER TABLE Deck_Cards CHANGE COLUMN print print_hash binary(16) DEFAULT NULL;
ALTER TABLE Deck_Cards ADD COLUMN print int DEFAULT NULL;
ALTER TABLE Deck_Cards DROP COLUMN print_hash;
