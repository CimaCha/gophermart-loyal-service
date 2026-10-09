-- +goose Up
CREATE TABLE IF NOT EXISTS transactions (
    uuid         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uuid    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_num    VARCHAR(255) NOT NULL UNIQUE,
    sum          NUMERIC(12, 2) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_uuid ON transactions(user_uuid);

-- +goose Down
DROP TABLE IF EXISTS transactions