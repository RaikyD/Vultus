-- name: CreateStream :one
INSERT INTO streams (user_id)
VALUES ($1)
RETURNING *;

-- name: GetStreamByID :one
SELECT s.*, u.username, u.avatar_url AS user_avatar_url
FROM streams s
JOIN users u ON u.id = s.user_id
WHERE s.id = $1;

-- name: GetStreamByUserID :one
SELECT * FROM streams
WHERE user_id = $1
LIMIT 1;

-- name: ListStreams :many
SELECT s.*, u.username, u.avatar_url AS user_avatar_url
FROM streams s
JOIN users u ON u.id = s.user_id
WHERE
    ($1 = '' OR s.status = $1)
    AND ($2::uuid IS NULL OR s.category_id = $2)
    AND ($3 = '' OR u.username ILIKE '%' || $3 || '%' OR s.title ILIKE '%' || $3 || '%')
ORDER BY
    CASE WHEN s.status = 'live' THEN 0 ELSE 1 END,
    s.viewer_count DESC;

-- name: UpdateStreamStatus :one
UPDATE streams
SET status     = $2,
    started_at = CASE WHEN $2::varchar = 'live' THEN now() ELSE started_at END,
    ended_at   = CASE WHEN $2::varchar = 'offline' THEN now() ELSE ended_at END
WHERE user_id = $1
RETURNING *;

-- name: UpdateStreamViewerCount :exec
UPDATE streams
SET viewer_count = $2
WHERE id = $1;

-- name: ListLiveStreams :many
SELECT * FROM streams WHERE status = 'live';

-- name: ResetLiveStreams :exec
UPDATE streams SET status = 'offline', ended_at = now() WHERE status = 'live';

-- name: ResetViewerCount :exec
UPDATE streams SET viewer_count = 0 WHERE status = 'live';

-- name: UpdateStreamMeta :one
UPDATE streams
SET title       = COALESCE($2, title),
    category_id = COALESCE($3, category_id)
WHERE user_id = $1
RETURNING *;
