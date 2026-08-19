-- name: ListVenues :many
SELECT * FROM venues ORDER BY name;

-- name: GetVenue :one
SELECT * FROM venues WHERE id = ?;

-- name: GetVenueBySource :one
SELECT * FROM venues WHERE source = ? AND source_id = ?;

-- name: UpsertVenue :one
INSERT INTO venues (id, name, city, country, country_code, latitude, longitude, timezone, source, source_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(source, source_id) WHERE source_id IS NOT NULL DO UPDATE SET
    name = excluded.name, city = excluded.city, country = excluded.country,
    country_code = excluded.country_code, latitude = excluded.latitude,
    longitude = excluded.longitude, timezone = excluded.timezone, updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteVenue :exec
DELETE FROM venues WHERE id = ?;

-- name: GetManualVenue :one
-- Manual venues carry no source_id, so venues_source_key cannot dedupe them; match on the
-- fields the user actually typed. COLLATE NOCASE is ASCII-only - "Cafe" and "cafe" match,
-- accented spellings do not.
SELECT * FROM venues
WHERE source = 'manual'
  AND source_id IS NULL
  AND name = sqlc.arg(name) COLLATE NOCASE
  AND COALESCE(city, '') = sqlc.arg(city) COLLATE NOCASE
LIMIT 1;

-- name: ListVenuesWithCounts :many
SELECT v.*, COUNT(e.id) AS event_count
FROM venues v
LEFT JOIN events e ON e.venue_id = v.id
GROUP BY v.id
ORDER BY v.name;

-- name: CountVenueEvents :one
SELECT COUNT(*) AS count FROM events WHERE venue_id = ?;
