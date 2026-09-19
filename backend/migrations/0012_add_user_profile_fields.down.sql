ALTER TABLE users
    DROP COLUMN IF EXISTS age,
    DROP COLUMN IF EXISTS avatar_data,
    DROP COLUMN IF EXISTS default_baseline_amount,
    DROP COLUMN IF EXISTS onboarding_completed_at;
