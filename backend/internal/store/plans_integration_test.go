package store

import (
	"context"
	"os"
	"testing"
)

func TestPlanStore_Integration(t *testing.T) {
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

	testUser, err := users.Create(ctx, "integration-plan-test-google-id", "plantest@example.com", "Plan Test")
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
	if len(plan.Categories) != 1 {
		t.Fatalf("got %d plan categories, want 1", len(plan.Categories))
	}
	planCategoryID := plan.Categories[0].ID

	fetched, found, err := plans.GetByDate(ctx, testUser.ID, planDate)
	if err != nil || !found {
		t.Fatalf("GetByDate: found=%v err=%v", found, err)
	}
	if fetched.BaselineAmount != 50000 {
		t.Errorf("got baseline %v, want 50000", fetched.BaselineAmount)
	}

	// No maximum enforced - a huge value must be accepted and persisted.
	updated, err := plans.UpdatePlanCategoryAmount(ctx, testUser.ID, planCategoryID, 999999999.99)
	if err != nil {
		t.Fatalf("UpdatePlanCategoryAmount: %v", err)
	}
	if updated.PlannedAmount != 999999999.99 {
		t.Errorf("got planned amount %v, want 999999999.99", updated.PlannedAmount)
	}

	var adjCount int
	var oldAmt, newAmt float64
	err = db.QueryRowContext(ctx,
		`SELECT count(*), min(old_amount), min(new_amount) FROM plan_adjustments WHERE plan_category_id = $1`,
		planCategoryID,
	).Scan(&adjCount, &oldAmt, &newAmt)
	if err != nil {
		t.Fatalf("querying plan_adjustments: %v", err)
	}
	if adjCount != 1 {
		t.Fatalf("got %d plan_adjustments rows, want 1", adjCount)
	}
	if oldAmt != 16000 || newAmt != 999999999.99 {
		t.Errorf("got adjustment old=%v new=%v, want old=16000 new=999999999.99", oldAmt, newAmt)
	}

	// Cross-user rejection: another user must not be able to edit this
	// plan category even if they somehow learn its id.
	otherUser, err := users.Create(ctx, "integration-plan-test-other-google-id", "other@example.com", "Other")
	if err != nil {
		t.Fatalf("creating other user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, otherUser.ID) })

	_, err = plans.UpdatePlanCategoryAmount(ctx, otherUser.ID, planCategoryID, 1)
	if err != ErrPlanCategoryNotFound {
		t.Errorf("got err %v, want ErrPlanCategoryNotFound for a plan category owned by a different user", err)
	}
}
