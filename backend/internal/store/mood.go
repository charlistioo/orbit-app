package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MoodEntry struct {
	ID         string
	Mood       string
	RecordedAt time.Time
}

type MoodStore struct {
	db *sql.DB
}

func NewMoodStore(db *sql.DB) *MoodStore {
	return &MoodStore{db: db}
}

// Create logs a mood entry with its own timestamp. Mood tracking is
// optional and can happen any number of times per day - there is no
// uniqueness constraint on (user, date).
func (s *MoodStore) Create(ctx context.Context, userID, mood string, recordedAt time.Time) (MoodEntry, error) {
	var e MoodEntry
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO mood_entries (user_id, mood, recorded_at) VALUES ($1, $2, $3)
		 RETURNING id, mood, recorded_at`,
		userID, mood, recordedAt,
	).Scan(&e.ID, &e.Mood, &e.RecordedAt)
	if err != nil {
		return MoodEntry{}, fmt.Errorf("creating mood entry: %w", err)
	}
	return e, nil
}

// Latest returns the user's most recently logged mood entry (regardless
// of date), for display as "current mood" on the Home screen. Mood is
// optional, so found=false is normal for a user who has never logged
// one.
func (s *MoodStore) Latest(ctx context.Context, userID string) (MoodEntry, bool, error) {
	var e MoodEntry
	err := s.db.QueryRowContext(ctx,
		`SELECT id, mood, recorded_at FROM mood_entries WHERE user_id = $1 ORDER BY recorded_at DESC LIMIT 1`,
		userID,
	).Scan(&e.ID, &e.Mood, &e.RecordedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return MoodEntry{}, false, nil
	}
	if err != nil {
		return MoodEntry{}, false, fmt.Errorf("reading latest mood: %w", err)
	}
	return e, true, nil
}
