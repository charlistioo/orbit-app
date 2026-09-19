package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// MoodStore is the subset of store.MoodStore this handler needs.
type MoodStore interface {
	Create(ctx context.Context, userID, mood string, recordedAt time.Time) (store.MoodEntry, error)
}

type MoodHandler struct {
	Mood MoodStore
}

var validMoods = map[string]bool{
	"very_good": true,
	"good":      true,
	"neutral":   true,
	"stressed":  true,
	"sad":       true,
}

type createMoodRequest struct {
	Mood       string  `json:"mood"`
	RecordedAt *string `json:"recorded_at"` // RFC3339, defaults to now
}

// CreateMood handles POST /api/v1/mood. Mood can be logged any number
// of times in a day - each call creates its own timestamped entry.
func (h *MoodHandler) CreateMood(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createMoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !validMoods[req.Mood] {
		writeJSONError(w, http.StatusBadRequest, "mood must be one of: very_good, good, neutral, stressed, sad")
		return
	}

	recordedAt := time.Now().UTC()
	if req.RecordedAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.RecordedAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "recorded_at must be RFC3339")
			return
		}
		recordedAt = parsed
	}

	entry, err := h.Mood.Create(r.Context(), userID, req.Mood, recordedAt)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to log mood")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}
