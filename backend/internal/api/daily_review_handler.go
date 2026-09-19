package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/reminder"
	"orbit/backend/internal/store"
)

// DailyReviewStore is the subset of store.DailyReviewStore this handler
// needs.
type DailyReviewStore interface {
	PendingCategories(ctx context.Context, userID, date string) ([]store.PendingCategory, error)
	Finalize(ctx context.Context, userID, date string, status store.DailyReviewStatus) (store.DailyReview, error)
}

type DailyReviewHandler struct {
	Reviews DailyReviewStore
}

type reminderResponse struct {
	Date              string   `json:"date"`
	PendingCategories []string `json:"pending_categories"`
	Message           string   `json:"message"`
}

// GetTodayReminder handles GET /api/v1/reminders/today?date=YYYY-MM-DD
// (defaults to today, UTC). Reports which planned categories have no
// recorded transaction yet for that date, with a neutral reminder
// message - empty when nothing is pending.
func (h *DailyReviewHandler) GetTodayReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	pending, err := h.Reviews.PendingCategories(r.Context(), userID, date)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load reminder state")
		return
	}

	names := make([]string, len(pending))
	messageInput := make([]reminder.CategoryName, len(pending))
	for i, p := range pending {
		names[i] = p.Name
		messageInput[i] = reminder.CategoryName{Name: p.Name}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminderResponse{
		Date:              date,
		PendingCategories: names,
		Message:           reminder.Message(messageInput),
	})
}

type dailyReviewRequest struct {
	Date   string `json:"date"`
	Action string `json:"action"` // "complete" | "skip"
}

type dailyReviewResponse struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Valid  bool   `json:"valid"`
}

// FinalizeDailyReview handles POST /api/v1/daily-review. Finalizes a
// day as either "complete" (reviewed) or "skip" (skipped without
// penalty) - both result in a valid, finalized day. There is no
// separate "penalized" state.
func (h *DailyReviewHandler) FinalizeDailyReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req dailyReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Date == "" {
		writeJSONError(w, http.StatusBadRequest, "date is required (YYYY-MM-DD)")
		return
	}

	var status store.DailyReviewStatus
	switch req.Action {
	case "complete":
		status = store.DailyReviewCompleted
	case "skip":
		status = store.DailyReviewSkipped
	default:
		writeJSONError(w, http.StatusBadRequest, "action must be \"complete\" or \"skip\"")
		return
	}

	review, err := h.Reviews.Finalize(r.Context(), userID, req.Date, status)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to finalize daily review")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dailyReviewResponse{
		Date:   review.ReviewDate,
		Status: string(review.Status),
		Valid:  true,
	})
}
