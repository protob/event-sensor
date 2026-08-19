-- name: ListUserLibrary :many
SELECT * FROM library_entries WHERE user_id = ?;

-- name: GetLibraryEntry :one
SELECT * FROM library_entries WHERE user_id = ? AND event_id = ?;

-- name: UpsertLibraryEntry :one
INSERT INTO library_entries (user_id, event_id, status, missed_reason, note, attended_date)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id, event_id) DO UPDATE SET
    status = excluded.status,
    missed_reason = excluded.missed_reason,
    note = excluded.note,
    attended_date = excluded.attended_date,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteLibraryEntry :exec
DELETE FROM library_entries WHERE user_id = ? AND event_id = ?;

-- name: CountUserByStatus :one
SELECT COUNT(*) AS count FROM library_entries WHERE user_id = ? AND status = ?;

-- name: ListUserLibraryByStatus :many
SELECT le.*, e.start_date, e.name, e.listing_state
FROM library_entries le JOIN events e ON e.id = le.event_id
WHERE le.user_id = ? AND le.status = ?
ORDER BY e.start_date DESC;

-- "needs resolution": going + the date has passed (derived; for the attendance nudge).
-- name: ListNeedsResolution :many
SELECT le.*, e.name, e.start_date
FROM library_entries le JOIN events e ON e.id = le.event_id
WHERE le.user_id = ? AND le.status = 'going'
  AND date(e.start_date) < date('now')
ORDER BY e.start_date ASC;
