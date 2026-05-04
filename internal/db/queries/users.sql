-- name: CreateUser :one
INSERT INTO users (email, password_hash, username)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username   = COALESCE($2, username),
    bio        = COALESCE($3, bio),
    avatar_url = COALESCE($4, avatar_url),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateUserAvatar :one
UPDATE users
SET avatar_url = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;
