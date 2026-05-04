-- +goose Up
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR     UNIQUE NOT NULL,
    password_hash VARCHAR     NOT NULL,
    username      VARCHAR     UNIQUE NOT NULL,
    avatar_url    VARCHAR,
    bio           TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
