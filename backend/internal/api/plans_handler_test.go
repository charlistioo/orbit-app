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

type planCategoryRecord struct {
	OldAmount, NewAmount float64
}

type fakePlanStore struct {
	plansByUserDate map[string]store.DailyPlan // key: userID+"|"+date
	planCategories  map[string]store.PlanCategory
	adjustments     []planCategoryRecord
	createCalls     int
}

func newFakePlanStore() *fakePlanStore {
	return &fakePlanStore{
		plansByUserDate: map[string]store.DailyPlan{},
		planCategories:  map[string]store.PlanCategory{},
	}
}

func key(userID, date string) string { return userID + "|" + date }

func (s *fakePlanStore) GetByDate(ctx context.Context, userID, planDate string) (store.DailyPlan, bool, error) {
	p, ok := s.plansByUserDate[key(userID, planDate)]
	return p, ok, nil
}

func (s *fakePlanStore) CreateOrReplace(ctx context.Context, userID, planDate string, baselineAmount float64, categoryPlans []store.CategoryPlanInput) (store.DailyPlan, error) {
	s.createCalls++
	var cats []store.PlanCategory
	for i, cp := range categoryPlans {
		pcID := planDate + "-pc-" + string(rune('a'+i))
		pc := store.PlanCategory{ID: pcID, CategoryID: cp.CategoryID, PlannedAmount: cp.PlannedAmount}
		cats = append(cats, pc)
		s.planCategories[pcID] = pc
	}
	plan := store.DailyPlan{ID: planDate + "-plan", PlanDate: planDate, BaselineAmount: baselineAmount, Categories: cats}
	s.plansByUserDate[key(userID, planDate)] = plan
	return plan, nil
}

func (s *fakePlanStore) UpdatePlanCategoryAmount(ctx context.Context, userID, planCategoryID string, newAmount float64) (store.PlanCategory, error) {
	pc, ok := s.planCategories[planCategoryID]
	if !ok {
		return store.PlanCategory{}, store.ErrPlanCategoryNotFound
	}
	s.adjustments = append(s.adjustments, planCategoryRecord{OldAmount: pc.PlannedAmount, NewAmount: newAmount})
	pc.PlannedAmount = newAmount
	s.planCategories[planCategoryID] = pc
	return pc, nil
}

func TestCreatePlan_NoMaximumEnforced(t *testing.T) {
	plans := newFakePlanStore()
	h := &PlansHandler{Plans: plans}

	// An enormous baseline and category amount - should not be rejected.
	body, _ := json.Marshal(map[string]any{
		"plan_date":       "2026-09-16",
		"baseline_amount": 999999999.99,
		"categories":      []map[string]any{{"category_id": "cat-1", "planned_amount": 500000000.0}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/plans", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreatePlan(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, body %s - large amounts should never be rejected (soft targets only)", rec.Code, rec.Body.String())
	}
}

func TestUpdatePlanCategory_RecordsAdjustment(t *testing.T) {
	plans := newFakePlanStore()
	plans.planCategories["pc-1"] = store.PlanCategory{ID: "pc-1", CategoryID: "cat-1", PlannedAmount: 16000}
	h := &PlansHandler{Plans: plans}

	body, _ := json.Marshal(map[string]float64{"planned_amount": 25000})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plans/plan-1/categories/pc-1", strings.NewReader(string(body)))
	req.SetPathValue("plan_id", "plan-1")
	req.SetPathValue("plan_category_id", "pc-1")
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.UpdatePlanCategory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if len(plans.adjustments) != 1 {
		t.Fatalf("got %d adjustments recorded, want 1", len(plans.adjustments))
	}
	adj := plans.adjustments[0]
	if adj.OldAmount != 16000 || adj.NewAmount != 25000 {
		t.Errorf("got adjustment %+v, want old=16000 new=25000", adj)
	}
}

func TestUpdatePlanCategory_UnknownIDIs404(t *testing.T) {
	plans := newFakePlanStore()
	h := &PlansHandler{Plans: plans}

	body, _ := json.Marshal(map[string]float64{"planned_amount": 1000})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/plans/plan-1/categories/nonexistent", strings.NewReader(string(body)))
	req.SetPathValue("plan_id", "plan-1")
	req.SetPathValue("plan_category_id", "nonexistent")
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.UpdatePlanCategory(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetPlan_NotFoundForUnknownDate(t *testing.T) {
	plans := newFakePlanStore()
	h := &PlansHandler{Plans: plans}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plans/2026-01-01", nil)
	req.SetPathValue("date", "2026-01-01")
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetPlan(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}
