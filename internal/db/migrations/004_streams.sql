-- +goose Up
CREATE TABLE streams (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id   UUID        REFERENCES categories(id) ON DELETE SET NULL,
    title         VARCHAR     NOT NULL DEFAULT 'Untitled Stream',
    status        VARCHAR     NOT NULL DEFAULT 'offline'
                              CHECK (status IN ('offline', 'live')),
    viewer_count  INT         NOT NULL DEFAULT 0,
    thumbnail_url VARCHAR,
    started_at    TIMESTAMPTZ,
    ended_at      TIMESTAMPTZ
);

-- One stream entity per user (their channel)
CREATE UNIQUE INDEX streams_user_idx ON streams (user_id);

-- +goose Down
DROP TABLE streams;
