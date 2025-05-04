CREATE TABLE IF NOT EXISTS cards (
    id           UUID PRIMARY KEY,
    account_id   UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    number_enc   BYTEA   NOT NULL,
    expiry_enc   BYTEA   NOT NULL,
    cvv_hash     TEXT    NOT NULL,
    hmac         TEXT    NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);