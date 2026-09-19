package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"orbit/backend/internal/store"
)

type fakeHomePlans struct {
	plan  store.DailyPlan
	found bool
}

func (s *fakeHomePlans) GetByDate(ctx context.Context, userID, planDate string) (store.DailyPlan, bool, error) {
	return s.plan, s.found, nil
}

type fakeHomeTransactions struct {
	sum float64
}

func (s *fakeHomeTransactions) SumForDate(ctx context.Context, userID, date string) (float64, error) {
	return s.sum, nil
}

type fakeHomeMood struct {
	entry store.MoodEntry
	found bool
}

func (s *fakeHomeMood) Latest(ctx context.Context, userID string) (store.MoodEntry, bool, error) {
	return s.entry, s.found, nil
}

type fakeHomeBank struct {
	balance float64
}

func (s *fakeHomeBank) CurrentBalance(ctx context.Context, userID string) (float64, error) {
	return s.balance, nil
}

func TestGetHome_FullyComputedSummary(t *testing.T) {
	h := &HomeHandler{
		Plans: &fakeHomePlans{
			found: true,
			plan: store.DailyPlan{
				BaselineAmount: 50000,
				Categories: []store.PlanCategory{
					{PlannedAmount: 30000},
					{PlannedAmount: 16000},
					{PlannedAmount: 15000},
				},
			},
		},
		Transactions: &fakeHomeTransactions{sum: 42000},
		Mood:         &fakeHomeMood{found: true, entry: store.MoodEntry{Mood: "good"}},
		Bank:         &fakeHomeBank{balance: 18000},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/home?date=2026-09-16", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetHome(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}

	var resp homeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if resp.Baseline != 50000 {
		t.Errorf("got baseline %v, want 50000", resp.Baseline)
	}
	if resp.Plan != 61000 {
		t.Errorf("got plan total %v, want 61000 (30000+16000+15000)", resp.Plan)
	}
	if resp.Actual != 42000 {
		t.Errorf("got actual %v, want 42000", resp.Actual)
	}
	if resp.Remaining != 19000 {
		t.Errorf("got remaining %v, want 19000 (plan 61000 - actual 42000)", resp.Remaining)
	}
	if resp.ConsumptionBank.Balance != 18000 || resp.ConsumptionBank.Badge != "" {
		t.Errorf("got bank %+v, want balance 18000, no badge yet (below 100000)", resp.ConsumptionBank)
	}
	if resp.Mood == nil || *resp.Mood != "good" {
		t.Errorf("got mood %v, want \"good\"", resp.Mood)
	}
}

func TestGetHome_NoPlanNoTransactionsNoMood(t *testing.T) {
	// A brand-new user with nothing recorded yet - must not error, just
	// return zeroes/nulls, since none of this is required.
	h := &HomeHandler{
		Plans:        &fakeHomePlans{found: false},
		Transactions: &fakeHomeTransactions{sum: 0},
		Mood:         &fakeHomeMood{found: false},
		Bank:         &fakeHomeBank{balance: 0},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetHome(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp homeResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Baseline != 0 || resp.Plan != 0 || resp.Actual != 0 || resp.Remaining != 0 {
		t.Errorf("got %+v, want all zeroes for a user with nothing recorded", resp)
	}
	if resp.Mood != nil {
		t.Errorf("got mood %v, want nil (never logged)", *resp.Mood)
	}
	if resp.Date == "" {
		t.Error("expected date to default to today when not provided")
	}
}
