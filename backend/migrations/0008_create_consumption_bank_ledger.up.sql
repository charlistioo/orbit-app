CREATE TABLE consumption_bank_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    delta_amount NUMERIC(12,2) NOT NULL,
    reason TEXT NOT NULL,
    related_transaction_id UUID REFERENCES transactions(id),
    balance_after NUMERIC(12,2) NOT NULL CHECK (balance_after >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_consumption_bank_ledger_user_id_created_at ON consumption_bank_ledger (user_id, created_at);
