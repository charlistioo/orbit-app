CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    item_id UUID REFERENCES items(id),
    plan_category_id UUID REFERENCES plan_categories(id),
    amount NUMERIC(12,2) NOT NULL,
    plan_amount_snapshot NUMERIC(12,2),
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_user_id_occurred_at ON transactions (user_id, occurred_at);
