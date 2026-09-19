package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CalendarDay struct {
	Date     string   `json:"date"`
	Baseline *float64 `json:"baseline"`
	Actual   float64  `json:"actual"`
}

type CalendarMonth struct {
	Month               string        `json:"month"`
	DaysInMonth         int           `json:"days_in_month"`
	Days                []CalendarDay `json:"days"`
	CumulativeBaseline  float64       `json:"cumulative_baseline"`
	CumulativeActual    float64       `json:"cumulative_actual"`
}

type CalendarStore struct {
	db *sql.DB
}

func NewCalendarStore(db *sql.DB) *CalendarStore {
	return &CalendarStore{db: db}
}

// GetMonth returns baseline vs actual for every day in a calendar
// month, plus cumulative totals - automatically using the correct
// number of days for that month (Go's time package already accounts
// for leap years, so no special-casing is needed here: time.Date with
// day 0 of the *next* month rolls back to the last day of *this* one).
func (s *CalendarStore) GetMonth(ctx context.Context, userID string, year, month int) (CalendarMonth, error) {
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	lastOfMonth := time.Date(year, time.Month(month), daysInMonth, 0, 0, 0, 0, time.UTC)

	startStr := firstOfMonth.Format("2006-01-02")
	endStr := lastOfMonth.Format("2006-01-02")

	baselineByDate := map[string]float64{}
	rows, err := s.db.QueryContext(ctx,
		`SELECT plan_date::text, baseline_amount FROM daily_plans
		 WHERE user_id = $1 AND plan_date BETWEEN $2 AND $3`,
		userID, startStr, endStr,
	)
	if err != nil {
		return CalendarMonth{}, fmt.Errorf("querying daily plans for calendar: %w", err)
	}
	for rows.Next() {
		var date string
		var baseline float64
		if err := rows.Scan(&date, &baseline); err != nil {
			rows.Close()
			return CalendarMonth{}, fmt.Errorf("scanning daily plan for calendar: %w", err)
		}
		baselineByDate[date] = baseline
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return CalendarMonth{}, fmt.Errorf("reading daily plans for calendar: %w", err)
	}

	actualByDate := map[string]float64{}
	actualRows, err := s.db.QueryContext(ctx,
		`SELECT occurred_at::date::text AS d, SUM(amount) FROM transactions
		 WHERE user_id = $1 AND occurred_at::date BETWEEN $2 AND $3
		 GROUP BY d`,
		userID, startStr, endStr,
	)
	if err != nil {
		return CalendarMonth{}, fmt.Errorf("querying transactions for calendar: %w", err)
	}
	for actualRows.Next() {
		var date string
		var total float64
		if err := actualRows.Scan(&date, &total); err != nil {
			actualRows.Close()
			return CalendarMonth{}, fmt.Errorf("scanning transaction sum for calendar: %w", err)
		}
		actualByDate[date] = total
	}
	actualRows.Close()
	if err := actualRows.Err(); err != nil {
		return CalendarMonth{}, fmt.Errorf("reading transactions for calendar: %w", err)
	}

	result := CalendarMonth{
		Month:        firstOfMonth.Format("2006-01"),
		DaysInMonth:  daysInMonth,
		Days:         make([]CalendarDay, 0, daysInMonth),
	}

	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		day := CalendarDay{Date: date, Actual: actualByDate[date]}
		if b, ok := baselineByDate[date]; ok {
			bCopy := b
			day.Baseline = &bCopy
			result.CumulativeBaseline += b
		}
		result.CumulativeActual += day.Actual
		result.Days = append(result.Days, day)
	}

	return result, nil
}
