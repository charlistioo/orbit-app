package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"orbit/backend/internal/store"
)

type fakeCalendarStore struct {
	result store.CalendarMonth
}

func (s *fakeCalendarStore) GetMonth(ctx context.Context, userID string, year, month int) (store.CalendarMonth, error) {
	return s.result, nil
}

func TestGetCalendar_ParsesMonthParam(t *testing.T) {
	fake := &fakeCalendarStore{result: store.CalendarMonth{Month: "2026-02", DaysInMonth: 28}}
	h := &CalendarHandler{Calendar: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/calendar?month=2026-02", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetCalendar(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp store.CalendarMonth
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Month != "2026-02" || resp.DaysInMonth != 28 {
		t.Errorf("got %+v", resp)
	}
}

func TestGetCalendar_InvalidMonthIsBadRequest(t *testing.T) {
	h := &CalendarHandler{Calendar: &fakeCalendarStore{}}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/calendar?month=not-a-month", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetCalendar(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
