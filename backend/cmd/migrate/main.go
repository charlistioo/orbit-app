// Command migrate applies pending .up.sql files from the migrations
// directory, in filename order, tracking what has already run in a
// schema_migrations table so re-running the command is a no-op once
// everything is applied.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("failed to reach database: %v", err)
	}

	if err := applyMigrations(db, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("all migrations applied successfully")
}

func applyMigrations(db *sql.DB, migrationsDir string) error {
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %q: %w", migrationsDir, err)
	}

	var upFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, filename := range upFiles {
		var alreadyApplied bool
		err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`,
			filename,
		).Scan(&alreadyApplied)
		if err != nil {
			return fmt.Errorf("checking migration status for %s: %w", filename, err)
		}
		if alreadyApplied {
			continue
		}

		contents, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("reading %s: %w", filename, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("starting transaction for %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording %s as applied: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing %s: %w", filename, err)
		}

		log.Printf("applied %s", filename)
	}

	return nil
}
