package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// InsightStore is the subset of store.InsightStore this handler needs.
type InsightStore interface {
	DailyInsights(ctx context.Context, userID, date string) (store.InsightResult, error)
	WeeklyInsights(ctx context.Context, userID, weekEndDate string) (store.InsightResult, error)
	MonthlyInsights(ctx context.Context, userID, month string) (store.InsightResult, error)
}

type InsightHandler struct {
	Insights InsightStore
}

// GetDaily handles GET /api/v1/insights/daily?date=YYYY-MM-DD (defaults
// to today, UTC).
func (h *InsightHandler) GetDaily(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	result, err := h.Insights.DailyInsights(r.Context(), userID, date)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate insights")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetWeekly handles GET /api/v1/insights/weekly?end=YYYY-MM-DD (defaults
// to today, UTC) - the 7-day window ending on that date.
func (h *InsightHandler) GetWeekly(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	end := r.URL.Query().Get("end")
	if end == "" {
		end = time.Now().UTC().Format("2006-01-02")
	}

	result, err := h.Insights.WeeklyInsights(r.Context(), userID, end)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate insights")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetMonthly handles GET /api/v1/insights/monthly?month=YYYY-MM (defaults
// to the current month, UTC).
func (h *InsightHandler) GetMonthly(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().UTC().Format("2006-01")
	}

	result, err := h.Insights.MonthlyInsights(r.Context(), userID, month)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate insights")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
