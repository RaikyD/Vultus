-- +goose Up
CREATE TABLE chat_messages (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id  UUID        NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    edited_at  TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX chat_messages_stream_id_idx ON chat_messages (stream_id, created_at DESC);

-- +goose Down
DROP TABLE chat_messages;
