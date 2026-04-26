-- Initial schema for Airstats database

-- Create aircraft_data table
CREATE TABLE aircraft_data (
    id SERIAL PRIMARY KEY,
    hex VARCHAR,
    flight VARCHAR,
    first_seen TIMESTAMPTZ,
    first_seen_epoch BIGINT,
    last_seen TIMESTAMPTZ,
    last_seen_epoch BIGINT,
    type VARCHAR,
    r VARCHAR,
    t VARCHAR,
    alt_baro INTEGER,
    alt_geom INTEGER,
    gs NUMERIC(6,1),
    ias INTEGER,
    tas INTEGER,
    mach NUMERIC(5,3),
    track NUMERIC(5,2),
    track_rate NUMERIC(5,2),
    roll NUMERIC(5,2),
    mag_heading NUMERIC(5,2),
    true_heading NUMERIC(5,2),
    baro_rate INTEGER,
    geom_rate INTEGER,
    squawk VARCHAR,
    emergency VARCHAR,
    nav_qnh NUMERIC(7,1),
    nav_altitude_mcp INTEGER,
    nav_heading NUMERIC(5,2),
    nav_modes TEXT[],
    lat NUMERIC(9,6),
    lon NUMERIC(9,6),
    nic INTEGER,
    rc INTEGER,
    seen_pos NUMERIC(9,3),
    r_dst NUMERIC(8,3),
    r_dir NUMERIC(8,3),
    version INTEGER,
    nic_baro INTEGER,
    nac_p INTEGER,
    nac_v INTEGER,
    sil INTEGER,
    sil_type VARCHAR,
    gva INTEGER,
    sda INTEGER,
    alert INTEGER,
    spi INTEGER,
    mlat TEXT[],
    tisb TEXT[],
    messages INTEGER,
    seen NUMERIC(8,3),
    rssi NUMERIC(6,1),
    highest_aircraft_processed BOOLEAN DEFAULT false,
    lowest_aircraft_processed BOOLEAN DEFAULT false,
    fastest_aircraft_processed BOOLEAN DEFAULT false,
    slowest_aircraft_processed BOOLEAN DEFAULT false,
    db_flags INTEGER,
    route_processed BOOLEAN DEFAULT false,
    registration_processed BOOLEAN DEFAULT false,
    interesting_processed BOOLEAN DEFAULT false,
    last_seen_lat NUMERIC(9,6),
    last_seen_lon NUMERIC(9,6),
    last_seen_distance NUMERIC(6,2),
    last_seen_bearing NUMERIC(6,3),
    destination_distance NUMERIC(8,2)
);

-- Create registration_data table
CREATE TABLE registration_data (
    id SERIAL PRIMARY KEY,
    type VARCHAR,
    icao_type VARCHAR,
    manufacturer VARCHAR,
    mode_s VARCHAR,
    registration VARCHAR,
    registered_owner_country_iso_name VARCHAR,
    registered_owner_country_name VARCHAR,
    registered_owner_operator_flag_code VARCHAR,
    registered_owner VARCHAR,
    url_photo VARCHAR,
    url_photo_thumbnail VARCHAR,
    CONSTRAINT mode_s_unique UNIQUE (mode_s)
);

-- Create fastest_aircraft table
CREATE TABLE fastest_aircraft (
    id SERIAL PRIMARY KEY,
    hex VARCHAR,
    flight VARCHAR,
    registration VARCHAR,
    type VARCHAR,
    first_seen TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,
    ground_speed NUMERIC(6,1),
    indicated_air_speed INTEGER,
    true_air_speed INTEGER,
    CONSTRAINT fastest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen)
);

