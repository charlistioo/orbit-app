package api

import (
	"context"
	"encoding/json"
	"net/http"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// TimelineStore is the subset of store.TimelineStore this handler
// needs.
type TimelineStore interface {
	ListEvents(ctx context.Context, userID string) ([]store.TimelineEvent, error)
}

type TimelineHandler struct {
	Timeline TimelineStore
}

// GetTransactions handles GET /api/v1/transactions - despite the name
// (matching the approved API design), this returns the merged Timeline:
// transactions, plan adjustments, mood changes, and Consumption Bank
// ledger entries, in one chronologically ordered list.
func (h *TimelineHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	events, err := h.Timeline.ListEvents(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load timeline")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"events": events})
}
