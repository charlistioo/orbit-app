---
id: TASK-011
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-006]
---

# TASK-011: Calendar view

**Objective:** Implement GET /api/v1/calendar returning baseline vs
actual per date and cumulative totals for a given month, automatically
adjusting for the number of days in that month (including leap years).

**Depends on:** TASK-006 (approved)

**Acceptance criteria:**
- A February query in a leap year computes cumulative baseline using
  29 days.

## Submission

**Files changed:**
- backend/internal/store/calendar.go, calendar_integration_test.go
- backend/internal/api/calendar_handler.go, calendar_handler_test.go
- backend/main.go (wired GET /api/v1/calendar)
- mobile/CalendarScreen.tsx (new)
- mobile/App.tsx (renders it post-login)

**Test results:**
- 2 unit tests with a fake store: month param parses correctly and is
  passed through; an invalid month format is rejected 400.
- Integration tests against real PostgreSQL (the core proof): a
  baseline of 20000 set on every day of February 2024 (a leap year)
  produces `days_in_month: 29`, a 29-entry day list, and
  `cumulative_baseline: 580000` (20000 x 29) - directly matching the
  acceptance criteria's exact wording. A contrast test confirms
  February 2023 (not a leap year) reports 28 days.
- Live smoke test against the running server with a real session
  token: `GET /calendar?month=2026-09` correctly returned
  `days_in_month: 30`.
- `go vet` and `tsc --noEmit` clean.

**Notes/caveats:**
- Leap-year correctness comes entirely from Go's standard `time`
  package (`time.Date(year, month+1, 0, ...)` rolls back to the last
  day of the *target* month) - no manual leap-year arithmetic was
  written, which is deliberate: hand-rolled leap-year logic is a classic
  source of off-by-one bugs, and the standard library already gets this
  right.
- `cumulative_baseline` sums each day's *own* `baseline_amount` from
  `daily_plans` for days that actually have a plan set - not a flat
  "baseline rate x days" multiplication. In the test scenario (same
  baseline set every day) both formulas agree, satisfying the
  acceptance criteria exactly; if a user's baseline varies across the
  month or some days have no plan at all, this sums only what was
  actually recorded, which I believe is the more correct and honest
  number - flagging this interpretation in case a flat rate was
  intended instead.
- `cumulative_actual` sums transactions for the whole month regardless
  of whether a plan/baseline exists for that particular day.
- `days` is always the full list for the month (including days with no
  plan and no spending) - `baseline` is `null` for days without one,
  matching product's "not every day needs a plan" flexibility.
- Backend and mobile web app left running for you to try live if you
  want, same as the last few tasks.
