package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// BankStore is the subset of store.ConsumptionBankStore this handler
// needs.
type BankStore interface {
	CurrentBalance(ctx context.Context, userID string) (float64, error)
	Ledger(ctx context.Context, userID string) ([]store.LedgerEntry, error)
	Apply(ctx context.Context, userID, transactionID string, requestedAmount *float64) (store.LedgerEntry, error)
}

type ConsumptionBankHandler struct {
	Bank BankStore
}

// badgeTierFor returns the badge name for a Consumption Bank balance,
// or "" if it hasn't reached the first tier yet.
func badgeTierFor(balance float64) string {
	switch {
	case balance >= 500000:
		return "Master"
	case balance >= 400000:
		return "Collector"
	case balance >= 300000:
		return "Economical"
	case balance >= 200000:
		return "Thrifty"
	case balance >= 100000:
		return "Frugal"
	default:
		return ""
	}
}

// GetBank handles GET /api/v1/consumption-bank - current balance, badge
// tier, and the full ledger.
func (h *ConsumptionBankHandler) GetBank(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	balance, err := h.Bank.CurrentBalance(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read balance")
		return
	}
	ledger, err := h.Bank.Ledger(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read ledger")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"balance": balance,
		"badge":   badgeTierFor(balance),
		"ledger":  ledger,
	})
}

type applyBankRequest struct {
	TransactionID string   `json:"transaction_id"`
	Amount        *float64 `json:"amount"`
}

// ApplyBank handles POST /api/v1/consumption-bank/apply - covers part
// of an overspend transaction using the bank balance, capped at 10% of
// the balance at the time of the request and at the remaining overspend
// on that transaction, never taking the balance below 0. This never
// alters the transaction's own recorded amount - it only ever appends a
// separate ledger entry.
func (h *ConsumptionBankHandler) ApplyBank(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req applyBankRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TransactionID == "" {
		writeJSONError(w, http.StatusBadRequest, "transaction_id is required")
		return
	}

	entry, err := h.Bank.Apply(r.Context(), userID, req.TransactionID, req.Amount)
	if errors.Is(err, store.ErrTransactionNotFound) {
		writeJSONError(w, http.StatusNotFound, "transaction not found")
		return
	}
	if errors.Is(err, store.ErrNothingToApply) {
		writeJSONError(w, http.StatusUnprocessableEntity, "nothing to apply - no remaining overspend or zero balance")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to apply Consumption Bank")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}
