package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

func contextWithTestUser(ctx context.Context, userID string) context.Context {
	return auth.ContextWithUserID(ctx, userID)
}

type fakeCategoryStore struct {
	byUser      map[string][]store.Category
	createCalls int
	lastCreatedName string
}

func newFakeCategoryStore() *fakeCategoryStore {
	return &fakeCategoryStore{byUser: map[string][]store.Category{}}
}

func (s *fakeCategoryStore) List(ctx context.Context, userID string) ([]store.Category, error) {
	return s.byUser[userID], nil
}

func (s *fakeCategoryStore) Create(ctx context.Context, userID, name string) (store.Category, error) {
	s.createCalls++
	s.lastCreatedName = name
	c := store.Category{ID: "cat-1", Name: name}
	s.byUser[userID] = append([]store.Category{c}, s.byUser[userID]...)
	return c, nil
}

func (s *fakeCategoryStore) Exists(ctx context.Context, userID, categoryID string) (bool, error) {
	for _, c := range s.byUser[userID] {
		if c.ID == categoryID {
			return true, nil
		}
	}
	return false, nil
}

type fakeItemStore struct {
	byCategory  map[string][]store.Item
	createCalls int
}

func newFakeItemStore() *fakeItemStore {
	return &fakeItemStore{byCategory: map[string][]store.Item{}}
}

func (s *fakeItemStore) ListByCategory(ctx context.Context, userID, categoryID string) ([]store.Item, error) {
	return s.byCategory[categoryID], nil
}

func (s *fakeItemStore) Create(ctx context.Context, userID, categoryID, name string) (store.Item, error) {
	s.createCalls++
	it := store.Item{ID: "item-1", CategoryID: categoryID, Name: name}
	s.byCategory[categoryID] = append([]store.Item{it}, s.byCategory[categoryID]...)
	return it, nil
}

func TestCreateCategory_StoresNameExactlyAsTyped(t *testing.T) {
	categories := newFakeCategoryStore()
	h := &CategoriesHandler{Categories: categories, Items: newFakeItemStore()}

	weirdName := "  Kopi Susu!! "
	body, _ := json.Marshal(map[string]string{"name": weirdName})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(string(body)))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if categories.lastCreatedName != weirdName {
		t.Errorf("got stored name %q, want exactly %q (no trimming/normalization)", categories.lastCreatedName, weirdName)
	}
}

func TestListCategories_MostRecentFirst(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-1"] = []store.Category{
		{ID: "cat-2", Name: "Newer"},
		{ID: "cat-1", Name: "Older"},
	}
	h := &CategoriesHandler{Categories: categories, Items: newFakeItemStore()}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ListCategories(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Categories []store.Category `json:"categories"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Categories) != 2 || resp.Categories[0].Name != "Newer" {
		t.Errorf("got %+v, want Newer first", resp.Categories)
	}
}

func TestListCategories_ScopedToAuthenticatedUser(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-1"] = []store.Category{{ID: "cat-1", Name: "User1 category"}}
	categories.byUser["user-2"] = []store.Category{{ID: "cat-2", Name: "User2 category"}}
	h := &CategoriesHandler{Categories: categories, Items: newFakeItemStore()}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ListCategories(rec, req)

	var resp struct {
		Categories []store.Category `json:"categories"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Categories) != 1 || resp.Categories[0].Name != "User1 category" {
		t.Errorf("got %+v, want only user-1's category", resp.Categories)
	}
}

func TestCreateItem_RejectsCategoryBelongingToAnotherUser(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-2"] = []store.Category{{ID: "cat-owned-by-user2", Name: "Not yours"}}
	items := newFakeItemStore()
	h := &CategoriesHandler{Categories: categories, Items: items}

	body, _ := json.Marshal(map[string]string{"name": "Sneaky item"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/cat-owned-by-user2/items", strings.NewReader(string(body)))
	req.SetPathValue("category_id", "cat-owned-by-user2")
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateItem(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d (category not owned by this user)", rec.Code, http.StatusNotFound)
	}
	if items.createCalls != 0 {
		t.Errorf("got %d item creates, want 0", items.createCalls)
	}
}

func TestListItems_MostRecentFirst(t *testing.T) {
	categories := newFakeCategoryStore()
	categories.byUser["user-1"] = []store.Category{{ID: "cat-1", Name: "Kopi"}}
	items := newFakeItemStore()
	items.byCategory["cat-1"] = []store.Item{
		{ID: "item-2", CategoryID: "cat-1", Name: "Newer item"},
		{ID: "item-1", CategoryID: "cat-1", Name: "Older item"},
	}
	h := &CategoriesHandler{Categories: categories, Items: items}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/cat-1/items", nil)
	req.SetPathValue("category_id", "cat-1")
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ListItems(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []store.Item `json:"items"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Items) != 2 || resp.Items[0].Name != "Newer item" {
		t.Errorf("got %+v, want Newer item first", resp.Items)
	}
}
