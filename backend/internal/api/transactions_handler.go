package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// TransactionStore is the subset of store.TransactionStore this handler
// needs.
type TransactionStore interface {
	Create(ctx context.Context, userID, categoryID string, itemID *string, amount float64, occurredAt time.Time) (store.Transaction, error)
}

// ItemOwnershipChecker is the subset of store.ItemStore needed to verify
// a client-supplied item_id actually belongs to the requesting user.
type ItemOwnershipChecker interface {
	Exists(ctx context.Context, userID, itemID string) (bool, error)
}

// BankReconciler is the subset of store.ConsumptionBankStore needed to
// recompute Consumption Bank credit/reversal right after a transaction
// is recorded.
type BankReconciler interface {
	ReconcileForTransaction(ctx context.Context, userID, categoryID string, occurredAt time.Time, newTransactionID string, plannedAmount float64) error
}

type TransactionsHandler struct {
	Transactions TransactionStore
	Categories   CategoryStore
	Items        ItemOwnershipChecker
	Bank         BankReconciler
}

type createTransactionRequest struct {
	CategoryID string   `json:"category_id"`
	ItemID     *string  `json:"item_id"`
	Amount     *float64 `json:"amount"`
	OccurredAt *string  `json:"occurred_at"` // RFC3339, defaults to now
}

// CreateTransaction handles POST /api/v1/transactions. Never blocks or
// reduces the recorded amount regardless of the plan - it only ever
// snapshots the plan amount in effect at this moment, if any.
func (h *TransactionsHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CategoryID == "" || req.Amount == nil {
		writeJSONError(w, http.StatusBadRequest, "category_id and amount are required")
		return
	}

	owned, err := h.Categories.Exists(r.Context(), userID, req.CategoryID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to verify category")
		return
	}
	if !owned {
		writeJSONError(w, http.StatusNotFound, "category not found")
		return
	}

	if req.ItemID != nil {
		itemOwned, err := h.Items.Exists(r.Context(), userID, *req.ItemID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to verify item")
			return
		}
		if !itemOwned {
			writeJSONError(w, http.StatusNotFound, "item not found")
			return
		}
	}

	occurredAt := time.Now().UTC()
	if req.OccurredAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.OccurredAt)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "occurred_at must be RFC3339")
			return
		}
		occurredAt = parsed
	}

	tx, err := h.Transactions.Create(r.Context(), userID, req.CategoryID, req.ItemID, *req.Amount, occurredAt)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to record transaction")
		return
	}

	// Real-time Consumption Bank reconciliation - only when a plan
	// existed for this category/date (plan_amount_snapshot is the same
	// value used for the transaction's own immutable snapshot). The
	// transaction itself is already recorded at this point and is never
	// rolled back or blocked because of a bank reconciliation problem -
	// a failure here is only logged, the response still reports success.
	if tx.PlanAmountSnapshot != nil {
		// Use the original occurredAt (already UTC, from the request),
		// not tx.OccurredAt as read back from the database - the driver
		// returns timestamptz in the local time zone, which can shift the
		// calendar date near midnight and silently pick the wrong day's
		// plan for reconciliation.
		if err := h.Bank.ReconcileForTransaction(r.Context(), userID, tx.CategoryID, occurredAt, tx.ID, *tx.PlanAmountSnapshot); err != nil {
			log.Printf("consumption bank reconciliation failed for transaction %s: %v", tx.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}
