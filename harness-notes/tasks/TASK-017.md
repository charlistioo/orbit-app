---
id: TASK-017
date: 2026-09-18
product_id: my-app
phase: 3
source: claude-code
status: approved
related: [techstack-002, TASK-016]
---

# TASK-017: Web core loop - Beranda dashboard, Catat + hasil, Mood, Adjust Plan

**Objective:** Build the full Beranda dashboard (screen 3), the daily
plan screen (Rencanakan harimu, screen 2b, pre-filled from
`default_baseline_amount`), Record Consumption + result screen
(screens 4-5), Update Mood (screen 6), and Adjust Plan (screen 7) in the
`web/` Next.js app - all against the existing, unmodified Go backend.

**Depends on:** TASK-016 (approved).

## Submission

**Files changed (all in `web/`, no backend changes needed):**
- lib/api.ts (added `Category`, `Item`, `DailyPlan`, `PlanCategory`,
  `HomeData`, `Transaction`, `TimelineEvent`, `MoodEntry` types,
  `todayStr`/`formatRupiah` helpers)
- app/page.tsx (rewritten) - Beranda now branches on whether a plan
  exists for today: no plan -> `PlanToday` (screen 2b: total budget
  pre-filled from the profile's `default_baseline_amount`, per-category
  allocation with inline "add category"); plan exists -> `MainDashboard`
  (screen 3: circular gauge via HeroUI `ProgressCircle`, per-category
  bars, 3 action buttons, mood card, Consumption Bank card)
- app/catat/page.tsx (new) - screen 4, quick-select category/item chips
  from real data, no fuzzy matching (TASK-004's rule)
- app/catat/hasil/page.tsx (new) - screen 5, neutral plan-vs-actual
  result (never framed as pass/fail; overspend shown as plain
  information, matching the non-punitive design principle)
- app/mood/page.tsx (new) - screen 6
- app/rencana/sesuaikan/page.tsx (new) - screen 7, two-step
  pick-category-then-edit-amount flow

**Test results:**
- `tsc --noEmit` clean; `next build` compiles and statically generates
  all 12 routes (the original 8 plus catat, catat/hasil, mood,
  rencana/sesuaikan) with no errors.
- Live smoke test directly against the real backend (same requests the
  new pages make): completed onboarding for the live test account,
  created today's plan (`POST /plans`, baseline 70000, Kopi 30000 +
  Makanan 40000), confirmed `GET /home` reflected it (`plan: 70000`),
  recorded a transaction (`POST /transactions`, Kopi 12000) and
  confirmed `PlanAmountSnapshot: 30000` came back exactly as the result
  screen needs it, confirmed `GET /home` updated correctly
  (`actual: 12000`, `remaining: 58000`, Consumption Bank credited
  `18000` from the underspend), adjusted the Kopi category via `PATCH
  /plans/{id}/categories/{id}` (30000 -> 40000) and confirmed `GET
  /home` reflected the new plan total (80000), and logged a mood via
  `POST /mood`. Every field name/shape the new pages depend on
  (`Categories[].CategoryID`, `PlanAmountSnapshot`, etc.) matched what
  the live backend actually returns.
- Left the live test account in a populated demo-ready state (onboarded,
  a real plan, one transaction, a mood entry) so opening the browser
  now shows the Main Dashboard directly rather than the empty-plan
  screen.

**Notes/caveats:**
- **Found during this task, not caused by it:** the live test account's
  onboarding data from the TASK-016 walkthrough (name "Tio", budget
  Rp70.000) was gone when I checked at the start of this task - the dev
  server/backend processes from that earlier session apparently didn't
  survive between conversation turns, even though the PostgreSQL
  container itself stayed up throughout (3 days uptime, data for other
  tasks intact). Nothing in this task's own code caused it; noting it in
  case you want a more durable way to keep the local dev backend running
  across sessions.
- The result screen's "encouraging" vs "neutral" messages are driven
  purely by the sign of `planned - actual` from real numbers the backend
  already returned on the just-created transaction - no new judgment
  logic invented.
- "Sesuaikan Rencana" edits one category at a time (matches screen 7
  exactly); there's no bulk-edit-all-categories screen since the
  reference doesn't show one.
- Both dev servers left running for you to try:
  `http://localhost:3000` (already onboarded, dashboard populated) and
  backend on `:8080`.

## Post-submission fix (before approval)

Product owner reported via live use: recording the same category twice
in one day (Transportasi Rp3.000, then Rp3.000 again against a
Rp15.000 plan) showed the result screen crediting Consumption Bank as
if each transaction were the only one that day, instead of accounting
for the cumulative total. Root cause: `app/catat/hasil/page.tsx`
computed the shown credit as `planned - thisTransactionAmount` client-
side, ignoring that the backend's own reconciliation
(`ReconcileForTransaction`) is already correctly incremental
(append-only, never double-credits) - the bug was purely in what the
result screen *displayed*, not in the backend's actual ledger.

Fixed in `app/catat/page.tsx` and `app/catat/hasil/page.tsx`:
- "Aktual" now shows the category's cumulative spend for the day
  (summed from Timeline), not just the single transaction just saved.
- The credit/debit badge now reads the *real* Consumption Bank balance
  before and after the save (two `GET /consumption-bank` calls around
  the `POST /transactions`) instead of recomputing it client-side.
- Found and fixed a second issue during live verification of the first
  fix: the result message conflated "bank credit decreased" with
  "you overspent," which is wrong - a credit can shrink via the
  backend's own auto-correction even while the category is still well
  under its plan. The message is now driven by the real
  planned-vs-cumulative-actual comparison, independent of the badge's
  sign.

Live-verified by reproducing the exact reported scenario directly
against the backend (Transportasi plan 15000, two 3000 transactions):
first transaction credited +12000, second correctly corrected it to
+9000 net (delta -3000 on the second), matching the product owner's
expected number exactly.

Also did a UI tidy-up pass per request: category allocation rows on
"Rencanakan Harimu" now sit in one bordered card with row dividers
instead of floating inputs, and the three dashboard action buttons
(Catat/Mood/Sesuaikan) are now bordered card-style buttons instead of
thin pills.

`tsc --noEmit` and `next build` both clean after the fix (all 12
routes). Approved by Charli Stiow after this fix, in the same
conversation turn as "Lanjutkan ke checkpoint selanjutnya."
