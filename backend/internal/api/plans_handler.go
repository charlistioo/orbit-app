package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// PlanStore is the subset of store.PlanStore this handler needs.
type PlanStore interface {
	GetByDate(ctx context.Context, userID, planDate string) (store.DailyPlan, bool, error)
	CreateOrReplace(ctx context.Context, userID, planDate string, baselineAmount float64, categoryPlans []store.CategoryPlanInput) (store.DailyPlan, error)
	UpdatePlanCategoryAmount(ctx context.Context, userID, planCategoryID string, newAmount float64) (store.PlanCategory, error)
}

type PlansHandler struct {
	Plans PlanStore
}

type categoryPlanRequest struct {
	CategoryID    string  `json:"category_id"`
	PlannedAmount float64 `json:"planned_amount"`
}

type createPlanRequest struct {
	PlanDate       string                `json:"plan_date"`
	BaselineAmount float64               `json:"baseline_amount"`
	Categories     []categoryPlanRequest `json:"categories"`
}

// CreatePlan handles POST /api/v1/plans - creates the user's daily plan
// for a date, or replaces the baseline/category amounts if one already
// exists. No maximum is enforced on baseline or planned amounts - they
// are soft targets, not hard limits.
func (h *PlansHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlanDate == "" {
		writeJSONError(w, http.StatusBadRequest, "plan_date is required")
		return
	}

	categoryPlans := make([]store.CategoryPlanInput, len(req.Categories))
	for i, c := range req.Categories {
		categoryPlans[i] = store.CategoryPlanInput{CategoryID: c.CategoryID, PlannedAmount: c.PlannedAmount}
	}

	plan, err := h.Plans.CreateOrReplace(r.Context(), userID, req.PlanDate, req.BaselineAmount, categoryPlans)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save plan")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

// GetPlan handles GET /api/v1/plans/{date}.
func (h *PlansHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	date := r.PathValue("date")
	plan, found, err := h.Plans.GetByDate(r.Context(), userID, date)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load plan")
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "no plan for this date")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

type updatePlanCategoryRequest struct {
	PlannedAmount float64 `json:"planned_amount"`
}

// UpdatePlanCategory handles
// PATCH /api/v1/plans/{plan_id}/categories/{plan_category_id} - the
// plan_category_id is the actual authority (scoped to the authenticated
// user via a join), the plan_id in the URL is not separately trusted.
func (h *PlansHandler) UpdatePlanCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	planCategoryID := r.PathValue("plan_category_id")

	var req updatePlanCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "planned_amount is required")
		return
	}

	updated, err := h.Plans.UpdatePlanCategoryAmount(r.Context(), userID, planCategoryID, req.PlannedAmount)
	if errors.Is(err, store.ErrPlanCategoryNotFound) {
		writeJSONError(w, http.StatusNotFound, "plan category not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update plan category")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
