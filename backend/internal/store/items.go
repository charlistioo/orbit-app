package store

import (
	"context"
	"database/sql"
	"fmt"
)

type Item struct {
	ID         string
	CategoryID string
	Name       string
}

type ItemStore struct {
	db *sql.DB
}

func NewItemStore(db *sql.DB) *ItemStore {
	return &ItemStore{db: db}
}

// ListByCategory returns the user's items under one category, most
// recently created first, for quick-select on the next transaction
// entry.
func (s *ItemStore) ListByCategory(ctx context.Context, userID, categoryID string) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, category_id, name FROM items
		 WHERE user_id = $1 AND category_id = $2
		 ORDER BY created_at DESC`,
		userID, categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing items: %w", err)
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.CategoryID, &it.Name); err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// Exists reports whether an item belongs to this user (regardless of
// category) - used to reject a client-supplied item_id that doesn't
// actually belong to them before attaching it to a transaction.
func (s *ItemStore) Exists(ctx context.Context, userID, itemID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM items WHERE id = $1 AND user_id = $2)`,
		itemID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking item ownership: %w", err)
	}
	return exists, nil
}

// Create stores a new item with the name exactly as the user typed it -
// no trimming, normalization, or fuzzy matching against existing items.
func (s *ItemStore) Create(ctx context.Context, userID, categoryID, name string) (Item, error) {
	var it Item
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO items (user_id, category_id, name) VALUES ($1, $2, $3)
		 RETURNING id, category_id, name`,
		userID, categoryID, name,
	).Scan(&it.ID, &it.CategoryID, &it.Name)
	if err != nil {
		return Item{}, fmt.Errorf("creating item: %w", err)
	}
	return it, nil
}
