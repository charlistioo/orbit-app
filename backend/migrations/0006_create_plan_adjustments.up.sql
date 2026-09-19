CREATE TABLE plan_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_category_id UUID NOT NULL REFERENCES plan_categories(id),
    old_amount NUMERIC(12,2) NOT NULL,
    new_amount NUMERIC(12,2) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
