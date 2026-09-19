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

type fakeTransactionStore struct {
	lastAmount float64
	createCalls int
}

func (s *fakeTransactionStore) Create(ctx context.Context, userID, categoryID string, itemID *string, amount float64, occurredAt time.Time) (store.Transaction, error) {
	s.createCalls++
	s.lastAmount = amount
	return store.Transaction{ID: "tx-1", CategoryID: categoryID, ItemID: itemID, Amount: amount, OccurredAt: occurredAt}, nil
}

type fakeItemOwnership struct {
	owned map[string]bool
}

func (s *fakeItemOwnership) Exists(ctx context.Context, userID, itemID string) (bool, error) {
	return s.owned[itemID], nil
}

func TestCreateTransaction_ExceedsPlan_StillRecordedInFull(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-1"] = []store.Category{{ID: "cat-1", Name: "Kopi"}}
	txStore := &fakeTransactionStore{}
	h := &TransactionsHandler{
		Transactions: txStore,
		Categories:   categories,
		Items:        &fakeItemOwnership{},
	}

	// Plan was 16000, user spends 26000 - must not be blocked or reduced.
	body, _ := json.Marshal(map[string]any{"category_id": "cat-1", "amount": 26000})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateTransaction(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if txStore.createCalls != 1 {
		t.Fatalf("got %d create calls, want 1", txStore.createCalls)
	}
	if txStore.lastAmount != 26000 {
		t.Errorf("got recorded amount %v, want 26000 (full amount, not reduced)", txStore.lastAmount)
	}
}

func TestCreateTransaction_UnownedCategoryRejected(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-2"] = []store.Category{{ID: "cat-other", Name: "Not yours"}}
	txStore := &fakeTransactionStore{}
	h := &TransactionsHandler{
		Transactions: txStore,
		Categories:   categories,
		Items:        &fakeItemOwnership{},
	}

	body, _ := json.Marshal(map[string]any{"category_id": "cat-other", "amount": 1000})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateTransaction(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
	if txStore.createCalls != 0 {
		t.Errorf("got %d create calls, want 0", txStore.createCalls)
	}
}

func TestCreateTransaction_MissingAmountIsBadRequest(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-1"] = []store.Category{{ID: "cat-1", Name: "Kopi"}}
	txStore := &fakeTransactionStore{}
	h := &TransactionsHandler{Transactions: txStore, Categories: categories, Items: &fakeItemOwnership{}}

	body, _ := json.Marshal(map[string]any{"category_id": "cat-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateTransaction(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateTransaction_UnownedItemRejected(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-1"] = []store.Category{{ID: "cat-1", Name: "Kopi"}}
	txStore := &fakeTransactionStore{}
	h := &TransactionsHandler{
		Transactions: txStore,
		Categories:   categories,
		Items:        &fakeItemOwnership{owned: map[string]bool{}},
	}

	itemID := "item-not-mine"
	body, _ := json.Marshal(map[string]any{"category_id": "cat-1", "amount": 1000, "item_id": itemID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateTransaction(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
	if txStore.createCalls != 0 {
		t.Errorf("got %d create calls, want 0", txStore.createCalls)
	}
}
