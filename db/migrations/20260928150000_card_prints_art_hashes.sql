-- Gives a printing an identity based on the artwork it actually shows.
--
-- Until now a printing was addressed by its release date, which is neither
-- unique (four Zendikar Forests share one) nor stable across printings that
-- reuse an illustration (a promo and its main set release the same art a week
-- apart). Both the CDN filename and Deck_Cards.print were built from that date,
-- so the first case collapsed distinct art onto one file and the second showed
-- the same art twice.
--
-- These hold the content hash of the rendered card image, written by the card
-- processor as it uploads. Both are nullable: the processor fills them over a
-- long run while the application keeps serving from released, and nothing reads
-- them until the cutover. A NULL front_hash is simply a row not yet processed,
-- which is also how a resumed run finds its remaining work.
--
-- Card_Prints.id is not referenced anywhere, so the processor is free to write
-- the upstream print id into it rather than the random UUID used today. That is
-- the identity this table always needed, and it makes a re-run idempotent
-- instead of duplicating every row.

-- migrate:up transaction:false
ALTER TABLE Card_Prints
    ADD COLUMN front_hash binary(16) DEFAULT NULL,
    ADD COLUMN back_hash binary(16) DEFAULT NULL,
    ADD KEY idx_card_prints_card_id_front_hash (card_id, front_hash);

-- migrate:down
ALTER TABLE Card_Prints
    DROP KEY idx_card_prints_card_id_front_hash,
    DROP COLUMN back_hash,
    DROP COLUMN front_hash;
