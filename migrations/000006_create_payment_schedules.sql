CREATE TABLE IF NOT EXISTS payment_schedules (
    id           UUID PRIMARY KEY,
    credit_id    UUID    NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
    due_date     DATE    NOT NULL,
    amount       NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    principal    NUMERIC(18,2) NOT NULL,
    interest     NUMERIC(18,2) NOT NULL,
    paid         BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_sched_credit   ON payment_schedules(credit_id);
CREATE INDEX IF NOT EXISTS idx_sched_due_date ON payment_schedules(due_date);