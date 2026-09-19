package api

import (
	"context"
	"encoding/json"
	"net/http"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// CategoryStore is the subset of store.CategoryStore this handler needs.
type CategoryStore interface {
	List(ctx context.Context, userID string) ([]store.Category, error)
	Create(ctx context.Context, userID, name string) (store.Category, error)
	Exists(ctx context.Context, userID, categoryID string) (bool, error)
}

// ItemStore is the subset of store.ItemStore this handler needs.
type ItemStore interface {
	ListByCategory(ctx context.Context, userID, categoryID string) ([]store.Item, error)
	Create(ctx context.Context, userID, categoryID, name string) (store.Item, error)
}

type CategoriesHandler struct {
	Categories CategoryStore
	Items      ItemStore
}

type createCategoryRequest struct {
	Name string `json:"name"`
}

// ListCategories handles GET /api/v1/categories - the user's own
// categories, most recently created first.
func (h *CategoriesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	categories, err := h.Categories.List(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"categories": categories})
}

// CreateCategory handles POST /api/v1/categories - stores the name
// exactly as the user typed it, no normalization.
func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	category, err := h.Categories.Create(r.Context(), userID, req.Name)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// ListItems handles GET /api/v1/categories/{category_id}/items.
func (h *CategoriesHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	categoryID := r.PathValue("category_id")

	owned, err := h.Categories.Exists(r.Context(), userID, categoryID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to verify category")
		return
	}
	if !owned {
		writeJSONError(w, http.StatusNotFound, "category not found")
		return
	}

	items, err := h.Items.ListByCategory(r.Context(), userID, categoryID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"items": items})
}

type createItemRequest struct {
	Name string `json:"name"`
}

// CreateItem handles POST /api/v1/categories/{category_id}/items -
// stores the name exactly as the user typed it, no normalization.
func (h *CategoriesHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	categoryID := r.PathValue("category_id")

	owned, err := h.Categories.Exists(r.Context(), userID, categoryID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to verify category")
		return
	}
	if !owned {
		writeJSONError(w, http.StatusNotFound, "category not found")
		return
	}

	var req createItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	item, err := h.Items.Create(r.Context(), userID, categoryID, req.Name)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}
