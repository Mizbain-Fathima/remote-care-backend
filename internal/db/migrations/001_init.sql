CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email      TEXT UNIQUE NOT NULL,
    full_name  TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wallets (
    user_id    UUID PRIMARY KEY REFERENCES users(id),
    balance    NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency   TEXT NOT NULL DEFAULT 'INR',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE vouchers (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             TEXT NOT NULL,
    description      TEXT,
    category         TEXT NOT NULL, -- HOME_OFFICE / SKINCARE
    brand            TEXT,
    discount_percent NUMERIC(5,2) NOT NULL,
    price            NUMERIC(12,2) NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'INR',
    stock            INT NOT NULL DEFAULT 100,
    active           BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id),
    voucher_id     UUID NOT NULL REFERENCES vouchers(id),
    amount         NUMERIC(12,2) NOT NULL,
    currency       TEXT NOT NULL DEFAULT 'INR',
    status         TEXT NOT NULL, -- PENDING/SUCCESS/FAILED
    payment_method TEXT NOT NULL, -- MOCK_UPI
    upi_id         TEXT,
    request_id     TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_transactions_request_id
    ON transactions(request_id);

CREATE INDEX idx_vouchers_category
    ON vouchers(category, active);
