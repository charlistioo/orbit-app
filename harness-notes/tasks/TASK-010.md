---
id: TASK-010
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-006, TASK-007, TASK-008]
---

# TASK-010: Timeline view

**Objective:** Implement GET /api/v1/transactions (chronological,
merged with plan_adjustments, mood entries, and bank ledger entries)
and the mobile Timeline UI.

**Depends on:** TASK-006, TASK-007, TASK-008 (all approved)

**Acceptance criteria:**
- Events from all four sources appear in one chronologically ordered
  list with correct timestamps.

## Submission

**Files changed:**
- backend/internal/store/timeline.go, timeline_integration_test.go
- backend/internal/api/timeline_handler.go, timeline_handler_test.go
- backend/main.go (wired GET /api/v1/transactions)
- mobile/TimelineScreen.tsx (new)
- mobile/App.tsx (renders it post-login)

**Test results:**
- 1 unit test with a fake store confirming the handler passes events
  through as given.
- Integration test against real PostgreSQL (the core proof): created a
  transaction, a bank ledger credit, a mood entry, and a plan
  adjustment - deliberately in a mix of explicit and server-assigned
  timestamps - and confirmed all 4 come back merged into one list,
  correctly sorted, with one event of each of the four types present.
- Live smoke test against the running server with a real session
  token: built up a real plan → transaction → auto-credit → plan edit
  sequence and confirmed `GET /transactions` returned all 4 event types
  in the correct chronological order with real data.
- `go vet` and `tsc --noEmit` clean.

**Bugs found and fixed while building this task's test (both my own
test-code mistakes, not product bugs):**
1. Assumed "now" during a test run would always be later than my
   hardcoded 2026-09-16 test timestamps - but this environment's clock
   is itself set to September 2026, so that assumption was false and
   made the sort-order assertion flaky. Fixed by anchoring the
   explicitly-timestamped events (transaction, mood) to a clearly-past
   date (2020) instead of relying on "now is always later."
2. Mixed `defer` and `t.Cleanup` for cleanup in the same test - they
   run in different phases (`defer` fires at function return, before
   `t.Cleanup` callbacks), so a parent row's `defer`-based delete ran
   *before* a child row's `t.Cleanup`-based delete, violating foreign
   keys and leaking test data on failure. Fixed by using `t.Cleanup`
   consistently everywhere in this test (including for `db.Close()`,
   registered first so it closes last). Verified by running the test
   3 times in a row with `-count=1` with no leftover-data collisions.
   Also caught a genuinely missing cleanup step: `plan_categories` rows
   (created implicitly inside `PlanStore.CreateOrReplace`) had no
   explicit delete at all.

**Notes/caveats:**
- Endpoint path stays `GET /api/v1/transactions` per the approved API
  design, even though it returns the merged Timeline (transactions +
  plan adjustments + mood + bank ledger), not just transactions - the
  design explicitly describes this endpoint as serving the Timeline
  view.
- Merging happens in application code (four separate queries, then
  sorted in Go), not a single SQL UNION - simpler to reason about and
  test given how different each source's columns are; fine at MVP data
  volumes.
- Cleaned up this task's live test data with targeted `DELETE`
  statements only.
- Backend and mobile web app left running for you to try live if you
  want, same as the last few tasks.
