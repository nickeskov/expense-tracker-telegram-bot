-- +goose Up
-- +goose StatementBegin

CREATE DOMAIN currency_code AS VARCHAR(8) NOT NULL CHECK ( VALUE <> '' );

CREATE TABLE users
(
    id            BIGINT PRIMARY KEY,
    currency      currency_code,
    monthly_limit NUMERIC(25, 5) CHECK ( monthly_limit >= 0 )
);

CREATE TABLE expenses
(
    id       BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id  BIGINT REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    category VARCHAR(256)   NOT NULL CHECK ( category <> '' ),
    amount   NUMERIC(25, 5) NOT NULL CHECK ( amount > 0 ),
    date     DATE           NOT NULL,
    comment  VARCHAR(4096)  NOT NULL
);

CREATE TABLE exchange_rates
(
    id       BIGINT GENERATED ALWAYS AS IDENTITY,
    currency currency_code,
    date     DATE           NOT NULL,
    rate     NUMERIC(16, 8) NOT NULL CHECK ( rate > 0 ),
    UNIQUE (currency, date)
);

CREATE INDEX expenses_user_id_date_idx ON expenses (user_id, date);

CREATE UNIQUE INDEX exchange_rates_currency_date_idx ON exchange_rates (currency, date) INCLUDE (rate);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX exchange_rates_currency_date_idx;

DROP INDEX expenses_user_id_date_idx;

DROP TABLE exchange_rates CASCADE;

DROP TABLE expenses CASCADE;

DROP TABLE users CASCADE;

DROP DOMAIN currency_code;

-- +goose StatementEnd
