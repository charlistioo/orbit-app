package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PlanCategory struct {
	ID             string
	CategoryID     string
	PlannedAmount  float64
}

type DailyPlan struct {
	ID              string
	PlanDate        string // YYYY-MM-DD
	BaselineAmount  float64
	Categories      []PlanCategory
}

type CategoryPlanInput struct {
	CategoryID    string
	PlannedAmount float64
}

var ErrPlanCategoryNotFound = errors.New("plan category not found for this user")

type PlanStore struct {
	db *sql.DB
}

func NewPlanStore(db *sql.DB) *PlanStore {
	return &PlanStore{db: db}
}

// GetByDate returns the user's daily plan (baseline + per-category
// planned amounts) for a date, or found=false if none exists yet.
func (s *PlanStore) GetByDate(ctx context.Context, userID, planDate string) (DailyPlan, bool, error) {
	var plan DailyPlan
	err := s.db.QueryRowContext(ctx,
		`SELECT id, plan_date::text, baseline_amount FROM daily_plans WHERE user_id = $1 AND plan_date = $2`,
		userID, planDate,
	).Scan(&plan.ID, &plan.PlanDate, &plan.BaselineAmount)

	if errors.Is(err, sql.ErrNoRows) {
		return DailyPlan{}, false, nil
	}
	if err != nil {
		return DailyPlan{}, false, fmt.Errorf("querying daily plan: %w", err)
	}

	categories, err := s.categoriesForPlan(ctx, plan.ID)
	if err != nil {
		return DailyPlan{}, false, err
	}
	plan.Categories = categories

	return plan, true, nil
}

func (s *PlanStore) categoriesForPlan(ctx context.Context, dailyPlanID string) ([]PlanCategory, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, category_id, planned_amount FROM plan_categories WHERE daily_plan_id = $1`,
		dailyPlanID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying plan categories: %w", err)
	}
	defer rows.Close()

	categories := []PlanCategory{}
	for rows.Next() {
		var pc PlanCategory
		if err := rows.Scan(&pc.ID, &pc.CategoryID, &pc.PlannedAmount); err != nil {
			return nil, fmt.Errorf("scanning plan category: %w", err)
		}
		categories = append(categories, pc)
	}
	return categories, rows.Err()
}

// CreateOrReplace creates the user's daily plan for a date (baseline +
// per-category planned amounts), or, if one already exists for that
// date, updates the baseline and per-category amounts in place. This is
// the initial "set my plan for today" call - PATCH is the mechanism for
// editing a single category afterwards and is what records
// plan_adjustments; a fresh create/replace here does not, since there is
// no prior in-effect value being adjusted for a category that either
// didn't exist yet or is being replaced wholesale.
func (s *PlanStore) CreateOrReplace(ctx context.Context, userID, planDate string, baselineAmount float64, categoryPlans []CategoryPlanInput) (DailyPlan, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DailyPlan{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	var planID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO daily_plans (user_id, plan_date, baseline_amount)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, plan_date)
		 DO UPDATE SET baseline_amount = EXCLUDED.baseline_amount
		 RETURNING id`,
		userID, planDate, baselineAmount,
	).Scan(&planID)
	if err != nil {
		return DailyPlan{}, fmt.Errorf("upserting daily plan: %w", err)
	}

	for _, cp := range categoryPlans {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO plan_categories (daily_plan_id, category_id, planned_amount)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (daily_plan_id, category_id)
			 DO UPDATE SET planned_amount = EXCLUDED.planned_amount, updated_at = now()`,
			planID, cp.CategoryID, cp.PlannedAmount,
		)
		if err != nil {
			return DailyPlan{}, fmt.Errorf("inserting plan category: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return DailyPlan{}, fmt.Errorf("committing plan: %w", err)
	}

	return s.mustGetByID(ctx, planID, planDate)
}

func (s *PlanStore) mustGetByID(ctx context.Context, planID, planDate string) (DailyPlan, error) {
	var plan DailyPlan
	err := s.db.QueryRowContext(ctx,
		`SELECT id, plan_date::text, baseline_amount FROM daily_plans WHERE id = $1`,
		planID,
	).Scan(&plan.ID, &plan.PlanDate, &plan.BaselineAmount)
	if err != nil {
		return DailyPlan{}, fmt.Errorf("re-reading daily plan: %w", err)
	}

	categories, err := s.categoriesForPlan(ctx, planID)
	if err != nil {
		return DailyPlan{}, err
	}
	plan.Categories = categories
	return plan, nil
}

// FindPlanCategoryForDate returns the plan category (and its currently
// planned amount) for a user's category on a given date, if a plan
// exists for that date at all - used to snapshot "the plan in effect
// right now" onto a new transaction. found=false (not an error) if the
// user simply has no plan for that date/category, which is a normal,
// allowed state (plans are optional).
func (s *PlanStore) FindPlanCategoryForDate(ctx context.Context, userID, categoryID, planDate string) (planCategoryID string, plannedAmount float64, found bool, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT pc.id, pc.planned_amount
		 FROM plan_categories pc
		 JOIN daily_plans dp ON dp.id = pc.daily_plan_id
		 WHERE dp.user_id = $1 AND dp.plan_date = $2 AND pc.category_id = $3`,
		userID, planDate, categoryID,
	).Scan(&planCategoryID, &plannedAmount)

	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, false, nil
	}
	if err != nil {
		return "", 0, false, fmt.Errorf("finding plan category for date: %w", err)
	}
	return planCategoryID, plannedAmount, true, nil
}

// UpdatePlanCategoryAmount changes a single category's planned amount at
// any time, scoped to the authenticated user (verified via a join
// through daily_plans), and records the old/new value plus a timestamp
// in plan_adjustments in the same transaction. There is no maximum
// enforced on the new amount - baselines and plans are soft targets.
func (s *PlanStore) UpdatePlanCategoryAmount(ctx context.Context, userID, planCategoryID string, newAmount float64) (PlanCategory, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PlanCategory{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	var pc PlanCategory
	var oldAmount float64
	err = tx.QueryRowContext(ctx,
		`SELECT pc.id, pc.category_id, pc.planned_amount
		 FROM plan_categories pc
		 JOIN daily_plans dp ON dp.id = pc.daily_plan_id
		 WHERE pc.id = $1 AND dp.user_id = $2
		 FOR UPDATE`,
		planCategoryID, userID,
	).Scan(&pc.ID, &pc.CategoryID, &oldAmount)

	if errors.Is(err, sql.ErrNoRows) {
		return PlanCategory{}, ErrPlanCategoryNotFound
	}
	if err != nil {
		return PlanCategory{}, fmt.Errorf("looking up plan category: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE plan_categories SET planned_amount = $1, updated_at = $2 WHERE id = $3`,
		newAmount, time.Now(), planCategoryID,
	)
	if err != nil {
		return PlanCategory{}, fmt.Errorf("updating plan category: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO plan_adjustments (plan_category_id, old_amount, new_amount) VALUES ($1, $2, $3)`,
		planCategoryID, oldAmount, newAmount,
	)
	if err != nil {
		return PlanCategory{}, fmt.Errorf("recording plan adjustment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return PlanCategory{}, fmt.Errorf("committing update: %w", err)
	}

	pc.PlannedAmount = newAmount
	return pc, nil
}
