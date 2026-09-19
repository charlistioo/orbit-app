---
id: TASK-005
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-004]
---

# TASK-005: Daily baseline & plan management

**Objective:** Implement daily baseline and per-category plan
creation/editing (POST /plans, PATCH /plans/.../categories/...),
recording plan_adjustments on every change.

**Depends on:** TASK-004 (approved)

**Acceptance criteria:**
- Editing a planned amount creates a plan_adjustment row with old/new
  value and timestamp.
- There is no maximum enforced on baseline or plan values - they are
  soft targets only.

## Submission

**Files changed:**
- backend/migrations/0010_plan_categories_unique.{up,down}.sql (new)
- backend/internal/store/plans.go, plans_integration_test.go
- backend/internal/api/plans_handler.go, plans_handler_test.go
- backend/main.go (wired 3 new routes)

**Test results:**
- 4 unit tests with fakes: a huge baseline/planned amount is accepted
  (never rejected as "too large"), PATCH records exactly one
  old/new-value adjustment, an unknown plan_category_id is 404, an
  unknown plan date is 404.
- Integration test against real PostgreSQL: create plan -> fetch it ->
  PATCH a category to a very large value (no rejection) -> verified the
  `plan_adjustments` row exists with the correct old/new amounts ->
  confirmed a *different* user cannot edit this plan category
  (`ErrPlanCategoryNotFound`, not silently allowed).
- Live smoke test against the running server with a real session token:
  created a plan (baseline 50000, one category planned at 16000),
  fetched it back, PATCHed the category to 25000, and confirmed the
  `plan_adjustments` row (old=16000, new=25000) directly in the
  database.
- `go vet` clean.

**Notes/caveats:**
- `plan_categories` was missing a `UNIQUE (daily_plan_id, category_id)`
  constraint in the TASK-002 schema, which this task's "create or
  replace" semantics for `POST /plans` actually needs (to upsert a
  category's planned amount instead of erroring or duplicating rows).
  Added it via a new migration (`0010`) rather than reworking the
  approved TASK-002 files.
- `POST /plans` (create/replace) does **not** itself write
  `plan_adjustments` - only `PATCH .../categories/...` does. This
  matches the API design's own wording ("PATCH ... records a
  plan_adjustment entry") and the schema (`plan_adjustments` only links
  to a `plan_category_id`, nothing tracks baseline history). Replacing
  a whole day's plan via POST is treated as "setting up the day," not
  "editing an existing value" - flagging this interpretation explicitly
  in case it's not what you had in mind.
- `plan_id` in the PATCH URL isn't separately trusted/validated - the
  actual authority is `plan_category_id` scoped to the session's
  `user_id` via a SQL join (per the standing security note). A
  mismatched `plan_id` in the URL is simply ignored rather than
  producing a confusing separate error.
- Backend left running for you to try live if you want, same as the
  last few tasks - though this task has no mobile UI (the objective
  didn't ask for one, unlike TASK-004).
