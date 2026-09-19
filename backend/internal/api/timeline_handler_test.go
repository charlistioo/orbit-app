package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"orbit/backend/internal/store"
)

type fakeTimelineStore struct {
	events []store.TimelineEvent
}

func (s *fakeTimelineStore) ListEvents(ctx context.Context, userID string) ([]store.TimelineEvent, error) {
	return s.events, nil
}

func TestGetTransactions_ReturnsEventsAsGiven(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2026-09-16T09:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2026-09-16T10:00:00Z")

	fake := &fakeTimelineStore{events: []store.TimelineEvent{
		{Type: "mood", OccurredAt: t1, Mood: &store.MoodEntry{Mood: "good"}},
		{Type: "transaction", OccurredAt: t2, Transaction: &store.TimelineTransaction{Amount: 12000}},
	}}
	h := &TimelineHandler{Timeline: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetTransactions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Events []store.TimelineEvent `json:"events"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Events) != 2 {
		t.Fatalf("got %d events, want 2", len(resp.Events))
	}
	if resp.Events[0].Type != "mood" || resp.Events[1].Type != "transaction" {
		t.Errorf("got event types %q, %q - order should be preserved as returned by the store", resp.Events[0].Type, resp.Events[1].Type)
	}
}
