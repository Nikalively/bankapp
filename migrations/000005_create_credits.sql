CREATE TABLE IF NOT EXISTS credits (
    id               UUID PRIMARY KEY,
    user_id          UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id       UUID    NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    principal        NUMERIC(18,2) NOT NULL CHECK (principal > 0),
    annual_rate      NUMERIC(5,4)  NOT NULL CHECK (annual_rate >= 0),
    term_months      INT          NOT NULL CHECK (term_months > 0),
    start_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    remaining        NUMERIC(18,2) NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_credits_user ON credits(user_id);