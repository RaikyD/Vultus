-- +goose Up
CREATE TABLE subscriptions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    streamer_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT subscriptions_unique UNIQUE (subscriber_id, streamer_id),
    CONSTRAINT subscriptions_no_self CHECK (subscriber_id <> streamer_id)
);

CREATE INDEX subscriptions_subscriber_idx ON subscriptions (subscriber_id);
CREATE INDEX subscriptions_streamer_idx ON subscriptions (streamer_id);

-- +goose Down
DROP TABLE subscriptions;
