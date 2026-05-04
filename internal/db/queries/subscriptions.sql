-- name: Subscribe :exec
INSERT INTO subscriptions (subscriber_id, streamer_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: Unsubscribe :exec
DELETE FROM subscriptions
WHERE subscriber_id = $1 AND streamer_id = $2;

-- name: IsSubscribed :one
SELECT EXISTS (
    SELECT 1 FROM subscriptions
    WHERE subscriber_id = $1 AND streamer_id = $2
) AS subscribed;

-- name: ListSubscriptions :many
SELECT u.id, u.username, u.avatar_url, s.created_at AS subscribed_at
FROM subscriptions s
JOIN users u ON u.id = s.streamer_id
WHERE s.subscriber_id = $1
ORDER BY s.created_at DESC;

-- name: GetStreamerByUsername :one
SELECT u.id, u.username, u.avatar_url, u.bio,
       st.id AS stream_id, st.title, st.status, st.viewer_count, st.category_id
FROM users u
JOIN streams st ON st.user_id = u.id
WHERE u.username = $1;

-- name: CountSubscribers :one
SELECT COUNT(*) FROM subscriptions WHERE streamer_id = $1;
