-- +goose Up
CREATE TYPE o_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders (
    order_num BIGINT PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL,
    order_status o_status DEFAULT 'NEW' NOT NULL,
    accrual NUMERIC(15, 2) DEFAULT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_orders_active_statuses
ON orders (order_status)
WHERE order_status IN ('NEW', 'PROCESSING');

-- +goose Down
DROP INDEX IF EXISTS idx_orders_active_statuses;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS o_status;