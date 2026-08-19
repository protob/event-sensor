-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE artists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    -- auto   = fetched by "Fetch All" (default).
    -- manual = excluded from "Fetch All"; only fetched when you click that artist's fetch button
    --          (disbanded / deceased / that-lineup-is-done artists — you keep them, just don't bulk-fetch).
    fetch_mode TEXT NOT NULL DEFAULT 'auto' CHECK (fetch_mode IN ('auto','manual')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE artist_categories (
    artist_id   TEXT NOT NULL REFERENCES artists(id)    ON DELETE CASCADE,
    category_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (artist_id, category_id)
) STRICT;
CREATE INDEX idx_artist_categories_artist   ON artist_categories(artist_id);
CREATE INDEX idx_artist_categories_category ON artist_categories(category_id);

CREATE TABLE venues (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT, country TEXT, country_code TEXT,
    latitude REAL, longitude REAL,
    timezone TEXT,                      -- IANA tz; for correct local-time rendering (nullable for now)
    source TEXT,                        -- 'ticketmaster' | 'manual'
    source_id TEXT,                     -- TM venue id; NULL for manual
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;
CREATE UNIQUE INDEX venues_source_key ON venues(source, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX idx_venues_country ON venues(country_code);

CREATE TABLE events (
    id TEXT PRIMARY KEY,                -- app-assigned UUID for EVERY event
    source TEXT NOT NULL,              -- 'ticketmaster' | 'manual'
    source_id TEXT,                    -- TM event id; NULL for manual

    name TEXT NOT NULL,
    -- A FESTIVAL *IS* an event (kind='festival') — the thing you attend & claim. Its artists are
    -- rows in `performances`. A normal gig is kind='concert' with its performances. There is NO
    -- separate festival entity and NO event->festival link: a festival cannot belong to itself.
    kind TEXT NOT NULL DEFAULT 'concert'
        CHECK (kind IN ('concert','festival','tribute','club','other')),
    description TEXT,
    about TEXT,

    start_date TEXT NOT NULL,          -- ISO-8601 UTC; or 'YYYY-MM-DD' if time TBA. Festival: first day.
    end_date TEXT,                     -- festival: last day (multi-day)
    has_time INTEGER NOT NULL DEFAULT 1,   -- 0 => start_date is date-only

    venue_id TEXT REFERENCES venues(id) ON DELETE SET NULL,   -- single-site festival = its venue; urban = NULL
    url TEXT, image_url TEXT,
    ticket_options TEXT NOT NULL DEFAULT '[]'                 -- JSON [{label,url}]; '[]' default (was '') so json_valid holds
        CHECK (json_valid(ticket_options)),

    listing_state TEXT NOT NULL DEFAULT 'listed'
        CHECK (listing_state IN ('listed','delisted','cancelled')),

    extra TEXT NOT NULL DEFAULT '{}'                          -- prototyping bag: source-specific extras, future fields
        CHECK (json_valid(extra)),
    discovered_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,    -- first-surfaced; set on INSERT, never updated

    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;
CREATE UNIQUE INDEX events_source_key ON events(source, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX idx_events_start_date ON events(start_date);
CREATE INDEX idx_events_listing    ON events(listing_state);

-- Who played at an event (concert OR festival).
--   artist_id SET   => an entity artist (in your catalog)
--   artist_id NULL  => a name-only artist (no catalog entry)
-- No (event_id, artist_id) unique key: an artist may appear twice at one festival (different
-- `date`). No stage / no position — that's noise for a diary.
CREATE TABLE performances (
    id           TEXT PRIMARY KEY,
    event_id     TEXT NOT NULL REFERENCES events(id)  ON DELETE CASCADE,
    artist_id    TEXT          REFERENCES artists(id) ON DELETE SET NULL,
    artist_name  TEXT NOT NULL,                          -- always present (display)
    is_headliner INTEGER NOT NULL DEFAULT 0,             -- concert: who I came for; ignored for festivals
    date         TEXT,                                   -- 'YYYY-MM-DD' which day at the festival; NULL = the event date
    venue_name   TEXT,                                   -- per-artist venue label (urban festivals: Unsound's halls); NULL = the event venue
    created_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;
CREATE INDEX idx_performances_event  ON performances(event_id);
CREATE INDEX idx_performances_artist ON performances(artist_id);

-- THE HEART: durable per-user claim. Existence => event is persistent. RESTRICT, never CASCADE.
CREATE TABLE library_entries (
    user_id       TEXT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    event_id      TEXT NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    status        TEXT NOT NULL CHECK (status IN ('interested','going','attended','missed')),
    missed_reason TEXT CHECK (missed_reason IS NULL OR status = 'missed'),
    note          TEXT,
    attended_date TEXT,
    created_at    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, event_id)
) STRICT;
CREATE INDEX idx_library_event       ON library_entries(event_id);   -- powers NOT EXISTS anti-join
CREATE INDEX idx_library_user_status ON library_entries(user_id, status);

CREATE TABLE fetch_log (
    id TEXT PRIMARY KEY,
    artist_id TEXT REFERENCES artists(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    status TEXT NOT NULL,           -- 'success'|'error'|'empty'|'partial'
    error_message TEXT,
    events_fetched INTEGER, events_kept INTEGER,
    events_inserted INTEGER, events_updated INTEGER,
    events_deleted INTEGER, events_delisted INTEGER
) STRICT;
CREATE INDEX idx_fetch_log_artist ON fetch_log(artist_id, started_at);

-- +goose Down
DROP TABLE IF EXISTS fetch_log;
DROP TABLE IF EXISTS library_entries;
DROP TABLE IF EXISTS performances;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS venues;
DROP TABLE IF EXISTS artist_categories;
DROP TABLE IF EXISTS artists;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
