-- Sleeves.image_url becomes a key rather than an absolute URL.
--
-- It stored a full origin, so CDN_URL moved every other asset while these rows
-- kept pointing at wherever they were uploaded. Holding the key instead makes
-- CDN_URL the one authority for where assets are served from, and lets the
-- delete path address exactly the object the upload wrote.
--
-- Everything from "/users/" onward is kept rather than replacing a known origin,
-- so a row written against any past host converts correctly. Rows that are
-- already keys have no scheme and are left alone, which makes this safe to run
-- against a database that has been partly converted by hand.

-- migrate:up transaction:false
UPDATE Sleeves
SET image_url = CONCAT('users/', SUBSTRING_INDEX(image_url, '/users/', -1))
WHERE image_url LIKE 'http%/users/%';

-- migrate:down
-- Not reversible: the origin each row was written against is not recorded, and
-- assuming one would invent history. Point CDN_URL at the old host instead.