-- Create route_data table
CREATE TABLE route_data (
    id SERIAL PRIMARY KEY,
    route_callsign VARCHAR,
    route_callsign_icao VARCHAR,
    route_callsign_iata VARCHAR,
    airline_name VARCHAR,
    airline_icao VARCHAR,
    airline_iata VARCHAR,
    airline_country VARCHAR,
    airline_country_iso VARCHAR,
    airline_callsign VARCHAR,
    origin_country_iso_name VARCHAR,
    origin_country_name VARCHAR,
    origin_elevation INTEGER,
    origin_iata_code VARCHAR,
    origin_icao_code VARCHAR,
    origin_latitude NUMERIC(9,6),
    origin_longitude NUMERIC(9,6),
    origin_municipality VARCHAR,
    origin_name VARCHAR,
    destination_country_iso_name VARCHAR,
    destination_country_name VARCHAR,
    destination_elevation INTEGER,
    destination_iata_code VARCHAR,
    destination_icao_code VARCHAR,
    destination_latitude NUMERIC(9,6),
    destination_longitude NUMERIC(9,6),
    destination_municipality VARCHAR,
    destination_name VARCHAR,
    last_updated TIMESTAMP,
    route_distance NUMERIC(8,2),
    CONSTRAINT route_callsign_unique UNIQUE (route_callsign)
);

-- Create highest_aircraft table
CREATE TABLE highest_aircraft (
    id SERIAL PRIMARY KEY,
    hex VARCHAR,
    flight VARCHAR,
    registration VARCHAR,
    type VARCHAR,
    first_seen TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,
    barometric_altitude INTEGER,
    geometric_altitude INTEGER,
    CONSTRAINT highest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen)
);

-- Create interesting_aircraft table
CREATE TABLE interesting_aircraft (
    icao TEXT,
    registration TEXT,
    operator TEXT,
    type TEXT,
    icao_type TEXT,
    "group" TEXT,
    tag1 TEXT,
    tag2 TEXT,
    tag3 TEXT,
    category TEXT,
    link TEXT,
    image_link_1 TEXT,
    image_link_2 TEXT,
    image_link_3 TEXT,
    image_link_4 TEXT
);

-- Create interesting_aircraft_seen table
CREATE TABLE interesting_aircraft_seen (
    icao TEXT,
    registration TEXT,
    operator TEXT,
    type TEXT,
    icao_type TEXT,
    "group" TEXT,
    tag1 TEXT,
    tag2 TEXT,
    tag3 TEXT,
    category TEXT,
    link TEXT,
    image_link_1 TEXT,
    image_link_2 TEXT,
    image_link_3 TEXT,
    image_link_4 TEXT,
    hex TEXT,
    flight TEXT,
    seen TIMESTAMPTZ,
    seen_epoch BIGINT,
    r TEXT,
    t TEXT,
    alt_baro INTEGER,
    alt_geom INTEGER,
    gs NUMERIC(6,1),
    ias INTEGER,
    tas INTEGER,
    track NUMERIC(5,2),
    baro_rate INTEGER,
    squawk TEXT,
    emergency TEXT,
    lat NUMERIC(9,6),
    lon NUMERIC(9,6),
    alert INTEGER,
    db_flags INTEGER
);

-- Create lowest_aircraft table
CREATE TABLE lowest_aircraft (
    id SERIAL PRIMARY KEY,
    hex VARCHAR,
    flight VARCHAR,
    registration VARCHAR,
    type VARCHAR,
    first_seen TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,
    barometric_altitude INTEGER,
    geometric_altitude INTEGER,
    CONSTRAINT lowest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen)
);

-- Create slowest_aircraft table
CREATE TABLE slowest_aircraft (
    id SERIAL PRIMARY KEY,
    hex VARCHAR,
    flight VARCHAR,
    registration VARCHAR,
    type VARCHAR,
    first_seen TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,
    ground_speed NUMERIC(6,1),
    indicated_air_speed INTEGER,
    true_air_speed INTEGER,
    CONSTRAINT slowest_aircraft_unique_hex_first_seen UNIQUE (hex, first_seen)
);

-- Create indexes
CREATE INDEX aircraft_data_hex ON aircraft_data USING btree (hex) WITH (deduplicate_items='true');
CREATE INDEX idx_aircraft_data_hex ON aircraft_data USING btree (hex);
CREATE INDEX idx_aircraft_data_hex_last_seen ON aircraft_data USING btree (hex, last_seen DESC);