CREATE TABLE IF NOT EXISTS authorizations (
    transaction_id VARCHAR(255) PRIMARY KEY,
    amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);