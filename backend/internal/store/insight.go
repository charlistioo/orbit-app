package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"orbit/backend/internal/insight"
)

type InsightResult struct {
	Period   string   `json:"period"` // daily | weekly | monthly
	Date     string   `json:"date"`
	Insights []string `json:"insights"`
}

type InsightStore struct {
	db *sql.DB
}

func NewInsightStore(db *sql.DB) *InsightStore {
	return &InsightStore{db: db}
}

// DailyInsights generates insight text for a single date.
func (s *InsightStore) DailyInsights(ctx context.Context, userID, date string) (InsightResult, error) {
	in, err := s.buildInputs(ctx, userID, "daily", date, date)
	if err != nil {
		return InsightResult{}, err
	}
	return InsightResult{Period: "daily", Date: date, Insights: insight.Generate(in)}, nil
}

// WeeklyInsights generates insight text for the 7-day window ending on
// (and including) weekEndDate.
func (s *InsightStore) WeeklyInsights(ctx context.Context, userID, weekEndDate string) (InsightResult, error) {
	end, err := time.Parse("2006-01-02", weekEndDate)
	if err != nil {
		return InsightResult{}, fmt.Errorf("parsing week end date: %w", err)
	}
	start := end.AddDate(0, 0, -6).Format("2006-01-02")

	in, err := s.buildInputs(ctx, userID, "weekly", start, weekEndDate)
	if err != nil {
		return InsightResult{}, err
	}
	return InsightResult{Period: "weekly", Date: weekEndDate, Insights: insight.Generate(in)}, nil
}

// MonthlyInsights generates insight text for a calendar month ("YYYY-MM").
func (s *InsightStore) MonthlyInsights(ctx context.Context, userID, month string) (InsightResult, error) {
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return InsightResult{}, fmt.Errorf("parsing month: %w", err)
	}
	monthEnd := monthStart.AddDate(0, 1, -1)
	start := monthStart.Format("2006-01-02")
	end := monthEnd.Format("2006-01-02")

	in, err := s.buildInputs(ctx, userID, "monthly", start, end)
	if err != nil {
		return InsightResult{}, err
	}
	return InsightResult{Period: "monthly", Date: month, Insights: insight.Generate(in)}, nil
}

// buildInputs fetches every aggregate insight.Generate needs for the
// [startDate, endDate] window and assembles them into insight.Inputs. The
// trend window and history-day count always look at the account's whole
// history up to endDate, independent of how wide [startDate, endDate]
// itself is - a weekly or monthly report still must not claim a trend for
// an account with under 3 total days of history.
func (s *InsightStore) buildInputs(ctx context.Context, userID, period, startDate, endDate string) (insight.Inputs, error) {
	planTotal, actualTotal, err := s.planVsActualForRange(ctx, userID, startDate, endDate)
	if err != nil {
		return insight.Inputs{}, err
	}

	historyDays, err := s.distinctTransactionDayCount(ctx, userID, endDate)
	if err != nil {
		return insight.Inputs{}, err
	}

	trendWindow, err := s.trendWindow(ctx, userID, endDate)
	if err != nil {
		return insight.Inputs{}, err
	}

	moodBeforeSpending, err := s.moodBeforeSpendingInRange(ctx, userID, startDate, endDate)
	if err != nil {
		return insight.Inputs{}, err
	}

	bankCredited, bankApplied, err := s.bankMovementForRange(ctx, userID, startDate, endDate)
	if err != nil {
		return insight.Inputs{}, err
	}

	categoryFrequency, err := s.categoryFrequencyForRange(ctx, userID, startDate, endDate)
	if err != nil {
		return insight.Inputs{}, err
	}

	return insight.Inputs{
		Period:              period,
		PlanTotal:           planTotal,
		ActualTotal:         actualTotal,
		DistinctHistoryDays: historyDays,
		TrendWindow:         trendWindow,
		MoodBeforeSpending:  moodBeforeSpending,
		BankCredited:        bankCredited,
		BankApplied:         bankApplied,
		CategoryFrequency:   categoryFrequency,
	}, nil
}

func (s *InsightStore) planVsActualForRange(ctx context.Context, userID, startDate, endDate string) (planTotal, actualTotal float64, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(pc.planned_amount), 0)
		 FROM plan_categories pc
		 JOIN daily_plans dp ON dp.id = pc.daily_plan_id
		 WHERE dp.user_id = $1 AND dp.plan_date BETWEEN $2 AND $3`,
		userID, startDate, endDate,
	).Scan(&planTotal)
	if err != nil {
		return 0, 0, fmt.Errorf("summing plan for range: %w", err)
	}

	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE user_id = $1 AND occurred_at::date BETWEEN $2 AND $3`,
		userID, startDate, endDate,
	).Scan(&actualTotal)
	if err != nil {
		return 0, 0, fmt.Errorf("summing actual for range: %w", err)
	}
	return planTotal, actualTotal, nil
}

