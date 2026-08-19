-- name: ListCategories :many
SELECT * FROM categories WHERE user_id = ? ORDER BY name;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = ?;

-- name: CreateCategory :one
INSERT INTO categories (id, name, user_id) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateCategory :one
UPDATE categories SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;

-- name: DeleteArtistCategoriesByCategory :exec
DELETE FROM artist_categories WHERE category_id = ?;
