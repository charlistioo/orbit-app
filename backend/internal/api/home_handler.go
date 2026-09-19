package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// HomePlanStore is the subset of store.PlanStore this handler needs.
type HomePlanStore interface {
	GetByDate(ctx context.Context, userID, planDate string) (store.DailyPlan, bool, error)
}

// HomeTransactionStore is the subset of store.TransactionStore this
// handler needs.
type HomeTransactionStore interface {
	SumForDate(ctx context.Context, userID, date string) (float64, error)
}

// HomeMoodStore is the subset of store.MoodStore this handler needs.
type HomeMoodStore interface {
	Latest(ctx context.Context, userID string) (store.MoodEntry, bool, error)
}

// HomeBankStore is the subset of store.ConsumptionBankStore this
// handler needs.
type HomeBankStore interface {
	CurrentBalance(ctx context.Context, userID string) (float64, error)
}

type HomeHandler struct {
	Plans        HomePlanStore
	Transactions HomeTransactionStore
	Mood         HomeMoodStore
	Bank         HomeBankStore
}

type homeResponse struct {
	Date            string  `json:"date"`
	Baseline        float64 `json:"baseline"`
	Plan            float64 `json:"plan"`
	Actual          float64 `json:"actual"`
	Remaining       float64 `json:"remaining"`
	ConsumptionBank struct {
		Balance float64 `json:"balance"`
		Badge   string  `json:"badge"`
	} `json:"consumption_bank"`
	Mood *string `json:"mood"`
}

// GetHome handles GET /api/v1/home?date=YYYY-MM-DD (date optional,
// defaults to today in UTC). Returns one fully computed summary - the
// mobile client does no aggregation of its own, per the acceptance
// criteria: baseline, plan total, actual total, remaining (plan -
// actual, matching the product's own worked examples), Consumption Bank
// balance/badge, and the user's most recently logged mood.
func (h *HomeHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	var resp homeResponse
	resp.Date = date

	plan, found, err := h.Plans.GetByDate(r.Context(), userID, date)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load plan")
		return
	}
	if found {
		resp.Baseline = plan.BaselineAmount
		for _, c := range plan.Categories {
			resp.Plan += c.PlannedAmount
		}
	}

	actual, err := h.Transactions.SumForDate(r.Context(), userID, date)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to sum transactions")
		return
	}
	resp.Actual = actual
	resp.Remaining = resp.Plan - resp.Actual

	balance, err := h.Bank.CurrentBalance(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read Consumption Bank balance")
		return
	}
	resp.ConsumptionBank.Balance = balance
	resp.ConsumptionBank.Badge = badgeTierFor(balance)

	moodEntry, moodFound, err := h.Mood.Latest(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read latest mood")
		return
	}
	if moodFound {
		resp.Mood = &moodEntry.Mood
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
