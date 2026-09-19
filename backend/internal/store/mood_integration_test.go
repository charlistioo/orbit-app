package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMoodStore_MultipleEntriesSameDay(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set - skipping integration test")
	}

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	users := NewUserStore(db)
	mood := NewMoodStore(db)

	testUser, err := users.Create(ctx, "integration-mood-test-google-id", "moodtest@example.com", "Mood Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID)
	defer db.ExecContext(ctx, `DELETE FROM mood_entries WHERE user_id = $1`, testUser.ID)

	morning, _ := time.Parse(time.RFC3339, "2026-09-16T09:00:00Z")
	afternoon, _ := time.Parse(time.RFC3339, "2026-09-16T15:00:00Z")
	evening, _ := time.Parse(time.RFC3339, "2026-09-16T20:00:00Z")

	e1, err := mood.Create(ctx, testUser.ID, "good", morning)
	if err != nil {
		t.Fatalf("Create (morning): %v", err)
	}
	e2, err := mood.Create(ctx, testUser.ID, "neutral", afternoon)
	if err != nil {
		t.Fatalf("Create (afternoon): %v", err)
	}
	e3, err := mood.Create(ctx, testUser.ID, "stressed", evening)
	if err != nil {
		t.Fatalf("Create (evening): %v", err)
	}

	if e1.ID == e2.ID || e2.ID == e3.ID || e1.ID == e3.ID {
		t.Fatal("expected 3 distinct mood entries, got overlapping ids")
	}

	var count int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM mood_entries WHERE user_id = $1`, testUser.ID).Scan(&count)
	if err != nil {
		t.Fatalf("counting mood entries: %v", err)
	}
	if count != 3 {
		t.Errorf("got %d mood_entries rows, want 3 (same day, each with its own timestamp)", count)
	}
}
