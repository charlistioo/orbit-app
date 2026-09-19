---
id: TASK-002
date: 2026-09-15
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, TASK-001]
---

# TASK-002: Database schema & migrations

**Objective:** Create PostgreSQL migrations for all entities: users,
categories, items, daily_plans, plan_categories, plan_adjustments,
transactions, consumption_bank_ledger, mood_entries.

**Depends on:** TASK-001 (approved)

**Acceptance criteria:**
- All 9 tables exist with the fields, constraints, and foreign keys from
  the approved database design.
- Migrations run cleanly on a fresh database.

**Approach:** Plain numbered `.up.sql`/`.down.sql` files in
`backend/migrations/`, applied by a small hand-written Go migration
runner (`backend/cmd/migrate`) using `database/sql` + `pgx` - no external
migration framework, consistent with the MVP "no unneeded technology"
principle. Verified against a real PostgreSQL 16 instance via Docker,
not just reviewed as SQL text.

## Submission

**Files changed:**
- backend/go.mod, backend/go.sum
- backend/migrations/0000_enable_extensions.{up,down}.sql
- backend/migrations/0001_create_users.{up,down}.sql
- backend/migrations/0002_create_categories.{up,down}.sql
- backend/migrations/0003_create_items.{up,down}.sql
- backend/migrations/0004_create_daily_plans.{up,down}.sql
- backend/migrations/0005_create_plan_categories.{up,down}.sql
- backend/migrations/0006_create_plan_adjustments.{up,down}.sql
- backend/migrations/0007_create_transactions.{up,down}.sql
- backend/migrations/0008_create_consumption_bank_ledger.{up,down}.sql
- backend/migrations/0009_create_mood_entries.{up,down}.sql
- backend/cmd/migrate/main.go

**Test results:**
- Spun up a real PostgreSQL 16 instance via Docker (not just reviewed as
  SQL text).
- Ran the migrate command against it: all 10 `.up.sql` files (extensions
  + 9 tables) applied cleanly with no errors.
- Verified via `information_schema.columns` that all 9 tables exist with
  exactly the fields from the approved database design (49 columns
  total, correct data types).
- Re-ran the migrate command against the same database: no-op, no
  errors (idempotency check).
- Ran the migrate command again against a brand-new, never-touched
  database (`orbit_fresh`): applied cleanly - directly satisfies
  "Migrations run cleanly on a fresh database."
- `go vet ./...` clean.

**Notes/caveats:**
- Used a small hand-written migration runner (`cmd/migrate`, ~100 lines,
  `database/sql` + `pgx`) instead of an external migration framework -
  no migration/ORM tooling had been decided in the technical design, and
  this keeps the dependency footprint minimal per the MVP simplicity
  principle. It tracks applied migrations in a `schema_migrations` table
  and applies each pending `.up.sql` file inside its own transaction.
- `.down.sql` files are included for every table (not required by the
  acceptance criteria, but standard practice for reversibility) - the
  migrate command does not currently apply them; that's for a future
  task if rollback tooling is ever needed.
- Test PostgreSQL container was removed after verification; this task
  does not include a docker-compose file for local dev Postgres, since
  that wasn't in the objective or expected_files - happy to add it as a
  follow-up if useful.
