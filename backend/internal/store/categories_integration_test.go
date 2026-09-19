package store

import (
	"context"
	"os"
	"testing"
)

// TestCategoryAndItemStore_Integration runs against a real PostgreSQL
// database (migrations already applied). Skipped unless DATABASE_URL is
// set. See TASK-003's users_integration_test.go for the same pattern.
func TestCategoryAndItemStore_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set - skipping integration test")
	}

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	users := NewUserStore(db)
	categories := NewCategoryStore(db)
	items := NewItemStore(db)

	// A category/item belongs to a user, so create one first.
	testUser, err := users.Create(ctx, "integration-cat-test-google-id", "cattest@example.com", "Cat Test")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, testUser.ID)

	const weirdName = "  Kopi Susu!! "
	cat, err := categories.Create(ctx, testUser.ID, weirdName)
	if err != nil {
		t.Fatalf("creating category: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, cat.ID)

	if cat.Name != weirdName {
		t.Errorf("got stored name %q, want exactly %q (no normalization)", cat.Name, weirdName)
	}

	owned, err := categories.Exists(ctx, testUser.ID, cat.ID)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !owned {
		t.Fatal("expected the creating user to own the category")
	}

	list, err := categories.List(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].ID != cat.ID {
		t.Errorf("got %+v, want exactly the one category just created", list)
	}

	item, err := items.Create(ctx, testUser.ID, cat.ID, "Kopi Susu Gula Aren")
	if err != nil {
		t.Fatalf("creating item: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM items WHERE id = $1`, item.ID)

	itemList, err := items.ListByCategory(ctx, testUser.ID, cat.ID)
	if err != nil {
		t.Fatalf("ListByCategory: %v", err)
	}
	if len(itemList) != 1 || itemList[0].Name != "Kopi Susu Gula Aren" {
		t.Errorf("got %+v, want exactly the one item just created", itemList)
	}
}
