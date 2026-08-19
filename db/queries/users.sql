-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CreateUser :one
INSERT INTO users (id, username, email, password, role) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateUserProfile :one
UPDATE users SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;
