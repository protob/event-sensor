-- name: GetUserSettings :many
SELECT id, user_id, key, value, created_at, updated_at
FROM user_settings
WHERE user_id = ?
ORDER BY key;

-- name: GetUserSetting :one
SELECT id, user_id, key, value, created_at, updated_at
FROM user_settings
WHERE user_id = ? AND key = ?;

-- name: UpsertUserSetting :one
INSERT INTO user_settings (id, user_id, key, value, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT (user_id, key) DO UPDATE SET
    value = excluded.value,
    updated_at = CURRENT_TIMESTAMP
RETURNING id, user_id, key, value, created_at, updated_at;

-- name: DeleteUserSetting :exec
DELETE FROM user_settings WHERE user_id = ? AND key = ?;
