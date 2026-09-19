package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

type PlanAdjustmentEvent struct {
	ID             string
	PlanCategoryID string
	CategoryID     string
	CategoryName   string
	OldAmount      float64
	NewAmount      float64
	ChangedAt      time.Time
}

// TimelineTransaction is a transaction event's payload - like
// store.Transaction, but with the category/item display names joined
// in (the mobile Timeline shows names, never raw ids).
type TimelineTransaction struct {
	ID                 string
	CategoryID         string
	CategoryName       string
	ItemID             *string
	ItemName           *string
	PlanCategoryID     *string
	Amount             float64
	PlanAmountSnapshot *float64
	OccurredAt         time.Time
}

// TimelineEvent is one entry in the merged, chronological timeline.
// Exactly one of the typed fields is set, matching Type.
type TimelineEvent struct {
	Type           string                `json:"type"` // transaction | plan_adjustment | mood | bank_ledger
	OccurredAt     time.Time             `json:"occurred_at"`
	Transaction    *TimelineTransaction  `json:"transaction,omitempty"`
	PlanAdjustment *PlanAdjustmentEvent  `json:"plan_adjustment,omitempty"`
	Mood           *MoodEntry            `json:"mood,omitempty"`
	BankLedger     *LedgerEntry          `json:"bank_ledger,omitempty"`
}

type TimelineStore struct {
	db *sql.DB
}

func NewTimelineStore(db *sql.DB) *TimelineStore {
	return &TimelineStore{db: db}
}

// ListEvents merges transactions, plan_adjustments, mood_entries, and
// consumption_bank_ledger for one user into a single list, ordered
// chronologically (oldest first) by each event's own timestamp.
func (s *TimelineStore) ListEvents(ctx context.Context, userID string) ([]TimelineEvent, error) {
	events := []TimelineEvent{}

	txRows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.category_id, c.name, t.item_id, i.name, t.plan_category_id, t.amount, t.plan_amount_snapshot, t.occurred_at
		 FROM transactions t
		 JOIN categories c ON c.id = t.category_id
		 LEFT JOIN items i ON i.id = t.item_id
		 WHERE t.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying transactions for timeline: %w", err)
	}
	for txRows.Next() {
		var tx TimelineTransaction
		if err := txRows.Scan(&tx.ID, &tx.CategoryID, &tx.CategoryName, &tx.ItemID, &tx.ItemName, &tx.PlanCategoryID, &tx.Amount, &tx.PlanAmountSnapshot, &tx.OccurredAt); err != nil {
			txRows.Close()
			return nil, fmt.Errorf("scanning transaction for timeline: %w", err)
		}
		txCopy := tx
		events = append(events, TimelineEvent{Type: "transaction", OccurredAt: tx.OccurredAt, Transaction: &txCopy})
	}
	txRows.Close()
	if err := txRows.Err(); err != nil {
		return nil, fmt.Errorf("reading transactions for timeline: %w", err)
	}

	adjRows, err := s.db.QueryContext(ctx,
		`SELECT pa.id, pa.plan_category_id, pc.category_id, c.name, pa.old_amount, pa.new_amount, pa.changed_at
		 FROM plan_adjustments pa
		 JOIN plan_categories pc ON pc.id = pa.plan_category_id
		 JOIN daily_plans dp ON dp.id = pc.daily_plan_id
		 JOIN categories c ON c.id = pc.category_id
		 WHERE dp.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying plan adjustments for timeline: %w", err)
	}
	for adjRows.Next() {
		var a PlanAdjustmentEvent
		if err := adjRows.Scan(&a.ID, &a.PlanCategoryID, &a.CategoryID, &a.CategoryName, &a.OldAmount, &a.NewAmount, &a.ChangedAt); err != nil {
			adjRows.Close()
			return nil, fmt.Errorf("scanning plan adjustment for timeline: %w", err)
		}
		aCopy := a
		events = append(events, TimelineEvent{Type: "plan_adjustment", OccurredAt: a.ChangedAt, PlanAdjustment: &aCopy})
	}
	adjRows.Close()
	if err := adjRows.Err(); err != nil {
		return nil, fmt.Errorf("reading plan adjustments for timeline: %w", err)
	}

	moodRows, err := s.db.QueryContext(ctx,
		`SELECT id, mood, recorded_at FROM mood_entries WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying mood entries for timeline: %w", err)
	}
	for moodRows.Next() {
		var m MoodEntry
		if err := moodRows.Scan(&m.ID, &m.Mood, &m.RecordedAt); err != nil {
			moodRows.Close()
			return nil, fmt.Errorf("scanning mood entry for timeline: %w", err)
		}
		mCopy := m
		events = append(events, TimelineEvent{Type: "mood", OccurredAt: m.RecordedAt, Mood: &mCopy})
	}
	moodRows.Close()
	if err := moodRows.Err(); err != nil {
		return nil, fmt.Errorf("reading mood entries for timeline: %w", err)
	}

	ledgerRows, err := s.db.QueryContext(ctx,
		`SELECT id, delta_amount, reason, related_transaction_id, balance_after, created_at
		 FROM consumption_bank_ledger WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying bank ledger for timeline: %w", err)
	}
	for ledgerRows.Next() {
		var l LedgerEntry
		if err := ledgerRows.Scan(&l.ID, &l.DeltaAmount, &l.Reason, &l.RelatedTransactionID, &l.BalanceAfter, &l.CreatedAt); err != nil {
			ledgerRows.Close()
			return nil, fmt.Errorf("scanning bank ledger entry for timeline: %w", err)
		}
		lCopy := l
		events = append(events, TimelineEvent{Type: "bank_ledger", OccurredAt: l.CreatedAt, BankLedger: &lCopy})
	}
	ledgerRows.Close()
	if err := ledgerRows.Err(); err != nil {
		return nil, fmt.Errorf("reading bank ledger for timeline: %w", err)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].OccurredAt.Before(events[j].OccurredAt)
	})

	return events, nil
}
