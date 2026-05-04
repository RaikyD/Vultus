-- name: CreateStreamKey :one
INSERT INTO stream_keys (user_id, key)
VALUES ($1, $2)
RETURNING *;

-- name: GetActiveStreamKey :one
SELECT * FROM stream_keys
WHERE user_id = $1 AND revoked_at IS NULL
LIMIT 1;

-- name: GetStreamKeyByKey :one
SELECT * FROM stream_keys
WHERE key = $1 AND revoked_at IS NULL
LIMIT 1;

-- name: RevokeStreamKey :exec
UPDATE stream_keys
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;
