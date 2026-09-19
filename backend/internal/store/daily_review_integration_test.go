package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestDailyReviewStore_PendingCategoriesExcludesRecordedOnes proves
// PendingCategories only reports planned categories with no transaction
// yet, and excludes ones that already have one.
func TestDailyReviewStore_PendingCategoriesExcludesRecordedOnes(t *testing.T) {
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
	categories := NewCategoryStore(db)
	plans := NewPlanStore(db)
	txStore := NewTransactionStore(db, plans)
	reviews := NewDailyReviewStore(db)

	user, err := users.Create(ctx, "review-pending-google-id", "review-pending@example.com", "Review Pending Test")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })

	kopi, err := categories.Create(ctx, user.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, kopi.ID) })

	makanan, err := categories.Create(ctx, user.ID, "Makanan")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, makanan.ID) })

	const date = "2020-06-01"
	plan, err := plans.CreateOrReplace(ctx, user.ID, date, 30000, []CategoryPlanInput{
		{CategoryID: kopi.ID, PlannedAmount: 15000},
		{CategoryID: makanan.ID, PlannedAmount: 15000},
	})
	if err != nil {
		t.Fatalf("creating plan: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM daily_plans WHERE id = $1`, plan.ID) })
	t.Cleanup(func() {
		for _, pc := range plan.Categories {
			db.ExecContext(ctx, `DELETE FROM plan_categories WHERE id = $1`, pc.ID)
		}
	})

	occurredAt := time.Date(2020, 6, 1, 9, 0, 0, 0, time.UTC)
	tx, err := txStore.Create(ctx, user.ID, kopi.ID, nil, 10000, occurredAt)
	if err != nil {
		t.Fatalf("creating transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	pending, err := reviews.PendingCategories(ctx, user.ID, date)
	if err != nil {
		t.Fatalf("PendingCategories: %v", err)
	}

	if len(pending) != 1 || pending[0].Name != "Makanan" {
		t.Fatalf("expected only Makanan pending (Kopi already has a transaction), got: %v", pending)
	}
}

// TestDailyReviewStore_SkipIsValidAndOverwritable proves skipping a day
// is stored as its own valid status, and that finalizing again (e.g.
// switching from skip to complete) overwrites rather than duplicating.
func TestDailyReviewStore_SkipIsValidAndOverwritable(t *testing.T) {
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
	reviews := NewDailyReviewStore(db)

	user, err := users.Create(ctx, "review-skip-google-id", "review-skip@example.com", "Review Skip Test")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM daily_reviews WHERE user_id = $1`, user.ID) })

	const date = "2020-06-02"

	review, err := reviews.Finalize(ctx, user.ID, date, DailyReviewSkipped)
	if err != nil {
		t.Fatalf("Finalize (skip): %v", err)
	}
	if review.Status != DailyReviewSkipped {
		t.Fatalf("got status %v, want skipped", review.Status)
	}

	fetched, found, err := reviews.GetByDate(ctx, user.ID, date)
	if err != nil || !found {
		t.Fatalf("GetByDate: found=%v err=%v", found, err)
	}
	if fetched.Status != DailyReviewSkipped {
		t.Fatalf("got persisted status %v, want skipped", fetched.Status)
	}

	// Switching to "completed" for the same date must overwrite, not
	// create a second row (enforced by the unique(user_id, review_date)
	// constraint plus ON CONFLICT DO UPDATE).
	updated, err := reviews.Finalize(ctx, user.ID, date, DailyReviewCompleted)
	if err != nil {
		t.Fatalf("Finalize (complete): %v", err)
	}
	if updated.Status != DailyReviewCompleted {
		t.Fatalf("got status %v after re-finalizing, want completed", updated.Status)
	}
	if updated.ID != review.ID {
		t.Fatalf("expected the same row to be updated (id %s), got a different id %s", review.ID, updated.ID)
	}
}
