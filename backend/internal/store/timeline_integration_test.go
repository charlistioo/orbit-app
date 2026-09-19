package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestTimelineStore_MergesAllFourSourcesChronologically is the core
// proof for TASK-010: a transaction, a plan adjustment, a mood entry,
// and a Consumption Bank ledger entry - created out of chronological
// order - must come back merged into one correctly time-sorted list.
func TestTimelineStore_MergesAllFourSourcesChronologically(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set - skipping integration test")
	}

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	// Registered first, so (LIFO) it closes last - after every other
	// t.Cleanup below has had a chance to use the connection to delete
	// its row. A plain `defer db.Close()` here would run at function
	// return, before t.Cleanup callbacks fire, silently breaking them.
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	users := NewUserStore(db)
	categories := NewCategoryStore(db)
	plans := NewPlanStore(db)
	transactions := NewTransactionStore(db, plans)
	bank := NewConsumptionBankStore(db)
	mood := NewMoodStore(db)
	timeline := NewTimelineStore(db)

	testUser, err := users.Create(ctx, "integration-timeline-test-google-id", "timelinetest@example.com", "Timeline Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	// t.Cleanup (not defer) throughout this test, consistently, so
	// cleanup runs in reverse-of-creation (LIFO) order and always
	// deletes child rows before the parents they reference - mixing
	// defer and t.Cleanup runs them in different phases and can delete
	// a parent while children still exist, as this test's first version
	// discovered the hard way (FK violations during cleanup).
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID) })

	cat, err := categories.Create(ctx, testUser.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, cat.ID) })

	// Use a plan_date far in the past (rather than "today") so that
	// transaction/mood timestamps we set explicitly are guaranteed to
	// sort before the bank ledger and plan adjustment events below,
	// which always stamp themselves with the real current time - we
	// can't control those, so we control everything else relative to
	// them instead of relying on "today" vs "now" being different.
	const pastDate = "2020-01-01"
	plan, err := plans.CreateOrReplace(ctx, testUser.ID, pastDate, 50000,
		[]CategoryPlanInput{{CategoryID: cat.ID, PlannedAmount: 16000}})
	if err != nil {
		t.Fatalf("CreateOrReplace: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM daily_plans WHERE id = $1`, plan.ID) })
	planCategoryID := plan.Categories[0].ID
	// plan_categories rows are created implicitly by CreateOrReplace and
	// have no cascade delete from daily_plans - must be cleaned up
	// explicitly, registered after (so it runs before) the daily_plans
	// cleanup above.
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM plan_categories WHERE daily_plan_id = $1`, plan.ID) })

	// Create events deliberately out of chronological order to prove
	// the merge actually sorts, rather than happening to preserve
	// insertion order.

	// 2nd chronologically (of the two we control): mood at 10:30.
	moodEntry, err := mood.Create(ctx, testUser.ID, "good", mustParse(t, pastDate+"T10:30:00Z"))
	if err != nil {
		t.Fatalf("mood.Create: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM mood_entries WHERE id = $1`, moodEntry.ID) })

	// 1st chronologically: transaction at 09:00.
	tx, err := transactions.Create(ctx, testUser.ID, cat.ID, nil, 12000, mustParse(t, pastDate+"T09:00:00Z"))
	if err != nil {
		t.Fatalf("transactions.Create: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	// The bank credit this transaction triggers and the plan adjustment
	// below both stamp themselves with the real current time (no
	// client-supplied override exists for either) - they'll always sort
	// after the two events above, which are pinned to 2020.
	if err := bank.ReconcileForTransaction(ctx, testUser.ID, cat.ID, mustParse(t, pastDate+"T09:00:00Z"), tx.ID, *tx.PlanAmountSnapshot); err != nil {
		t.Fatalf("ReconcileForTransaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM consumption_bank_ledger WHERE user_id = $1`, testUser.ID) })

	if _, err := plans.UpdatePlanCategoryAmount(ctx, testUser.ID, planCategoryID, 25000); err != nil {
		t.Fatalf("UpdatePlanCategoryAmount: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM plan_adjustments WHERE plan_category_id = $1`, planCategoryID) })

	events, err := timeline.ListEvents(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("got %d events, want 4 (one per source)", len(events))
	}

	// The bank ledger entry and plan adjustment both get "now" as their
	// timestamp (created_at / changed_at have no client-supplied
	// override), so their exact relative order versus each other isn't
	// deterministic in this test - what we CAN assert deterministically:
	// the transaction (explicit 09:00) must come first, and the mood
	// entry (explicit 10:30) must come before both "now" events, since
	// "now" during a test run is far later than 2026-09-16.
	if events[0].Type != "transaction" {
		t.Errorf("got first event type %q, want \"transaction\" (earliest explicit timestamp)", events[0].Type)
	}
	if events[1].Type != "mood" {
		t.Errorf("got second event type %q, want \"mood\"", events[1].Type)
	}
	// Verify the list is actually sorted, regardless of which of the
	// remaining two comes first.
	for i := 1; i < len(events); i++ {
		if events[i].OccurredAt.Before(events[i-1].OccurredAt) {
			t.Fatalf("timeline is not chronologically sorted: event %d (%v) is before event %d (%v)",
				i, events[i].OccurredAt, i-1, events[i-1].OccurredAt)
		}
	}

	types := map[string]bool{}
	for _, e := range events {
		types[e.Type] = true
	}
	for _, want := range []string{"transaction", "plan_adjustment", "mood", "bank_ledger"} {
		if !types[want] {
			t.Errorf("expected an event of type %q in the timeline, none found", want)
		}
	}
}

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parsing time %q: %v", s, err)
	}
	return parsed
}
