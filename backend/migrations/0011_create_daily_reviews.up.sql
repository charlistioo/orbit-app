CREATE TABLE daily_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    review_date DATE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('completed', 'skipped')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, review_date)
);

CREATE INDEX idx_daily_reviews_user_id_review_date ON daily_reviews (user_id, review_date);
