-- +goose Up
CREATE TABLE stream_keys (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        VARCHAR     UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

-- Only one active key per user at a time
CREATE UNIQUE INDEX stream_keys_active_user_idx
    ON stream_keys (user_id)
    WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE stream_keys;
