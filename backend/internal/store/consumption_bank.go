package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type LedgerEntry struct {
	ID                   string
	DeltaAmount          float64
	Reason               string
	RelatedTransactionID *string
	BalanceAfter         float64
	CreatedAt            time.Time
}

var (
	ErrTransactionNotFound = errors.New("transaction not found for this user")
	ErrNothingToApply      = errors.New("no overspend remains to cover, or Consumption Bank balance is 0")
)

type ConsumptionBankStore struct {
	db *sql.DB
}

func NewConsumptionBankStore(db *sql.DB) *ConsumptionBankStore {
	return &ConsumptionBankStore{db: db}
}

// CurrentBalance returns the user's Consumption Bank balance - the
// balance_after of their most recent ledger entry, or 0 if they have
// none yet.
func (s *ConsumptionBankStore) CurrentBalance(ctx context.Context, userID string) (float64, error) {
	var balance float64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE((SELECT balance_after FROM consumption_bank_ledger WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1), 0)`,
		userID,
	).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("reading current balance: %w", err)
	}
	return balance, nil
}

// Ledger returns the user's full Consumption Bank ledger, oldest first -
// the source of truth for how the balance was reached.
func (s *ConsumptionBankStore) Ledger(ctx context.Context, userID string) ([]LedgerEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, delta_amount, reason, related_transaction_id, balance_after, created_at
		 FROM consumption_bank_ledger WHERE user_id = $1 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing ledger: %w", err)
	}
	defer rows.Close()

	entries := []LedgerEntry{}
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.DeltaAmount, &e.Reason, &e.RelatedTransactionID, &e.BalanceAfter, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning ledger entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

const autoReconcileReasonPrefix = "auto-reconcile: "
const applyReasonPrefix = "apply: "

func clamp0(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

// appendLedgerEntry writes one ledger row, flooring the resulting
// balance at 0 (per-user, never below zero, even if that means writing
// a smaller delta than requested). Serializes concurrent writers for
// the same user with an advisory lock, so two simultaneous requests
// can't both read the same stale balance and together push it negative
// (the race condition flagged in the Phase 2 tech-stack risk review).
func appendLedgerEntry(ctx context.Context, tx *sql.Tx, userID string, requestedDelta float64, reason string, relatedTransactionID *string) (LedgerEntry, error) {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, userID); err != nil {
		return LedgerEntry{}, fmt.Errorf("acquiring per-user ledger lock: %w", err)
	}

	var currentBalance float64
	err := tx.QueryRowContext(ctx,
		`SELECT COALESCE((SELECT balance_after FROM consumption_bank_ledger WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1), 0)`,
		userID,
	).Scan(&currentBalance)
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("reading current balance: %w", err)
	}

	actualDelta := requestedDelta
	if actualDelta < 0 && currentBalance+actualDelta < 0 {
		actualDelta = -currentBalance
	}
	newBalance := currentBalance + actualDelta

	var e LedgerEntry
	err = tx.QueryRowContext(ctx,
		`INSERT INTO consumption_bank_ledger (user_id, delta_amount, reason, related_transaction_id, balance_after)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, delta_amount, reason, related_transaction_id, balance_after, created_at`,
		userID, actualDelta, reason, relatedTransactionID, newBalance,
	).Scan(&e.ID, &e.DeltaAmount, &e.Reason, &e.RelatedTransactionID, &e.BalanceAfter, &e.CreatedAt)
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("inserting ledger entry: %w", err)
	}
	return e, nil
}

// ReconcileForTransaction is called right after a transaction is
// recorded for a category that has a plan. It recomputes "how much of
// today's planned budget for this category is still unspent" and
// appends the incremental change (positive = newly-recognized
// underspend credit, negative = a reversal because more has now been
// spent than was previously credited) as one new ledger entry - never
// editing a prior entry, per the append-only ledger design.
//
// This is real-time and self-correcting: crediting a category's
// leftover as each transaction comes in, and clawing part of it back
// (down to, but never causing the overall balance to go below zero) if
// a later transaction in the same category/day eats into that leftover.
func (s *ConsumptionBankStore) ReconcileForTransaction(ctx context.Context, userID, categoryID string, occurredAt time.Time, newTransactionID string, plannedAmount float64) error {
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting reconcile transaction: %w", err)
	}
	defer dbTx.Rollback()

	planDate := occurredAt.UTC().Format("2006-01-02")

	var cumulativeBefore, cumulativeAfter float64
	err = dbTx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE user_id = $1 AND category_id = $2 AND occurred_at::date = $3 AND id != $4`,
		userID, categoryID, planDate, newTransactionID,
	).Scan(&cumulativeBefore)
	if err != nil {
		return fmt.Errorf("summing prior transactions: %w", err)
	}

	var thisAmount float64
	err = dbTx.QueryRowContext(ctx, `SELECT amount FROM transactions WHERE id = $1`, newTransactionID).Scan(&thisAmount)
	if err != nil {
		return fmt.Errorf("reading new transaction amount: %w", err)
	}
	cumulativeAfter = cumulativeBefore + thisAmount

	var recordedBefore float64
	err = dbTx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(cbl.delta_amount), 0)
		 FROM consumption_bank_ledger cbl
		 JOIN transactions t ON t.id = cbl.related_transaction_id
		 WHERE t.user_id = $1 AND t.category_id = $2 AND t.occurred_at::date = $3
		   AND cbl.reason LIKE $4`,
		userID, categoryID, planDate, autoReconcileReasonPrefix+"%",
	).Scan(&recordedBefore)
	if err != nil {
		return fmt.Errorf("summing prior reconciliation entries: %w", err)
	}

	recordedAfter := clamp0(plannedAmount - cumulativeAfter)
	delta := recordedAfter - recordedBefore

	if delta != 0 {
		reason := autoReconcileReasonPrefix + fmt.Sprintf("category %s on %s", categoryID, planDate)
		if _, err := appendLedgerEntry(ctx, dbTx, userID, delta, reason, &newTransactionID); err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

// Apply lets the user cover part of an overspend transaction using
// their Consumption Bank balance. The transaction's own recorded amount
// is never touched - this only ever appends a separate ledger entry.
// requestedAmount, if nil, applies the maximum allowed.
func (s *ConsumptionBankStore) Apply(ctx context.Context, userID, transactionID string, requestedAmount *float64) (LedgerEntry, error) {
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("starting apply transaction: %w", err)
	}
	defer dbTx.Rollback()

	var amount float64
	var planSnapshot sql.NullFloat64
	err = dbTx.QueryRowContext(ctx,
		`SELECT amount, plan_amount_snapshot FROM transactions WHERE id = $1 AND user_id = $2`,
		transactionID, userID,
	).Scan(&amount, &planSnapshot)
	if errors.Is(err, sql.ErrNoRows) {
		return LedgerEntry{}, ErrTransactionNotFound
	}
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("looking up transaction: %w", err)
	}

	if !planSnapshot.Valid || amount <= planSnapshot.Float64 {
		return LedgerEntry{}, ErrNothingToApply
	}
	overspend := amount - planSnapshot.Float64

	var alreadyApplied float64
	err = dbTx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(-delta_amount), 0) FROM consumption_bank_ledger
		 WHERE related_transaction_id = $1 AND reason LIKE $2`,
		transactionID, applyReasonPrefix+"%",
	).Scan(&alreadyApplied)
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("summing prior applies: %w", err)
	}
	remainingOverspend := overspend - alreadyApplied
	if remainingOverspend <= 0 {
		return LedgerEntry{}, ErrNothingToApply
	}

	if _, err := dbTx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, userID); err != nil {
		return LedgerEntry{}, fmt.Errorf("acquiring per-user ledger lock: %w", err)
	}
	var currentBalance float64
	err = dbTx.QueryRowContext(ctx,
		`SELECT COALESCE((SELECT balance_after FROM consumption_bank_ledger WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1), 0)`,
		userID,
	).Scan(&currentBalance)
	if err != nil {
		return LedgerEntry{}, fmt.Errorf("reading current balance: %w", err)
	}

	maxAllowed := 0.10 * currentBalance
	if remainingOverspend < maxAllowed {
		maxAllowed = remainingOverspend
	}
	if maxAllowed <= 0 {
		return LedgerEntry{}, ErrNothingToApply
	}

	toApply := maxAllowed
	if requestedAmount != nil {
		toApply = *requestedAmount
		if toApply > maxAllowed {
			toApply = maxAllowed
		}
		if toApply <= 0 {
			return LedgerEntry{}, ErrNothingToApply
		}
	}

	reason := applyReasonPrefix + fmt.Sprintf("overspend on transaction %s", transactionID)
	entry, err := appendLedgerEntry(ctx, dbTx, userID, -toApply, reason, &transactionID)
	if err != nil {
		return LedgerEntry{}, err
	}

	if err := dbTx.Commit(); err != nil {
		return LedgerEntry{}, fmt.Errorf("committing apply: %w", err)
	}
	return entry, nil
}
