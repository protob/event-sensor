-- name: ListEvents :many
SELECT * FROM events ORDER BY start_date ASC;

-- name: ListEventsByDateRange :many
SELECT * FROM events WHERE start_date >= ? AND start_date <= ? ORDER BY start_date ASC;

-- name: GetEvent :one
SELECT * FROM events WHERE id = ?;

-- name: GetEventBySource :one
SELECT * FROM events WHERE source = ? AND source_id = ?;

-- name: ListEventsByArtist :many
-- Events with a performance by this tracked artist. GROUP BY: an artist may have >1 performance
-- at one event (festival two-sets), so de-dup to one event row.
SELECT e.* FROM events e
JOIN performances p ON e.id = p.event_id
WHERE p.artist_id = ?
GROUP BY e.id
ORDER BY e.start_date ASC;

-- name: SearchEvents :many
-- Search event name, venue name, or ANY performer name (entity or name-only).
SELECT e.* FROM events e
LEFT JOIN venues v ON e.venue_id = v.id
LEFT JOIN performances p ON e.id = p.event_id
WHERE e.name LIKE ? OR v.name LIKE ? OR p.artist_name LIKE ?
GROUP BY e.id
ORDER BY e.start_date ASC;

-- name: InsertEvent :one
INSERT INTO events (id, source, source_id, name, kind, description, about,
                    start_date, end_date, has_time, venue_id, url, image_url,
                    ticket_options, listing_state)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- Upsert a TM event by its natural key. Conflict target matches the partial unique index.
-- name: UpsertTmEvent :one
INSERT INTO events (id, source, source_id, name, kind, description, about,
                    start_date, end_date, has_time, venue_id, url, image_url,
                    ticket_options, listing_state)
VALUES (?, 'ticketmaster', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'listed')
ON CONFLICT(source, source_id) WHERE source_id IS NOT NULL DO UPDATE SET
    name = excluded.name, kind = excluded.kind,
    description = excluded.description, about = excluded.about,
    start_date = excluded.start_date, end_date = excluded.end_date, has_time = excluded.has_time,
    venue_id = excluded.venue_id, url = excluded.url, image_url = excluded.image_url,
    ticket_options = excluded.ticket_options,
    listing_state = 'listed',                 -- TM has it => re-list it
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: SetListingState :exec
UPDATE events SET listing_state = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateEvent :one
-- Edits a (manual) event's mutable facts in place; identity + source unchanged.
UPDATE events SET
    name = ?, kind = ?, description = ?, about = ?,
    start_date = ?, end_date = ?, has_time = ?, venue_id = ?, url = ?, image_url = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: AddPerformance :exec
-- One artist at an event. artist_id NULL = name-only. No unique key: an artist may
-- appear twice at one festival (different date/venue).
INSERT INTO performances (id, event_id, artist_id, artist_name, is_headliner, date, venue_name)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListPerformancesByEvent :many
SELECT id, event_id, artist_id, artist_name, is_headliner, date, venue_name
FROM performances WHERE event_id = ?
ORDER BY is_headliner DESC, date, artist_name;

-- name: ListAllPerformances :many
SELECT id, event_id, artist_id, artist_name, is_headliner, date, venue_name
FROM performances ORDER BY is_headliner DESC;

-- name: DeletePerformancesByEvent :exec
DELETE FROM performances WHERE event_id = ?;

-- name: CountArtistSourceEvents :one
SELECT COUNT(DISTINCT e.id) AS count FROM events e
JOIN performances p ON e.id = p.event_id
WHERE p.artist_id = ? AND e.source = ?;

-- name: DeleteEvent :exec
-- Guarded by ON DELETE RESTRICT: errors if a library_entries row references it.
DELETE FROM events WHERE id = ?;

-- name: DeleteAllUnclaimedEvents :exec
-- "Clear all events" = drop everything no one has claimed (claimed survive via RESTRICT predicate).
DELETE FROM events
WHERE NOT EXISTS (SELECT 1 FROM library_entries le WHERE le.event_id = events.id);

-- name: DeletePastUnclaimedTmEvents :exec
-- On-demand purge of past Ticketmaster events no one claimed. Claimed events are safe.
DELETE FROM events
WHERE source = 'ticketmaster'
  AND date(start_date) < date('now')
  AND NOT EXISTS (SELECT 1 FROM library_entries le WHERE le.event_id = events.id);

-- name: DeleteUnclaimedArtistEvents :exec
-- Clear one artist's un-claimed TM events (claim-safe; claimed survive).
DELETE FROM events
WHERE source = 'ticketmaster'
  AND id IN (SELECT event_id FROM performances WHERE artist_id = ?)
  AND NOT EXISTS (SELECT 1 FROM library_entries le WHERE le.event_id = events.id);
