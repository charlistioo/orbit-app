package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConsumptionBank_ReconciliationAcrossMultipleTransactions(t *testing.T) {
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
	bank := NewConsumptionBankStore(db)

	testUser, err := users.Create(ctx, "integration-bank-test-google-id", "banktest@example.com", "Bank Test")
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

	occurredAt, _ := time.Parse("2006-01-02T15:04:05Z", "2026-09-16T09:00:00Z")

	record := func(amount float64) Transaction {
		tx, err := transactions.Create(ctx, testUser.ID, cat.ID, nil, amount, occurredAt)
		if err != nil {
			t.Fatalf("Create transaction: %v", err)
		}
		t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })
		if tx.PlanAmountSnapshot == nil {
			t.Fatalf("expected a plan snapshot for this transaction")
		}
		if err := bank.ReconcileForTransaction(ctx, testUser.ID, cat.ID, occurredAt, tx.ID, *tx.PlanAmountSnapshot); err != nil {
			t.Fatalf("ReconcileForTransaction: %v", err)
		}
		return tx
	}

	// tx1 = 12000, under the 16000 plan -> credit 4000.
	record(12000)
	balance, err := bank.CurrentBalance(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("CurrentBalance: %v", err)
	}
	if balance != 4000 {
		t.Fatalf("after tx1 (12000): got balance %v, want 4000", balance)
	}

	// tx2 = 3000 more (cumulative 15000, still under 16000) -> credit
	// shrinks to 1000 (a -3000 reversal entry, never editing the first).
	record(3000)
	balance, err = bank.CurrentBalance(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("CurrentBalance: %v", err)
	}
	if balance != 1000 {
		t.Fatalf("after tx2 (cumulative 15000): got balance %v, want 1000", balance)
	}

	// tx3 = 5000 more (cumulative 20000, now over the 16000 plan) ->
	// credit fully reversed to 0 for this category/day.
	record(5000)
	balance, err = bank.CurrentBalance(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("CurrentBalance: %v", err)
	}
	if balance != 0 {
		t.Fatalf("after tx3 (cumulative 20000, over plan): got balance %v, want 0", balance)
	}

	// Registered here (after all 3 transaction cleanups above) so it
	// executes before them - consumption_bank_ledger.related_transaction_id
	// references transactions, so the ledger rows must be gone before their
	// transactions are deleted.
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM consumption_bank_ledger WHERE user_id = $1`, testUser.ID) })

	// The ledger must be append-only: 3 reconciliation entries, no
	// updates, summing to exactly the final balance.
	ledger, err := bank.Ledger(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("Ledger: %v", err)
	}
	if len(ledger) != 3 {
		t.Fatalf("got %d ledger entries, want 3 (one per transaction, never edited)", len(ledger))
	}
	var sum float64
	for _, e := range ledger {
		sum += e.DeltaAmount
	}
	if sum != 0 {
		t.Errorf("ledger deltas sum to %v, want 0 (matching final balance)", sum)
	}
}

func TestConsumptionBank_ApplyRespectsCapAndFloor(t *testing.T) {
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
	bank := NewConsumptionBankStore(db)

	testUser, err := users.Create(ctx, "integration-bank-apply-test-google-id", "bankapplytest@example.com", "Bank Apply Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID) })

	catA, _ := categories.Create(ctx, testUser.ID, "Kopi")
	catB, _ := categories.Create(ctx, testUser.ID, "Makanan")
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id IN ($1, $2)`, catA.ID, catB.ID) })

	const planDate = "2026-09-16"
	plan, err := plans.CreateOrReplace(ctx, testUser.ID, planDate, 50000, []CategoryPlanInput{
		{CategoryID: catA.ID, PlannedAmount: 16000},
		{CategoryID: catB.ID, PlannedAmount: 20000},
	})
	if err != nil {
		t.Fatalf("CreateOrReplace: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM daily_plans WHERE id = $1`, plan.ID) })
	t.Cleanup(func() {
		for _, pc := range plan.Categories {
			db.ExecContext(ctx, `DELETE FROM plan_categories WHERE id = $1`, pc.ID)
		}
	})

	occurredAt, _ := time.Parse("2006-01-02T15:04:05Z", "2026-09-16T09:00:00Z")

	// Underspend on Kopi: plan 16000, spend 6000 -> credit 10000.
	underTx, err := transactions.Create(ctx, testUser.ID, catA.ID, nil, 6000, occurredAt)
	if err != nil {
		t.Fatalf("Create underspend transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, underTx.ID) })
	if err := bank.ReconcileForTransaction(ctx, testUser.ID, catA.ID, occurredAt, underTx.ID, *underTx.PlanAmountSnapshot); err != nil {
		t.Fatalf("ReconcileForTransaction: %v", err)
	}
	balance, _ := bank.CurrentBalance(ctx, testUser.ID)
	if balance != 10000 {
		t.Fatalf("got balance %v after underspend, want 10000", balance)
	}

	// Overspend on Makanan: plan 20000, spend 30000 -> overspend 10000.
	overTx, err := transactions.Create(ctx, testUser.ID, catB.ID, nil, 30000, occurredAt)
	if err != nil {
		t.Fatalf("Create overspend transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, overTx.ID) })
	// No reconciliation call needed - overspend transactions don't
	// generate a positive reconcile delta (recordedAfter clamps to 0),
	// but call it anyway since a real client always would.
	if err := bank.ReconcileForTransaction(ctx, testUser.ID, catB.ID, occurredAt, overTx.ID, *overTx.PlanAmountSnapshot); err != nil {
		t.Fatalf("ReconcileForTransaction (overspend): %v", err)
	}

	// Registered here (after both transaction cleanups above) so it
	// executes before them - consumption_bank_ledger.related_transaction_id
	// references transactions, so the ledger rows must be gone before their
	// transactions are deleted.
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM consumption_bank_ledger WHERE user_id = $1`, testUser.ID) })

	// Apply with no explicit amount -> should cap at 10% of the 10000
	// balance = 1000, well under the 10000 overspend on catB.
	entry, err := bank.Apply(ctx, testUser.ID, overTx.ID, nil)
	if err != nil {
		t.Fatalf("Apply (max): %v", err)
	}
	if entry.DeltaAmount != -1000 {
		t.Errorf("got applied delta %v, want -1000 (10%% of 10000 balance, capped)", entry.DeltaAmount)
	}
	if entry.BalanceAfter != 9000 {
		t.Errorf("got balance_after %v, want 9000", entry.BalanceAfter)
	}

	// Try to apply more than the cap explicitly - must still be capped
	// at 10% of the NEW balance (9000 -> max 900), never at the client's
	// requested value, and must never push the balance below 0.
	huge := 999999.0
	entry2, err := bank.Apply(ctx, testUser.ID, overTx.ID, &huge)
	if err != nil {
		t.Fatalf("Apply (huge request): %v", err)
	}
	if entry2.DeltaAmount != -900 {
		t.Errorf("got applied delta %v, want -900 (10%% of 9000 balance)", entry2.DeltaAmount)
	}
	if entry2.BalanceAfter < 0 {
		t.Fatalf("balance went negative: %v", entry2.BalanceAfter)
	}

	// The overspend transaction's own recorded amount must be totally
	// unchanged by any of this.
	var storedAmount float64
	err = db.QueryRowContext(ctx, `SELECT amount FROM transactions WHERE id = $1`, overTx.ID).Scan(&storedAmount)
	if err != nil {
		t.Fatalf("re-reading overspend transaction: %v", err)
	}
	if storedAmount != 30000 {
		t.Errorf("got transaction amount %v after applying bank, want unchanged 30000", storedAmount)
	}
}
