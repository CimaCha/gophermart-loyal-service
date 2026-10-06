-- +goose Up
CREATE TYPE accrual_order_status AS ENUM ('REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS accrual_orders (
                                      order_num VARCHAR(255) PRIMARY KEY NOT NULL,
                                      order_status accrual_order_status DEFAULT 'REGISTERED' NOT NULL,
                                      accrual NUMERIC(15, 2) DEFAULT NULL,
                                      uploaded_at TIMESTAMPTZ DEFAULT NOW()
                                  );

CREATE TABLE IF NOT EXISTS goods (
                                      good_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      description        VARCHAR(255) NOT NULL,
                                      price NUMERIC(15, 2) NOT NULL,
                                      order_num           VARCHAR(255) REFERENCES accrual_orders(order_num)
);

CREATE INDEX IF NOT EXISTS idx_accrual_orders_active_statuses
    ON accrual_orders (order_status)
    WHERE order_status IN ('REGISTERED', 'PROCESSING');

CREATE INDEX idx_goods_order_num ON goods (order_num);

-- +goose Down
DROP INDEX IF EXISTS idx_goods_order_num;
DROP INDEX IF EXISTS idx_accrual_orders_active_statuses;
DROP TABLE IF EXISTS goods;
DROP TABLE IF EXISTS accrual_orders;
DROP TYPE IF EXISTS accrual_order_status;
