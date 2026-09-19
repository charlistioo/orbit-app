---
id: TASK-014
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-007]
---

# TASK-014: Badge tier display

**Objective:** Implement the mobile UI component showing the current
Consumption Bank badge tier (Frugal/Thrifty/Economical/Collector/
Master) based on balance.

**Depends on:** TASK-007 (approved)

**Acceptance criteria:**
- The correct badge is shown for each of the five threshold tiers.

## Submission

**Files changed:**
- mobile/BadgeDisplay.tsx (new)
- mobile/App.tsx (renders it post-login)

**Test results:**
- The 5-tier threshold math itself (`badgeTierFor`, >=100k/200k/300k/
  400k/500k) was already implemented and unit-tested in TASK-007
  (`TestBadgeTierFor`, still green - full `internal/api` package
  47/47 green, full backend suite green across all 4 packages). This
  task is mobile-only: display whichever tier the server reports.
- `mobile/BadgeDisplay.tsx` exports a pure `badgeVisual(badge)`
  function with an exhaustive `switch` over the 5 tiers plus the
  "no badge yet" empty-string case, no `default` branch. Verified this
  is actually exhaustive (not just visually) by temporarily deleting
  the `Collector` case and re-running `tsc --noEmit`: it failed with
  "Function lacks ending return statement" as expected, then restored
  the case and confirmed `tsc --noEmit` clean again - proves the
  compiler itself will catch a missing tier if one is ever dropped,
  which is the most direct guarantee available for "the correct badge
  is shown for each of the five tiers" on a project with no JS test
  runner configured yet.
- Live smoke test: rebuilt backend, started it against the real test
  Postgres container, created one throwaway user, and inserted
  `consumption_bank_ledger` rows via SQL to push its balance through
  50000 -> 150000 -> 250000 -> 350000 -> 450000 -> 550000. Queried
  `GET /consumption-bank` with a minted session token after each step
  and got badge `"" -> "Frugal" -> "Thrifty" -> "Economical" ->
  "Collector" -> "Master"` - all 5 tiers plus the empty pre-tier state,
  confirmed end-to-end through the real API this component consumes.
  Cleaned up the throwaway user and ledger rows afterward.
- `npx tsc --noEmit` clean on the final component.

**Notes/caveats:**
- No backend changes were needed or made - `GET /api/v1/consumption-
  bank` already returned a `badge` field since TASK-007. This task's
  entire scope was the display component.
- Kept `BadgeDisplay` as its own dedicated component (fetches
  `/consumption-bank` itself) rather than folding it into the existing
  inline badge text already shown on `HomeScreen` (TASK-009), since
  the Work Order specifically names "the mobile UI component showing
  the current Consumption Bank badge tier" as this task's own
  deliverable - `HomeScreen`'s existing inline badge text was left
  untouched, out of scope for this task.
- This project has no JS/TS unit test runner configured (`package.json`
  has no `jest`/`vitest`/etc.) - all prior mobile components have been
  verified via `tsc --noEmit` plus a live smoke test in the running
  app, and this task follows the same pattern. Happy to set one up as
  a follow-up if you'd like actual `badgeVisual` unit tests going
  forward instead of the compiler-exhaustiveness proof used here.
