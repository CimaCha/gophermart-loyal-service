-- +goose Up
CREATE TYPE gophermart_order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS gophermart_orders (
    order_num VARCHAR(255) PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL,
    order_status gophermart_order_status DEFAULT 'NEW' NOT NULL,
    accrual NUMERIC(15, 2) DEFAULT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_gophermart_orders_active_statuses
ON gophermart_orders (order_status)
WHERE order_status IN ('NEW', 'PROCESSING');

-- +goose Down
DROP INDEX IF EXISTS idx_gophermart_orders_active_statuses;
DROP TABLE IF EXISTS gophermart_orders;
DROP TYPE IF EXISTS gophermart_order_status;