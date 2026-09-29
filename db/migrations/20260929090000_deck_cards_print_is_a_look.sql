-- Deck_Cards.print becomes the look a player chose.
--
-- It held a release date, which never identified a look: one date can cover
-- several alternate arts, and one look can span several dates. A front_hash does
-- identify one, and unlike a print id it survives a Scryfall refresh that
-- renumbers printings.
--
-- Translating a date into a hash has three outcomes and they are not the same
-- kind of thing, so they are not treated alike:
--
--   one art at that date     resolve it; the choice is unambiguous
--   several arts at that date leave NULL; the date never recorded which one, and
--                            picking by print id shows a player art they did not
--                            choose. Falling back to the default printing is a
--                            smaller wrong than quietly swapping the artwork.
--   no printing at that date leave NULL; upstream dropped it -- a prerelease
--                            promo folded into its main set, say -- and no
--                            amount of waiting brings it back.
--
-- The dates move to print_released_legacy instead of being dropped, so the rows
-- left NULL stay recoverable. A later sweep can revisit them as the data
-- improves rather than this migration deciding them permanently.

-- migrate:up transaction:false

-- The one outcome worth refusing to boot over: a choice whose printing still
-- exists but has no hash yet. That is a half finished processor run, and
-- resolving the rows it reached while emptying the rest is silent loss of what
-- players picked -- recoverable by waiting, so waiting is what should happen.
-- Dates with no printing at all are unrecoverable by anyone and fall through to
-- the backfill; conflating the two is what made this abort on every boot.
--
-- Left behind if it raises, since transaction:false commits it; the DROP ... IF
-- EXISTS makes the next attempt's CREATE succeed regardless.
DROP PROCEDURE IF EXISTS dd_assert_choices_resolvable;
CREATE PROCEDURE dd_assert_choices_resolvable()
BEGIN
    DECLARE pending INT;
    SELECT COUNT(*) INTO pending
    FROM Deck_Cards dc
    WHERE dc.print IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM Card_Prints cp
          WHERE cp.card_id = dc.card_id AND cp.released = dc.print
            AND cp.front_hash IS NOT NULL)
      AND EXISTS (
          SELECT 1 FROM Card_Prints cp
          WHERE cp.card_id = dc.card_id AND cp.released = dc.print);
    IF pending > 0 THEN
        -- MESSAGE_TEXT truncates at 128 bytes and raises 1648 rather than
        -- shortening, which would replace this explanation with a MySQL error
        -- about its own error. Keep the text well inside that.
        SET @dd_msg = CONCAT(pending, ' recorded print choices have no hash yet; let the card processor finish, or prune the printings it replaced');
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = @dd_msg;
    END IF;
END;
CALL dd_assert_choices_resolvable();
DROP PROCEDURE dd_assert_choices_resolvable;

ALTER TABLE Deck_Cards ADD COLUMN print_hash binary(16) DEFAULT NULL;

-- MIN over a set already proven to hold one distinct value is that value. The
-- HAVING discards the group where the date covers several arts, and an empty
-- input produces a COUNT of zero that it discards too, so both land on NULL
-- through the same branch. There is no tiebreak because there is no honest way
-- to break the tie.
UPDATE Deck_Cards dc SET print_hash = (
    SELECT MIN(cp.front_hash) FROM Card_Prints cp
    WHERE cp.card_id = dc.card_id AND cp.released = dc.print
      AND cp.front_hash IS NOT NULL
    HAVING COUNT(DISTINCT cp.front_hash) = 1
) WHERE dc.print IS NOT NULL;

ALTER TABLE Deck_Cards CHANGE COLUMN print print_released_legacy int DEFAULT NULL;
ALTER TABLE Deck_Cards CHANGE COLUMN print_hash print binary(16) DEFAULT NULL;

-- migrate:down
-- A real rollback, unlike the shape-only one this replaced: the dates are still
-- in print_released_legacy, so putting the column back brings the choices with
-- it. Anything resolved to a hash is rebuilt from its date on the way up again.
ALTER TABLE Deck_Cards DROP COLUMN print;
ALTER TABLE Deck_Cards CHANGE COLUMN print_released_legacy print int DEFAULT NULL;
