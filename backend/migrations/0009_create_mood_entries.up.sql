CREATE TABLE mood_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    mood TEXT NOT NULL CHECK (mood IN ('very_good', 'good', 'neutral', 'stressed', 'sad')),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_mood_entries_user_id_recorded_at ON mood_entries (user_id, recorded_at);
