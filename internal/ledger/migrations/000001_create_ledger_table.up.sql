CREATE TABLE IF NOT EXISTS ledger (
    entry_id BIGSERIAL PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL,
    entry_type VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ledger_transaction_id ON ledger (transaction_id);