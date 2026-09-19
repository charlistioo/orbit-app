---
id: TASK-009
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-006, TASK-007, TASK-008]
---

# TASK-009: Home screen aggregation

**Objective:** Implement GET /api/v1/home returning today's baseline,
plan, actual, remaining, Consumption Bank balance/badge, and current
mood in one response, plus the mobile Home screen.

**Depends on:** TASK-006, TASK-007, TASK-008 (all approved)

**Acceptance criteria:**
- The endpoint returns a fully computed summary requiring no further
  client-side aggregation.

## Submission

**Files changed:**
- backend/internal/store/transactions.go (added `SumForDate`)
- backend/internal/store/mood.go (added `Latest`)
- backend/internal/api/home_handler.go, home_handler_test.go
- backend/main.go (wired GET /api/v1/home)
- mobile/HomeScreen.tsx (new)
- mobile/App.tsx (renders it post-login)

**Test results:**
- 2 unit tests with fakes: a full scenario (plan + transactions + mood
  + bank balance all present) produces the exact right numbers
  (baseline 50000, plan total 61000 summed across 3 categories, actual
  42000, remaining 19000 = plan - actual, bank balance/badge passed
  through, mood passed through); a brand-new user with nothing recorded
  yet gets all zeroes/nulls rather than an error.
- Live smoke test against the running server with a real session
  token: created a plan (baseline 50000, one category planned at
  16000), recorded a transaction of 12000 under it, logged a mood, then
  called `GET /home` - got back
  `{"baseline":50000,"plan":16000,"actual":12000,"remaining":4000,"consumption_bank":{"balance":4000,"badge":""},"mood":"good"}`
  in one response, matching every underlying record exactly with no
  further computation needed.
- `go vet` and `tsc --noEmit` clean.

**Notes/caveats:**
- `remaining` is defined as `plan - actual` (not `baseline - actual`) -
  this matches the product's own worked example exactly (baseline
  50000, plan 60000, actual 42000, remaining 18000 = 60000-42000).
- `date` is an optional query parameter (`?date=YYYY-MM-DD`), defaulting
  to today in UTC if omitted - the original API design didn't specify
  how the client communicates "today" in its own time zone, so I added
  this as an escape hatch the mobile client can use later without
  requiring it now (matches the no-param `GET /api/v1/home` shape from
  the approved API design).
- "Current mood" is the user's most recently logged mood entry overall
  (not filtered to today) - mood is optional and sparse, so "the last
  one you told us" is a more useful signal than "none today."
- No consumption-bank-badge threshold was hit in the live test (balance
  4000, well under the 100000 Frugal tier) - badge logic itself was
  already unit- and integration-tested in TASK-007.
- I cleaned up this task's live test data with targeted `DELETE`
  statements only (not `TRUNCATE`), after the TASK-007/008 incident -
  your "Kopi" category and account are untouched.
- Backend and mobile web app left running for you to try live if you
  want, same as the last few tasks.
