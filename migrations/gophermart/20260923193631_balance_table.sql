-- +goose Up
CREATE TABLE IF NOT EXISTS balance (
    user_uuid  UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current    NUMERIC(12, 2) NOT NULL DEFAULT 0,
    withdrawn  NUMERIC(12, 2) NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS balance;