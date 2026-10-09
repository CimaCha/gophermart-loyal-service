-- +goose Up

CREATE TYPE reward_type_enum AS ENUM ('POINTS', 'PERCENT');

CREATE TABLE rewards (
    uuid         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match        VARCHAR(255) NOT NULL UNIQUE,
    reward_value NUMERIC(15, 2) NOT NULL,
    reward_type  reward_type_enum DEFAULT 'POINTS' NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS rewards;