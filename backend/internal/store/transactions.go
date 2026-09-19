package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Transaction struct {
	ID                  string
	CategoryID          string
	ItemID              *string
	PlanCategoryID      *string
	Amount              float64
	PlanAmountSnapshot  *float64
	OccurredAt          time.Time
}

type TransactionStore struct {
	db    *sql.DB
	plans *PlanStore
}

func NewTransactionStore(db *sql.DB, plans *PlanStore) *TransactionStore {
	return &TransactionStore{db: db, plans: plans}
}

// Create records an actual consumption transaction. It is never blocked
// or reduced regardless of the plan - it snapshots whatever plan amount
// is in effect for that category on that date (if any) at the moment of
// recording, into plan_amount_snapshot, so that a later plan edit can
// never retroactively change what this transaction shows.
func (s *TransactionStore) Create(ctx context.Context, userID, categoryID string, itemID *string, amount float64, occurredAt time.Time) (Transaction, error) {
	planDate := occurredAt.UTC().Format("2006-01-02")

	var planCategoryID *string
	var planAmountSnapshot *float64

	pcID, plannedAmount, found, err := s.plans.FindPlanCategoryForDate(ctx, userID, categoryID, planDate)
	if err != nil {
		return Transaction{}, fmt.Errorf("looking up plan for snapshot: %w", err)
	}
	if found {
		planCategoryID = &pcID
		planAmountSnapshot = &plannedAmount
	}

	var tx Transaction
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO transactions (user_id, category_id, item_id, plan_category_id, amount, plan_amount_snapshot, occurred_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, category_id, item_id, plan_category_id, amount, plan_amount_snapshot, occurred_at`,
		userID, categoryID, itemID, planCategoryID, amount, planAmountSnapshot, occurredAt,
	).Scan(&tx.ID, &tx.CategoryID, &tx.ItemID, &tx.PlanCategoryID, &tx.Amount, &tx.PlanAmountSnapshot, &tx.OccurredAt)
	if err != nil {
		return Transaction{}, fmt.Errorf("creating transaction: %w", err)
	}

	return tx, nil
}

// SumForDate returns the total of all transactions (any category) a
// user recorded on a given calendar date - "actual" spend for the Home
// screen's plan-vs-actual comparison.
func (s *TransactionStore) SumForDate(ctx context.Context, userID, date string) (float64, error) {
	var total float64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND occurred_at::date = $2`,
		userID, date,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("summing transactions for date: %w", err)
	}
	return total, nil
}
