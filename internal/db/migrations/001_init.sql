CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    description      TEXT,
    category         TEXT NOT NULL,
    brand            TEXT,
    discount_percent NUMERIC(5,2) NOT NULL,
    price            NUMERIC(12,2) NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'INR',
    stock            INT NOT NULL DEFAULT 100,
    active           BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id),
    voucher_id     UUID NOT NULL REFERENCES vouchers(id),
    amount         NUMERIC(12,2) NOT NULL,
    currency       TEXT NOT NULL DEFAULT 'INR',
    status         TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    upi_id         TEXT,
    request_id     TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_transactions_request_id ON transactions(request_id);

CREATE INDEX idx_vouchers_category ON vouchers(category, active);


INSERT INTO vouchers (name, description, category, brand, discount_percent, price, currency, stock, active)
VALUES
    ('GlowBoost Vitamin C Serum', 'Brightening face serum with 15% Vitamin C', 'SKINCARE', 'GlowBoost', 20, 499, 'INR', 100, true),
    ('HydraSoft Moisturizer', 'Deep hydration daily cream for all skin types', 'SKINCARE', 'HydraSoft', 25, 699, 'INR', 100, true),
    ('PureSkin Organic Face Wash', 'Gentle foaming cleanser with aloe vera', 'SKINCARE', 'PureSkin', 15, 299, 'INR', 100, true),
    ('DermaShield Sunscreen SPF 50', 'Broad-spectrum UV protection sunscreen', 'SKINCARE', 'DermaShield', 30, 399, 'INR', 100, true);

INSERT INTO vouchers (name, description, category, brand, discount_percent, price, currency, stock, active)
VALUES
    ('ErgoPro Adjustable Chair', 'High-back ergonomic chair with lumbar support', 'HOME_OFFICE', 'ErgoPro', 18, 8999, 'INR', 50, true),
    ('LiftUp Standing Desk', 'Electric height-adjustable standing desk', 'HOME_OFFICE', 'LiftUp', 22, 15999, 'INR', 40, true),
    ('FlexiLamp LED Desk Light', 'Smart LED lamp with adjustable brightness', 'HOME_OFFICE', 'FlexiLamp', 10, 1999, 'INR', 100, true),
    ('PostureMaster Foot Rest', 'Memory foam ergonomic footrest for desk', 'HOME_OFFICE', 'PostureMaster', 15, 1299, 'INR', 80, true);

INSERT INTO users (email, full_name)
VALUES ('testuser@example.com', 'Test User')
RETURNING id;

-- 0024fd90-b092-4a5e-aee6-d48d10d8a66e

INSERT INTO wallets (user_id, balance, currency)
VALUES ('0024fd90-b092-4a5e-aee6-d48d10d8a66e', 5000, 'INR');
