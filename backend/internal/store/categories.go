package store

import (
	"context"
	"database/sql"
	"fmt"
)

type Category struct {
	ID   string
	Name string
}

type CategoryStore struct {
	db *sql.DB
}

func NewCategoryStore(db *sql.DB) *CategoryStore {
	return &CategoryStore{db: db}
}

// List returns the user's categories, most recently created first, for
// quick-select on the next transaction entry.
func (s *CategoryStore) List(ctx context.Context, userID string) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name FROM categories WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// Exists reports whether a category with this id belongs to this user -
// used to reject operations on another user's category id before
// touching items, rather than trusting a client-supplied category_id.
func (s *CategoryStore) Exists(ctx context.Context, userID, categoryID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)`,
		categoryID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking category ownership: %w", err)
	}
	return exists, nil
}

// Create stores a new category with the name exactly as the user typed
// it - no trimming, normalization, or fuzzy matching against existing
// categories.
func (s *CategoryStore) Create(ctx context.Context, userID, name string) (Category, error) {
	var c Category
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO categories (user_id, name) VALUES ($1, $2) RETURNING id, name`,
		userID, name,
	).Scan(&c.ID, &c.Name)
	if err != nil {
		return Category{}, fmt.Errorf("creating category: %w", err)
	}
	return c, nil
}
