-- name: CreateChatMessage :one
INSERT INTO chat_messages (stream_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListRecentMessages :many
SELECT m.id, m.stream_id, m.user_id, m.content, m.created_at,
       u.username
FROM chat_messages m
JOIN users u ON u.id = m.user_id
WHERE m.stream_id = $1
  AND m.deleted_at IS NULL
ORDER BY m.created_at DESC
LIMIT $2;
