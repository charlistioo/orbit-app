package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"orbit/backend/internal/store"
)

type fakeMoodStore struct {
	createCalls int
	lastMood    string
}

func (s *fakeMoodStore) Create(ctx context.Context, userID, mood string, recordedAt time.Time) (store.MoodEntry, error) {
	s.createCalls++
	s.lastMood = mood
	return store.MoodEntry{ID: "mood-1", Mood: mood, RecordedAt: recordedAt}, nil
}

func TestCreateMood_ValidMoodAccepted(t *testing.T) {
	moods := &fakeMoodStore{}
	h := &MoodHandler{Mood: moods}

	body, _ := json.Marshal(map[string]string{"mood": "stressed"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mood", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateMood(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if moods.lastMood != "stressed" {
		t.Errorf("got mood %q, want stressed", moods.lastMood)
	}
}

func TestCreateMood_InvalidMoodRejected(t *testing.T) {
	moods := &fakeMoodStore{}
	h := &MoodHandler{Mood: moods}

	body, _ := json.Marshal(map[string]string{"mood": "ecstatic"}) // not one of the 5 allowed values
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mood", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateMood(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if moods.createCalls != 0 {
		t.Errorf("got %d create calls, want 0", moods.createCalls)
	}
}

func TestCreateMood_MultipleEntriesSameDayAllowed(t *testing.T) {
	moods := &fakeMoodStore{}
	h := &MoodHandler{Mood: moods}

	for _, m := range []string{"good", "neutral", "stressed"} {
		body, _ := json.Marshal(map[string]string{"mood": m})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mood", strings.NewReader(string(body)))
		req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
		rec := httptest.NewRecorder()
		h.CreateMood(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("mood %q: got status %d", m, rec.Code)
		}
	}
	if moods.createCalls != 3 {
		t.Errorf("got %d create calls, want 3 (each mood change logged independently)", moods.createCalls)
	}
}
