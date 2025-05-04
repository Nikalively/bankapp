CREATE TABLE IF NOT EXISTS transactions (
    id                UUID PRIMARY KEY,
    from_account_id   UUID        NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    to_account_id     UUID        REFERENCES accounts(id) ON DELETE SET NULL,
    amount            NUMERIC(18,2) NOT NULL CHECK (amount >= 0),
    transaction_type  VARCHAR(20) NOT NULL,
    note              TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account   ON transactions(to_account_id);