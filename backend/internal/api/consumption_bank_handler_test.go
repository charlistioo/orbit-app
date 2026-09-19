package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"orbit/backend/internal/store"
)

type fakeBankStore struct {
	balance      float64
	ledger       []store.LedgerEntry
	applyErr     error
	lastRequested *float64
	applyCalls   int
}

func (s *fakeBankStore) CurrentBalance(ctx context.Context, userID string) (float64, error) {
	return s.balance, nil
}

func (s *fakeBankStore) Ledger(ctx context.Context, userID string) ([]store.LedgerEntry, error) {
	return s.ledger, nil
}

func (s *fakeBankStore) Apply(ctx context.Context, userID, transactionID string, requestedAmount *float64) (store.LedgerEntry, error) {
	s.applyCalls++
	s.lastRequested = requestedAmount
	if s.applyErr != nil {
		return store.LedgerEntry{}, s.applyErr
	}
	return store.LedgerEntry{ID: "ledger-1", DeltaAmount: -1000, BalanceAfter: s.balance - 1000}, nil
}

func TestBadgeTierFor(t *testing.T) {
	cases := []struct {
		balance float64
		want    string
	}{
		{0, ""},
		{99999, ""},
		{100000, "Frugal"},
		{199999, "Frugal"},
		{200000, "Thrifty"},
		{300000, "Economical"},
		{400000, "Collector"},
		{500000, "Master"},
		{1000000, "Master"},
	}
	for _, c := range cases {
		got := badgeTierFor(c.balance)
		if got != c.want {
			t.Errorf("badgeTierFor(%v) = %q, want %q", c.balance, got, c.want)
		}
	}
}

func TestGetBank_ReturnsBalanceBadgeAndLedger(t *testing.T) {
	bank := &fakeBankStore{balance: 250000, ledger: []store.LedgerEntry{{ID: "l1", DeltaAmount: 4000}}}
	h := &ConsumptionBankHandler{Bank: bank}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/consumption-bank", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetBank(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Balance float64            `json:"balance"`
		Badge   string             `json:"badge"`
		Ledger  []store.LedgerEntry `json:"ledger"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Balance != 250000 || resp.Badge != "Thrifty" || len(resp.Ledger) != 1 {
		t.Errorf("got %+v", resp)
	}
}

func TestApplyBank_MissingTransactionIDIsBadRequest(t *testing.T) {
	bank := &fakeBankStore{}
	h := &ConsumptionBankHandler{Bank: bank}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/consumption-bank/apply", strings.NewReader(`{}`))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ApplyBank(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if bank.applyCalls != 0 {
		t.Errorf("got %d apply calls, want 0", bank.applyCalls)
	}
}

func TestApplyBank_NothingToApplyIs422(t *testing.T) {
	bank := &fakeBankStore{applyErr: store.ErrNothingToApply}
	h := &ConsumptionBankHandler{Bank: bank}

	body, _ := json.Marshal(map[string]string{"transaction_id": "tx-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/consumption-bank/apply", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ApplyBank(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestApplyBank_Success(t *testing.T) {
	bank := &fakeBankStore{balance: 10000}
	h := &ConsumptionBankHandler{Bank: bank}

	body, _ := json.Marshal(map[string]string{"transaction_id": "tx-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/consumption-bank/apply", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ApplyBank(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if bank.applyCalls != 1 {
		t.Errorf("got %d apply calls, want 1", bank.applyCalls)
	}
}
