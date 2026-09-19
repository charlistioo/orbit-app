package store

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// TestCalendarStore_LeapYearFebruaryUses29Days is the core proof for
// TASK-011's acceptance criteria: a baseline set on every day of
// February in a leap year (2024) must cumulate using 29 days, not 28.
func TestCalendarStore_LeapYearFebruaryUses29Days(t *testing.T) {
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
	plans := NewPlanStore(db)
	calendar := NewCalendarStore(db)

	testUser, err := users.Create(ctx, "integration-calendar-leap-test-google-id", "calendarleap@example.com", "Calendar Leap Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID) })

	const dailyBaseline = 20000.0
	var planIDs []string
	for day := 1; day <= 29; day++ {
		date := february2024Date(day)
		plan, err := plans.CreateOrReplace(ctx, testUser.ID, date, dailyBaseline, nil)
		if err != nil {
			t.Fatalf("CreateOrReplace for %s: %v", date, err)
		}
		planIDs = append(planIDs, plan.ID)
	}
	t.Cleanup(func() {
		for _, id := range planIDs {
			db.ExecContext(ctx, `DELETE FROM daily_plans WHERE id = $1`, id)
		}
	})

	result, err := calendar.GetMonth(ctx, testUser.ID, 2024, 2)
	if err != nil {
		t.Fatalf("GetMonth: %v", err)
	}

	if result.DaysInMonth != 29 {
		t.Fatalf("got DaysInMonth %d, want 29 (2024 is a leap year)", result.DaysInMonth)
	}
	if len(result.Days) != 29 {
		t.Fatalf("got %d days in the response, want 29", len(result.Days))
	}
	wantCumulative := dailyBaseline * 29
	if result.CumulativeBaseline != wantCumulative {
		t.Errorf("got cumulative baseline %v, want %v (20000 x 29 days)", result.CumulativeBaseline, wantCumulative)
	}
}

// TestCalendarStore_NonLeapFebruaryUses28Days is the contrast case -
// the same month/day-of-month logic, one year earlier, where February
// is NOT a leap month.
func TestCalendarStore_NonLeapFebruaryUses28Days(t *testing.T) {
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
	calendar := NewCalendarStore(db)

	testUser, err := users.Create(ctx, "integration-calendar-nonleap-test-google-id", "calendarnonleap@example.com", "Calendar Non-Leap Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID) })

	result, err := calendar.GetMonth(ctx, testUser.ID, 2023, 2)
	if err != nil {
		t.Fatalf("GetMonth: %v", err)
	}
	if result.DaysInMonth != 28 {
		t.Errorf("got DaysInMonth %d, want 28 (2023 is not a leap year)", result.DaysInMonth)
	}
}

func february2024Date(day int) string {
	return fmt.Sprintf("2024-02-%02d", day)
}
