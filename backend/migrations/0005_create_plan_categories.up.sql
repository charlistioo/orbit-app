CREATE TABLE plan_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    daily_plan_id UUID NOT NULL REFERENCES daily_plans(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    planned_amount NUMERIC(12,2) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
