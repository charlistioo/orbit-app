// Package store holds database access for ORBIT's entities.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	ID                    string
	GoogleID              string
	Email                 string
	DisplayName           string
	Age                   *int
	AvatarData            *string
	DefaultBaselineAmount *float64
	OnboardingCompleted   bool
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

const userSelectColumns = `id, google_id, email, display_name, age, avatar_data, default_baseline_amount, (onboarding_completed_at IS NOT NULL)`

func scanUser(row interface{ Scan(...any) error }, u *User) error {
	return row.Scan(&u.ID, &u.GoogleID, &u.Email, &u.DisplayName, &u.Age, &u.AvatarData, &u.DefaultBaselineAmount, &u.OnboardingCompleted)
}

// FindByGoogleID returns the user with this Google account id, or
// (User{}, false, nil) if none exists yet.
func (s *UserStore) FindByGoogleID(ctx context.Context, googleID string) (User, bool, error) {
	var u User
	err := scanUser(s.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE google_id = $1`,
		googleID,
	), &u)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("querying user by google_id: %w", err)
	}
	return u, true, nil
}

// GetByID returns the user with this id, or (User{}, false, nil) if none
// exists.
func (s *UserStore) GetByID(ctx context.Context, userID string) (User, bool, error) {
	var u User
	err := scanUser(s.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE id = $1`,
		userID,
	), &u)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("querying user by id: %w", err)
	}
	return u, true, nil
}

// Create inserts a new user profile for a Google account that has never
// signed in before.
func (s *UserStore) Create(ctx context.Context, googleID, email, displayName string) (User, error) {
	var u User
	err := scanUser(s.db.QueryRowContext(ctx,
		`INSERT INTO users (google_id, email, display_name)
		 VALUES ($1, $2, $3)
		 RETURNING `+userSelectColumns,
		googleID, email, displayName,
	), &u)
	if err != nil {
		return User{}, fmt.Errorf("creating user: %w", err)
	}
	return u, nil
}

// ProfileUpdate holds the fields a PATCH /me/profile call may change -
// nil means "leave unchanged", matching the pattern of a partial update.
type ProfileUpdate struct {
	DisplayName           *string
	Age                   *int
	AvatarData            *string
	DefaultBaselineAmount *float64
}

// UpdateProfile applies a partial update to the user's profile fields
// and marks onboarding complete the first time this is ever called
// (onboarding_completed_at is set only if it was still null - later
// calls, e.g. from Settings, never un-complete it).
func (s *UserStore) UpdateProfile(ctx context.Context, userID string, update ProfileUpdate) (User, error) {
	var u User
	err := scanUser(s.db.QueryRowContext(ctx,
		`UPDATE users SET
			display_name = COALESCE($2, display_name),
			age = COALESCE($3, age),
			avatar_data = COALESCE($4, avatar_data),
			default_baseline_amount = COALESCE($5, default_baseline_amount),
			onboarding_completed_at = COALESCE(onboarding_completed_at, now())
		 WHERE id = $1
		 RETURNING `+userSelectColumns,
		userID, update.DisplayName, update.Age, update.AvatarData, update.DefaultBaselineAmount,
	), &u)
	if err != nil {
		return User{}, fmt.Errorf("updating profile: %w", err)
	}
	return u, nil
}
