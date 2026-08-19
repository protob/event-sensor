-- name: ListArtists :many
SELECT id, name, fetch_mode, created_at, updated_at FROM artists ORDER BY name;

-- name: GetArtist :one
SELECT * FROM artists WHERE id = ?;

-- name: CreateArtist :one
INSERT INTO artists (id, name, fetch_mode) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateArtist :one
UPDATE artists SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING *;

-- name: SetArtistFetchMode :one
UPDATE artists SET fetch_mode = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING *;

-- name: DeleteArtist :exec
DELETE FROM artists WHERE id = ?;

-- name: ListArtistsByCategory :many
SELECT a.id, a.name, a.fetch_mode, a.created_at, a.updated_at
FROM artists a JOIN artist_categories ac ON a.id = ac.artist_id
WHERE ac.category_id = ? ORDER BY a.name;

-- name: AddArtistToCategory :exec
INSERT OR IGNORE INTO artist_categories (artist_id, category_id) VALUES (?, ?);

-- name: RemoveArtistFromCategory :exec
DELETE FROM artist_categories WHERE artist_id = ? AND category_id = ?;

-- name: GetArtistCategoryCount :one
SELECT COUNT(*) as count FROM artist_categories WHERE artist_id = ?;

-- name: DeleteArtistCategoriesByArtist :exec
DELETE FROM artist_categories WHERE artist_id = ?;

-- name: ListArtistCategoryPairs :many
-- Every artist-to-category membership for one user, in one round trip. Replaces the
-- per-category fetch the category tree fires on mount.
SELECT ac.artist_id, c.id AS category_id, c.name AS category_name
FROM artist_categories ac
JOIN categories c ON c.id = ac.category_id
WHERE c.user_id = ?
ORDER BY c.name;

-- name: ListArtistEventCounts :many
-- Per-artist totals. upcoming_listed is what the tree's "dormant" marker reads.
SELECT p.artist_id,
       COUNT(DISTINCT e.id) AS event_count,
       COUNT(DISTINCT CASE
           WHEN e.listing_state = 'listed' AND date(e.start_date) >= date('now')
           THEN e.id END) AS upcoming_listed_count
FROM performances p
JOIN events e ON e.id = p.event_id
WHERE p.artist_id IS NOT NULL
GROUP BY p.artist_id;

-- name: ListArtistCountryCounts :many
-- Countries an artist's upcoming listed events are in - the flag row on the artist page and
-- in the category tree.
SELECT p.artist_id, v.country_code, COUNT(DISTINCT e.id) AS n
FROM performances p
JOIN events e ON e.id = p.event_id
JOIN venues v ON v.id = e.venue_id
WHERE p.artist_id IS NOT NULL
  AND v.country_code IS NOT NULL
  AND e.listing_state = 'listed'
  AND date(e.start_date) >= date('now')
GROUP BY p.artist_id, v.country_code;

-- name: ListArtistClaimedCounts :many
-- How many of an artist's events this user has claimed (any status).
SELECT p.artist_id, COUNT(DISTINCT e.id) AS claimed_count
FROM performances p
JOIN events e ON e.id = p.event_id
JOIN library_entries le ON le.event_id = e.id
WHERE p.artist_id IS NOT NULL AND le.user_id = ?
GROUP BY p.artist_id;
