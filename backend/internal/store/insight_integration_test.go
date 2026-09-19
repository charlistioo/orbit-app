package store

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestInsightStore_NoTrendClaimWithOneDayOfHistory is the end-to-end
// proof of TASK-012's first acceptance criterion: a user with only 1
// day of recorded transactions must never receive a multi-day trend
// claim, even when that one day's spend is high.
func TestInsightStore_NoTrendClaimWithOneDayOfHistory(t *testing.T) {
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
	insights := NewInsightStore(db)

	user, err := users.Create(ctx, "insight-onedays-google-id", "insight-oneday@example.com", "Insight One Day Test")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })

	category, err := categories.Create(ctx, user.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, category.ID) })

	day := time.Date(2020, 1, 10, 9, 0, 0, 0, time.UTC)
	tx, err := txStore.Create(ctx, user.ID, category.ID, nil, 30000, day)
	if err != nil {
		t.Fatalf("creating transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	result, err := insights.DailyInsights(ctx, user.ID, "2020-01-10")
	if err != nil {
		t.Fatalf("DailyInsights: %v", err)
	}

	for _, msg := range result.Insights {
		if strings.Contains(msg, "berturut-turut") {
			t.Fatalf("expected no trend claim with only 1 day of history, got: %q", msg)
		}
	}
}

// TestInsightStore_TrendObservationFor3ConsecutiveRisingDays is the
// end-to-end proof of the second acceptance criterion.
func TestInsightStore_TrendObservationFor3ConsecutiveRisingDays(t *testing.T) {
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
	insights := NewInsightStore(db)

	user, err := users.Create(ctx, "insight-trend-google-id", "insight-trend@example.com", "Insight Trend Test")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })

	category, err := categories.Create(ctx, user.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, category.ID) })

	amounts := []float64{10000, 20000, 30000}
	dates := []time.Time{
		time.Date(2020, 2, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2020, 2, 2, 9, 0, 0, 0, time.UTC),
		time.Date(2020, 2, 3, 9, 0, 0, 0, time.UTC),
	}
	for i, occurredAt := range dates {
		tx, err := txStore.Create(ctx, user.ID, category.ID, nil, amounts[i], occurredAt)
		if err != nil {
			t.Fatalf("creating transaction: %v", err)
		}
		txID := tx.ID
		t.Cleanup(func() {
			db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, txID)
		})
	}

	result, err := insights.DailyInsights(ctx, user.ID, "2020-02-03")
	if err != nil {
		t.Fatalf("DailyInsights: %v", err)
	}

	found := false
	for _, msg := range result.Insights {
		if strings.Contains(msg, "Kopi") && strings.Contains(msg, "berturut-turut") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a trend observation for Kopi's 3 consecutive rising days, got: %v", result.Insights)
	}
}

// TestInsightStore_MoodNeverStatedAsCause is the end-to-end proof of the
// third acceptance criterion.
func TestInsightStore_MoodNeverStatedAsCause(t *testing.T) {
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
	moods := NewMoodStore(db)
	insights := NewInsightStore(db)

	user, err := users.Create(ctx, "insight-mood-google-id", "insight-mood@example.com", "Insight Mood Test")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })

	category, err := categories.Create(ctx, user.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, category.ID) })

	moodEntry, err := moods.Create(ctx, user.ID, "stressed", time.Date(2020, 3, 5, 8, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("creating mood entry: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM mood_entries WHERE id = $1`, moodEntry.ID) })

	tx, err := txStore.Create(ctx, user.ID, category.ID, nil, 30000, time.Date(2020, 3, 5, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("creating transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	result, err := insights.DailyInsights(ctx, user.ID, "2020-03-05")
	if err != nil {
		t.Fatalf("DailyInsights: %v", err)
	}

	found := false
	causalPhrases := []string{"karena mood", "akibat mood", "disebabkan", "menyebabkan pengeluaran"}
	for _, msg := range result.Insights {
		if strings.Contains(msg, "mood") {
			found = true
			for _, phrase := range causalPhrases {
				if strings.Contains(msg, phrase) {
					t.Fatalf("mood insight must never state causation, got: %q", msg)
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected a mood-related insight when a mood entry preceded a transaction, got: %v", result.Insights)
	}
}

// TestInsightStore_ConsumptionBankMovementReported proves the added
// Consumption Bank insight: an auto-reconcile credit recorded on the
// date is surfaced in the daily report.
func TestInsightStore_ConsumptionBankMovementReported(t *testing.T) {
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
	bank := NewConsumptionBankStore(db)
	insights := NewInsightStore(db)

	user, err := users.Create(ctx, "insight-bank-google-id", "insight-bank@example.com", "Insight Bank Test")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID) })

	category, err := categories.Create(ctx, user.ID, "Kopi")
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, category.ID) })

	const date = "2020-04-01"
	plan, err := plans.CreateOrReplace(ctx, user.ID, date, 20000, []CategoryPlanInput{
		{CategoryID: category.ID, PlannedAmount: 20000},
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

	occurredAt := time.Date(2020, 4, 1, 9, 0, 0, 0, time.UTC)
	tx, err := txStore.Create(ctx, user.ID, category.ID, nil, 12000, occurredAt)
	if err != nil {
		t.Fatalf("creating transaction: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, tx.ID) })

	if err := bank.ReconcileForTransaction(ctx, user.ID, category.ID, occurredAt, tx.ID, 20000); err != nil {
		t.Fatalf("reconciling: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM consumption_bank_ledger WHERE user_id = $1`, user.ID) })

	result, err := insights.DailyInsights(ctx, user.ID, date)
	if err != nil {
		t.Fatalf("DailyInsights: %v", err)
	}

	found := false
	for _, msg := range result.Insights {
		if strings.Contains(msg, "masuk ke Consumption Bank") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a Consumption Bank credited insight (underspent Rp8.000), got: %v", result.Insights)
	}
}

// TestInsightStore_MostFrequentCategoryReported proves the added
// category-frequency insight over a multi-day window.
func TestInsightStore_MostFrequentCategoryReported(t *testing.T) {
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
	insights := NewInsightStore(db)

	user, err := users.Create(ctx, "insight-freq-google-id", "insight-freq@example.com", "Insight Frequency Test")
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

	txTimes := []struct {
		categoryID string
		occurredAt time.Time
	}{
		{kopi.ID, time.Date(2020, 5, 1, 8, 0, 0, 0, time.UTC)},
		{kopi.ID, time.Date(2020, 5, 2, 8, 0, 0, 0, time.UTC)},
		{kopi.ID, time.Date(2020, 5, 3, 8, 0, 0, 0, time.UTC)},
		{makanan.ID, time.Date(2020, 5, 3, 12, 0, 0, 0, time.UTC)},
	}
	for _, tt := range txTimes {
		tx, err := txStore.Create(ctx, user.ID, tt.categoryID, nil, 10000, tt.occurredAt)
		if err != nil {
			t.Fatalf("creating transaction: %v", err)
		}
		txID := tx.ID
		t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM transactions WHERE id = $1`, txID) })
	}

	result, err := insights.WeeklyInsights(ctx, user.ID, "2020-05-03")
	if err != nil {
		t.Fatalf("WeeklyInsights: %v", err)
	}

	found := false
	for _, msg := range result.Insights {
		if strings.Contains(msg, "paling sering dipakai") && strings.Contains(msg, "Kopi") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Kopi reported as most frequent category, got: %v", result.Insights)
	}
}
