package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// CalendarStore is the subset of store.CalendarStore this handler
// needs.
type CalendarStore interface {
	GetMonth(ctx context.Context, userID string, year, month int) (store.CalendarMonth, error)
}

type CalendarHandler struct {
	Calendar CalendarStore
}

// GetCalendar handles GET /api/v1/calendar?month=YYYY-MM (defaults to
// the current month if omitted). Returns baseline vs actual per date
// plus cumulative totals for the month, with the day count (28/29/30/31)
// determined automatically from the actual calendar - no special-casing
// needed for leap years, Go's time package already knows.
func (h *CalendarHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	monthParam := r.URL.Query().Get("month")
	var year, month int
	if monthParam == "" {
		now := time.Now().UTC()
		year, month = now.Year(), int(now.Month())
	} else {
		parts := strings.Split(monthParam, "-")
		var err error
		if len(parts) != 2 {
			err = errInvalidMonth
		} else {
			year, err = strconv.Atoi(parts[0])
			if err == nil {
				month, err = strconv.Atoi(parts[1])
			}
		}
		if err != nil || month < 1 || month > 12 {
			writeJSONError(w, http.StatusBadRequest, "month must be in YYYY-MM format")
			return
		}
	}

	result, err := h.Calendar.GetMonth(r.Context(), userID, year, month)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load calendar")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

var errInvalidMonth = errors.New("invalid month format")
