ALTER TABLE interesting_aircraft
    DROP COLUMN IF EXISTS link,
    DROP COLUMN IF EXISTS image_link_1,
    DROP COLUMN IF EXISTS image_link_2,
    DROP COLUMN IF EXISTS image_link_3,
    DROP COLUMN IF EXISTS image_link_4;

ALTER TABLE interesting_aircraft_seen
    DROP COLUMN IF EXISTS link,
    DROP COLUMN IF EXISTS image_link_1,
    DROP COLUMN IF EXISTS image_link_2,
    DROP COLUMN IF EXISTS image_link_3,
    DROP COLUMN IF EXISTS image_link_4;
