---
id: TASK-006
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-005]
---

# TASK-006: Transaction recording with plan snapshot

**Objective:** Implement POST /api/v1/transactions, snapshotting the
plan amount in effect at the time of recording (plan_amount_snapshot).

**Depends on:** TASK-005 (approved)

**Acceptance criteria:**
- A transaction exceeding the plan is still recorded in full, never
  blocked or reduced.
- Editing the plan after a transaction was recorded does not change
  that transaction's stored plan_amount_snapshot.

## Submission

**Files changed:**
- backend/internal/store/transactions.go, transactions_integration_test.go
- backend/internal/store/plans.go (added `FindPlanCategoryForDate`)
- backend/internal/store/items.go (added `Exists` ownership check)
- backend/internal/api/transactions_handler.go, transactions_handler_test.go
- backend/main.go (wired POST /api/v1/transactions)

**Test results:**
- 4 unit tests with fakes: an amount exceeding the plan is recorded in
  full (never reduced/blocked), an unowned category is rejected 404
  before any transaction is created, a missing amount is 400, an
  unowned item_id is rejected 404.
- Integration test against real PostgreSQL - the core proof for this
  task: created a plan (category planned at 16000), recorded a
  transaction for 26000 (over plan, not reduced), confirmed
  `plan_amount_snapshot = 16000` on that transaction, then edited the
  plan category to 99999, and re-read the transaction directly from the
  database - `plan_amount_snapshot` was still 16000, completely
  unaffected by the later edit. A second integration test confirms
  recording a transaction with no plan at all for that date/category
  works fine (`plan_amount_snapshot` is simply null).
- Live smoke test against the running server with a real session token:
  same sequence as above (create plan -> record over-plan transaction
  -> edit plan -> re-check transaction) reproduced with real HTTP calls
  and a direct database check - identical result.
- `go vet` clean.

**Notes/caveats:**
- Added ownership checks for both `category_id` and (when provided)
  `item_id` before recording anything - consistent with the standing
  security note and the same pattern used in TASK-004/005 for
  category/plan-category ownership.
- The plan snapshot is looked up by matching the transaction's
  `occurred_at` date against the user's daily plan for that date (if
  any) and category - if there's no plan for that date/category,
  `plan_category_id` and `plan_amount_snapshot` are simply left null;
  the transaction is still recorded normally either way.
- This task only implements `POST /api/v1/transactions`, per the Work
  Order's objective. `GET /api/v1/transactions` (listing, for the
  Timeline view) is TASK-010's scope, not built here.
- Backend left running for you to try live if you want, same as the
  last few tasks.
