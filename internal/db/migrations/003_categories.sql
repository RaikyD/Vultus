-- +goose Up
CREATE TABLE categories (
    id   UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR UNIQUE NOT NULL
);

INSERT INTO categories (name) VALUES
    ('Gaming'),
    ('Music'),
    ('Just Chatting'),
    ('Art'),
    ('Sports'),
    ('Technology'),
    ('Education'),
    ('Other');

-- +goose Down
DROP TABLE categories;
