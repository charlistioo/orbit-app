package store

import (
	"context"
	"os"
	"testing"
)

// TestUserStore_CreateAndFindByGoogleID runs against a real PostgreSQL
// database (with migrations already applied) to verify the SQL in
// users.go is actually correct, not just reviewed as text. Skipped
// unless DATABASE_URL is set, since it needs a live database - see
// harness-notes/tasks/TASK-003.md for how to run it.
func TestUserStore_CreateAndFindByGoogleID(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set - skipping integration test (see TASK-003 notes)")
	}

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()

	googleID := "integration-test-google-id-12345"

	_, found, err := store.FindByGoogleID(ctx, googleID)
	if err != nil {
		t.Fatalf("FindByGoogleID (before create): %v", err)
	}
	if found {
		t.Fatal("expected no user before creation - test data collision, clean the test DB")
	}

	created, err := store.Create(ctx, googleID, "integration@example.com", "Integration Test")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected a generated id after create")
	}
	if created.GoogleID != googleID {
		t.Errorf("got GoogleID %q, want %q", created.GoogleID, googleID)
	}

	found2, ok, err := store.FindByGoogleID(ctx, googleID)
	if err != nil {
		t.Fatalf("FindByGoogleID (after create): %v", err)
	}
	if !ok {
		t.Fatal("expected to find the just-created user")
	}
	if found2.ID != created.ID {
		t.Errorf("got id %q on re-fetch, want %q", found2.ID, created.ID)
	}
	if found2.Email != "integration@example.com" {
		t.Errorf("got email %q, want %q", found2.Email, "integration@example.com")
	}

	// Cleanup so re-runs against the same DB don't collide.
	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE google_id = $1`, googleID)
}

// TestUserStore_UpdateProfileIsPartialAndCompletesOnboardingOnce proves
// TASK-016's onboarding/profile fields against real PostgreSQL: a
// partial update only changes the fields given, onboarding_completed_at
// is set on the first call and never reset by later calls.
func TestUserStore_UpdateProfileIsPartialAndCompletesOnboardingOnce(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set - skipping integration test")
	}

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	users := NewUserStore(db)

	created, err := users.Create(ctx, "profile-integration-google-id", "profile@example.com", "Original Name")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, created.ID) })

	if created.OnboardingCompleted {
		t.Fatal("expected a freshly created user to not have completed onboarding")
	}

	age := 21
	baseline := 90000.0
	updated, err := users.UpdateProfile(ctx, created.ID, ProfileUpdate{Age: &age, DefaultBaselineAmount: &baseline})
	if err != nil {
		t.Fatalf("UpdateProfile (first): %v", err)
	}
	if updated.DisplayName != "Original Name" {
		t.Errorf("got display name %q after a partial update that didn't touch it, want unchanged %q", updated.DisplayName, "Original Name")
	}
	if updated.Age == nil || *updated.Age != 21 {
		t.Errorf("got age %v, want 21", updated.Age)
	}
	if !updated.OnboardingCompleted {
		t.Error("expected onboarding_completed to become true after the first profile update")
	}

	newName := "Renamed"
	updated2, err := users.UpdateProfile(ctx, created.ID, ProfileUpdate{DisplayName: &newName})
	if err != nil {
		t.Fatalf("UpdateProfile (second): %v", err)
	}
	if updated2.DisplayName != "Renamed" {
		t.Errorf("got display name %q, want %q", updated2.DisplayName, "Renamed")
	}
	if updated2.Age == nil || *updated2.Age != 21 {
		t.Errorf("got age %v after a partial update that didn't touch it, want unchanged 21", updated2.Age)
	}
	if !updated2.OnboardingCompleted {
		t.Error("expected onboarding_completed to remain true on a later update")
	}

	fetched, found, err := users.GetByID(ctx, created.ID)
	if err != nil || !found {
		t.Fatalf("GetByID: found=%v err=%v", found, err)
	}
	if fetched.DisplayName != "Renamed" || fetched.Age == nil || *fetched.Age != 21 {
		t.Errorf("GetByID returned stale data: name=%q age=%v", fetched.DisplayName, fetched.Age)
	}
}
