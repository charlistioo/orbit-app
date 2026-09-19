ALTER TABLE users
    ADD COLUMN age INT,
    ADD COLUMN avatar_data TEXT,
    ADD COLUMN default_baseline_amount NUMERIC,
    ADD COLUMN onboarding_completed_at TIMESTAMPTZ;
