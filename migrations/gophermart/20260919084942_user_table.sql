-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    login VARCHAR(20) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_login_lower_idx ON users (LOWER(login));

-- +goose Down
DROP INDEX IF EXISTS users_login_lower_idx;
DROP TABLE IF EXISTS users;
