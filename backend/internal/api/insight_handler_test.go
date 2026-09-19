package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

type fakeInsightStore struct {
	lastUserID string
	lastArg    string
	result     store.InsightResult
}

func (f *fakeInsightStore) DailyInsights(ctx context.Context, userID, date string) (store.InsightResult, error) {
	f.lastUserID = userID
	f.lastArg = date
	return f.result, nil
}

func (f *fakeInsightStore) WeeklyInsights(ctx context.Context, userID, weekEndDate string) (store.InsightResult, error) {
	f.lastUserID = userID
	f.lastArg = weekEndDate
	return f.result, nil
}

func (f *fakeInsightStore) MonthlyInsights(ctx context.Context, userID, month string) (store.InsightResult, error) {
	f.lastUserID = userID
	f.lastArg = month
	return f.result, nil
}

func withAuthedContext(req *http.Request, userID string) *http.Request {
	return req.WithContext(auth.ContextWithUserID(req.Context(), userID))
}

func TestInsightHandler_GetDaily_PassesDateThrough(t *testing.T) {
	fake := &fakeInsightStore{result: store.InsightResult{Period: "daily", Date: "2026-09-16", Insights: []string{"ok"}}}
	h := &InsightHandler{Insights: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/daily?date=2026-09-16", nil)
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.GetDaily(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if fake.lastArg != "2026-09-16" {
		t.Fatalf("expected date passed through, got %q", fake.lastArg)
	}
	if fake.lastUserID != "user-1" {
		t.Fatalf("expected authenticated user id passed through, got %q", fake.lastUserID)
	}
}

func TestInsightHandler_GetWeekly_Unauthenticated(t *testing.T) {
	fake := &fakeInsightStore{}
	h := &InsightHandler{Insights: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/weekly", nil)
	rec := httptest.NewRecorder()

	h.GetWeekly(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth context, got %d", rec.Code)
	}
}

func TestInsightHandler_GetMonthly_DefaultsToCurrentMonth(t *testing.T) {
	fake := &fakeInsightStore{result: store.InsightResult{Period: "monthly"}}
	h := &InsightHandler{Insights: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/insights/monthly", nil)
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.GetMonthly(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if fake.lastArg == "" {
		t.Fatalf("expected a default month to be passed, got empty string")
	}
}
