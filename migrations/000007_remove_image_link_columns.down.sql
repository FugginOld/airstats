ALTER TABLE interesting_aircraft
    ADD COLUMN IF NOT EXISTS link TEXT,
    ADD COLUMN IF NOT EXISTS image_link_1 TEXT,
    ADD COLUMN IF NOT EXISTS image_link_2 TEXT,
    ADD COLUMN IF NOT EXISTS image_link_3 TEXT,
    ADD COLUMN IF NOT EXISTS image_link_4 TEXT;

ALTER TABLE interesting_aircraft_seen
    ADD COLUMN IF NOT EXISTS link TEXT,
    ADD COLUMN IF NOT EXISTS image_link_1 TEXT,
    ADD COLUMN IF NOT EXISTS image_link_2 TEXT,
    ADD COLUMN IF NOT EXISTS image_link_3 TEXT,
    ADD COLUMN IF NOT EXISTS image_link_4 TEXT;