// distinctTransactionDayCount reports how many distinct calendar dates
// (up to and including untilDate) have at least one transaction - used
// to decide whether there is enough history to make any multi-day trend
// claim at all. A user with only 1 day of data must never get one.
func (s *InsightStore) distinctTransactionDayCount(ctx context.Context, userID, untilDate string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT occurred_at::date) FROM transactions
		 WHERE user_id = $1 AND occurred_at::date <= $2`,
		userID, untilDate,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting distinct transaction days: %w", err)
	}
	return count, nil
}

// categoryTotalsForDate returns category-name -> total spend for one date.
func (s *InsightStore) categoryTotalsForDate(ctx context.Context, userID, date string) (insight.DayCategoryTotals, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.name, SUM(t.amount)
		 FROM transactions t
		 JOIN categories c ON c.id = t.category_id
		 WHERE t.user_id = $1 AND t.occurred_at::date = $2
		 GROUP BY c.name`,
		userID, date,
	)
	if err != nil {
		return nil, fmt.Errorf("querying category totals for date: %w", err)
	}
	defer rows.Close()

	totals := insight.DayCategoryTotals{}
	for rows.Next() {
		var name string
		var total float64
		if err := rows.Scan(&name, &total); err != nil {
			return nil, fmt.Errorf("scanning category total: %w", err)
		}
		totals[name] = total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(totals) == 0 {
		return nil, nil
	}
	return totals, nil
}

// trendWindow returns the 3 most recent distinct days' category totals up
// to and including endDate, oldest first - the raw material for the
// "3 consecutive rising days" rule.
func (s *InsightStore) trendWindow(ctx context.Context, userID, endDate string) ([3]insight.DayCategoryTotals, error) {
	var window [3]insight.DayCategoryTotals

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return window, fmt.Errorf("parsing end date: %w", err)
	}

	for i, offset := range []int{-2, -1, 0} {
		date := end.AddDate(0, 0, offset).Format("2006-01-02")
		totals, err := s.categoryTotalsForDate(ctx, userID, date)
		if err != nil {
			return window, err
		}
		window[i] = totals
	}
	return window, nil
}

// bankMovementForRange sums Consumption Bank ledger deltas whose related
// transaction occurred within the range, split into credited (positive,
// from auto-reconcile entries) and applied (negative, from apply
// entries) - each returned as a single non-negative total. This joins
// through the related transaction's occurred_at rather than the ledger
// row's own created_at, since a reconciliation can be written well after
// the transaction it's for (e.g. a backfilled/historical entry) and it's
// the transaction's date the report is organized around.
func (s *InsightStore) bankMovementForRange(ctx context.Context, userID, startDate, endDate string) (credited, applied float64, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(cbl.delta_amount) FILTER (WHERE cbl.delta_amount > 0 AND cbl.reason LIKE $4), 0),
			COALESCE(-SUM(cbl.delta_amount) FILTER (WHERE cbl.delta_amount < 0 AND cbl.reason LIKE $5), 0)
		 FROM consumption_bank_ledger cbl
		 JOIN transactions t ON t.id = cbl.related_transaction_id
		 WHERE cbl.user_id = $1 AND t.occurred_at::date BETWEEN $2 AND $3`,
		userID, startDate, endDate, autoReconcileReasonPrefix+"%", applyReasonPrefix+"%",
	).Scan(&credited, &applied)
	if err != nil {
		return 0, 0, fmt.Errorf("summing bank movement for range: %w", err)
	}
	return credited, applied, nil
}

// categoryFrequencyForRange returns category name -> number of
// transactions within the range - the raw material for the
// "most frequently used category" insight.
func (s *InsightStore) categoryFrequencyForRange(ctx context.Context, userID, startDate, endDate string) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.name, COUNT(*)
		 FROM transactions t
		 JOIN categories c ON c.id = t.category_id
		 WHERE t.user_id = $1 AND t.occurred_at::date BETWEEN $2 AND $3
		 GROUP BY c.name`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("querying category frequency for range: %w", err)
	}
	defer rows.Close()

	frequency := map[string]int{}
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, fmt.Errorf("scanning category frequency: %w", err)
		}
		frequency[name] = count
	}
	return frequency, rows.Err()
}

// moodBeforeSpendingInRange reports whether any mood entry in the range
// was logged earlier in its day than a transaction on that same day -
// used only to phrase a *temporal* observation, never a causal claim.
func (s *InsightStore) moodBeforeSpendingInRange(ctx context.Context, userID, startDate, endDate string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM mood_entries m
			WHERE m.user_id = $1 AND m.recorded_at::date BETWEEN $2 AND $3
			AND EXISTS (
				SELECT 1 FROM transactions t
				WHERE t.user_id = $1 AND t.occurred_at::date = m.recorded_at::date AND t.occurred_at > m.recorded_at
			)
		)`,
		userID, startDate, endDate,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking mood-then-spending correlation: %w", err)
	}
	return exists, nil
}
