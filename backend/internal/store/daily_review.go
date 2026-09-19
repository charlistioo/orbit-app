package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PendingCategory is a planned category that has no recorded transaction
// yet for a given date - the raw material for a neutral reminder, never
// phrased as something the user failed to do.
type PendingCategory struct {
	CategoryID    string
	Name          string
	PlannedAmount float64
}

// DailyReviewStatus is either "completed" (the user reviewed the day) or
// "skipped" (the user chose not to review it) - both are equally valid,
// final states for the day. There is no penalty state: skipping never
// marks the day invalid or flags it negatively.
type DailyReviewStatus string

const (
	DailyReviewCompleted DailyReviewStatus = "completed"
	DailyReviewSkipped   DailyReviewStatus = "skipped"
)

type DailyReview struct {
	ID         string
	ReviewDate string
	Status     DailyReviewStatus
}

var ErrInvalidReviewStatus = errors.New("status must be \"completed\" or \"skipped\"")

type DailyReviewStore struct {
	db *sql.DB
}

func NewDailyReviewStore(db *sql.DB) *DailyReviewStore {
	return &DailyReviewStore{db: db}
}

// PendingCategories returns the planned categories for a date that have
// no recorded transaction yet - what a reminder should list. A category
// with no plan at all for the date is not "pending" (plans are optional,
// there's nothing to remind about).
func (s *DailyReviewStore) PendingCategories(ctx context.Context, userID, date string) ([]PendingCategory, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.name, pc.planned_amount
		 FROM plan_categories pc
		 JOIN daily_plans dp ON dp.id = pc.daily_plan_id
		 JOIN categories c ON c.id = pc.category_id
		 WHERE dp.user_id = $1 AND dp.plan_date = $2
		 AND NOT EXISTS (
		     SELECT 1 FROM transactions t
		     WHERE t.user_id = dp.user_id AND t.category_id = pc.category_id AND t.occurred_at::date = dp.plan_date
		 )
		 ORDER BY c.name`,
		userID, date,
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending categories: %w", err)
	}
	defer rows.Close()

	pending := []PendingCategory{}
	for rows.Next() {
		var p PendingCategory
		if err := rows.Scan(&p.CategoryID, &p.Name, &p.PlannedAmount); err != nil {
			return nil, fmt.Errorf("scanning pending category: %w", err)
		}
		pending = append(pending, p)
	}
	return pending, rows.Err()
}

// Finalize records the user's daily review decision for a date -
// "completed" or "skipped" - creating the record if none exists yet, or
// overwriting a prior decision for the same date otherwise. Skipping is
// a normal, valid way to finalize a day: it is stored as its own status,
// never as an absence of review or a penalty marker.
func (s *DailyReviewStore) Finalize(ctx context.Context, userID, date string, status DailyReviewStatus) (DailyReview, error) {
	if status != DailyReviewCompleted && status != DailyReviewSkipped {
		return DailyReview{}, ErrInvalidReviewStatus
	}

	var review DailyReview
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO daily_reviews (user_id, review_date, status)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, review_date)
		 DO UPDATE SET status = EXCLUDED.status
		 RETURNING id, review_date::text, status`,
		userID, date, string(status),
	).Scan(&review.ID, &review.ReviewDate, &review.Status)
	if err != nil {
		return DailyReview{}, fmt.Errorf("upserting daily review: %w", err)
	}
	return review, nil
}

// GetByDate returns the user's review decision for a date, if any.
func (s *DailyReviewStore) GetByDate(ctx context.Context, userID, date string) (DailyReview, bool, error) {
	var review DailyReview
	err := s.db.QueryRowContext(ctx,
		`SELECT id, review_date::text, status FROM daily_reviews WHERE user_id = $1 AND review_date = $2`,
		userID, date,
	).Scan(&review.ID, &review.ReviewDate, &review.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return DailyReview{}, false, nil
	}
	if err != nil {
		return DailyReview{}, false, fmt.Errorf("querying daily review: %w", err)
	}
	return review, true, nil
}
