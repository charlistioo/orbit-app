package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestTransactionStore_PlanSnapshotIsImmutable is the core proof for
// TASK-006: a transaction's plan_amount_snapshot must reflect the plan
// value in effect when it was recorded, and must never change even if
// the plan is edited afterwards.
func TestTransactionStore_PlanSnapshotIsImmutable(t *testing.T) {
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
	transactions := NewTransactionStore(db, plans)

	testUser, err := users.Create(ctx, "integration-tx-test-google-id", "txtest@example.com", "Tx Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID) })

	cat, err := categories.Create(ctx, testUser.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, cat.ID) })

	const planDate = "2026-09-16"
	plan, err := plans.CreateOrReplace(ctx, testUser.ID, planDate, 50000,
		[]CategoryPlanInput{{CategoryID: cat.ID, PlannedAmount: 16000}})
	if err != nil {
		t.Fatalf("CreateOrReplace: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM daily_plans WHERE id = $1`, plan.ID) })
	t.Cleanup(func() {
		for _, pc := range plan.Categories {
			db.ExecContext(ctx, `DELETE FROM plan_categories WHERE id = $1`, pc.ID)
		}
	})
	t.Cleanup(func() {
		for _, pc := range plan.Categories {
			db.ExecContext(ctx, `DELETE FROM plan_adjustments WHERE plan_category_id = $1`, pc.ID)
		}
	})
	planCategoryID := plan.Categories[0].ID

	occurredAt, _ := time.Parse("2006-01-02T15:04:05Z", "2026-09-16T12:00:00Z")

	// Record a transaction that exceeds the plan (16000) - must be
	// recorded in full, not blocked or reduced.
	tx, err := transactions.Create(ctx, testUser.ID, cat.ID, nil, 26000, occurredAt)
	if err != nil {
		t.Fatalf("Create transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	if tx.Amount != 26000 {
		t.Errorf("got recorded amount %v, want 26000 (full, unreduced)", tx.Amount)
	}
	if tx.PlanAmountSnapshot == nil || *tx.PlanAmountSnapshot != 16000 {
		t.Fatalf("got plan_amount_snapshot %v, want 16000 (the plan in effect at record time)", tx.PlanAmountSnapshot)
	}

	// Now edit the plan category to a different value.
	_, err = plans.UpdatePlanCategoryAmount(ctx, testUser.ID, planCategoryID, 99999)
	if err != nil {
		t.Fatalf("UpdatePlanCategoryAmount: %v", err)
	}

	// Re-read the transaction directly from the database - its snapshot
	// must be untouched by the plan edit above.
	var snapshotAfterEdit float64
	err = db.QueryRowContext(ctx, `SELECT plan_amount_snapshot FROM transactions WHERE id = $1`, tx.ID).Scan(&snapshotAfterEdit)
	if err != nil {
		t.Fatalf("re-reading transaction: %v", err)
	}
	if snapshotAfterEdit != 16000 {
		t.Errorf("got plan_amount_snapshot %v after later plan edit, want unchanged 16000 (immutability violated)", snapshotAfterEdit)
	}
}

// TestTransactionStore_NoPlanIsFine verifies recording a transaction
// with no plan at all for that date/category doesn't fail - plans are
// optional.
func TestTransactionStore_NoPlanIsFine(t *testing.T) {
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
	transactions := NewTransactionStore(db, plans)

	testUser, err := users.Create(ctx, "integration-tx-test-noplan-google-id", "txtest-noplan@example.com", "No Plan")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID) })

	cat, err := categories.Create(ctx, testUser.ID, "Jajan")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, cat.ID) })

	tx, err := transactions.Create(ctx, testUser.ID, cat.ID, nil, 5000, time.Now())
	if err != nil {
		t.Fatalf("Create transaction with no plan: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	if tx.PlanAmountSnapshot != nil {
		t.Errorf("got plan_amount_snapshot %v, want nil (no plan existed)", *tx.PlanAmountSnapshot)
	}
}
