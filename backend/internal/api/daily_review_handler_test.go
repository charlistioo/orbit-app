package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"orbit/backend/internal/store"
)

type fakeDailyReviewStore struct {
	pending      []store.PendingCategory
	finalizeErr  error
	lastStatus   store.DailyReviewStatus
	lastDate     string
	finalizeCall bool
}

func (f *fakeDailyReviewStore) PendingCategories(ctx context.Context, userID, date string) ([]store.PendingCategory, error) {
	return f.pending, nil
}

func (f *fakeDailyReviewStore) Finalize(ctx context.Context, userID, date string, status store.DailyReviewStatus) (store.DailyReview, error) {
	f.finalizeCall = true
	f.lastStatus = status
	f.lastDate = date
	if f.finalizeErr != nil {
		return store.DailyReview{}, f.finalizeErr
	}
	return store.DailyReview{ID: "review-1", ReviewDate: date, Status: status}, nil
}

func TestGetTodayReminder_NoPendingIsEmptyMessage(t *testing.T) {
	fake := &fakeDailyReviewStore{pending: nil}
	h := &DailyReviewHandler{Reviews: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reminders/today?date=2026-09-16", nil)
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.GetTodayReminder(rec, req)

	var resp reminderResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Message != "" {
		t.Fatalf("expected empty message with nothing pending, got: %q", resp.Message)
	}
	if len(resp.PendingCategories) != 0 {
		t.Fatalf("expected no pending categories, got: %v", resp.PendingCategories)
	}
}

func TestGetTodayReminder_ListsPendingCategoriesWithNeutralMessage(t *testing.T) {
	fake := &fakeDailyReviewStore{pending: []store.PendingCategory{
		{CategoryID: "cat-1", Name: "Kopi", PlannedAmount: 20000},
	}}
	h := &DailyReviewHandler{Reviews: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reminders/today", nil)
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.GetTodayReminder(rec, req)

	var resp reminderResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(resp.PendingCategories) != 1 || resp.PendingCategories[0] != "Kopi" {
		t.Fatalf("expected [Kopi], got: %v", resp.PendingCategories)
	}
	if resp.Message == "" {
		t.Fatalf("expected a non-empty reminder message when something is pending")
	}
}

func TestGetTodayReminder_Unauthenticated(t *testing.T) {
	fake := &fakeDailyReviewStore{}
	h := &DailyReviewHandler{Reviews: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reminders/today", nil)
	rec := httptest.NewRecorder()

	h.GetTodayReminder(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// TestFinalizeDailyReview_SkipIsValidNoPenalty is the direct proof of
// TASK-013's first acceptance criterion: choosing "Lewati" (skip) still
// finalizes the day as valid, with no separate penalty state.
func TestFinalizeDailyReview_SkipIsValidNoPenalty(t *testing.T) {
	fake := &fakeDailyReviewStore{}
	h := &DailyReviewHandler{Reviews: fake}

	body, _ := json.Marshal(map[string]string{"date": "2026-09-16", "action": "skip"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-review", bytes.NewReader(body))
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.FinalizeDailyReview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp dailyReviewResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Status != "skipped" {
		t.Fatalf("got status %q, want skipped", resp.Status)
	}
	if !resp.Valid {
		t.Fatalf("expected valid=true for a skipped day, got false (skip must never be treated as invalid/penalized)")
	}
	if fake.lastStatus != store.DailyReviewSkipped {
		t.Fatalf("expected store called with DailyReviewSkipped, got %v", fake.lastStatus)
	}
}

func TestFinalizeDailyReview_CompleteIsValid(t *testing.T) {
	fake := &fakeDailyReviewStore{}
	h := &DailyReviewHandler{Reviews: fake}

	body, _ := json.Marshal(map[string]string{"date": "2026-09-16", "action": "complete"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-review", bytes.NewReader(body))
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.FinalizeDailyReview(rec, req)

	var resp dailyReviewResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Status != "completed" || !resp.Valid {
		t.Fatalf("got status=%q valid=%v, want completed/true", resp.Status, resp.Valid)
	}
}

func TestFinalizeDailyReview_InvalidActionIsBadRequest(t *testing.T) {
	fake := &fakeDailyReviewStore{}
	h := &DailyReviewHandler{Reviews: fake}

	body, _ := json.Marshal(map[string]string{"date": "2026-09-16", "action": "something-else"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-review", bytes.NewReader(body))
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.FinalizeDailyReview(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if fake.finalizeCall {
		t.Fatalf("store should not be called for an invalid action")
	}
}

func TestFinalizeDailyReview_MissingDateIsBadRequest(t *testing.T) {
	fake := &fakeDailyReviewStore{}
	h := &DailyReviewHandler{Reviews: fake}

	body, _ := json.Marshal(map[string]string{"action": "skip"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-review", bytes.NewReader(body))
	req = withAuthedContext(req, "user-1")
	rec := httptest.NewRecorder()

	h.FinalizeDailyReview(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestFinalizeDailyReview_Unauthenticated(t *testing.T) {
	fake := &fakeDailyReviewStore{}
	h := &DailyReviewHandler{Reviews: fake}

	body, _ := json.Marshal(map[string]string{"date": "2026-09-16", "action": "skip"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/daily-review", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.FinalizeDailyReview(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
