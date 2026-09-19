---
id: TASK-018
date: 2026-09-19
product_id: my-app
phase: 3
source: claude-code
status: approved
related: [techstack-002, TASK-017]
---

# TASK-018: Web Bank Konsumsi, Riwayat (Timeline), Wawasan (daily/weekly insight)

**Objective:** Build a dedicated Consumption Bank page (screen 8), the
Riwayat tab as a daily timeline (screen 9), and the Wawasan tab with
daily insight (screen 10) and weekly insight with a bar comparison
(screen 11) - all against the existing, unmodified Go backend.

**Depends on:** TASK-017 (approved).

## Submission

**Files changed (all in `web/`, no backend changes needed):**
- lib/api.ts (added `LedgerEntry`, `BankData`, `InsightResult` types;
  extended `TimelineEvent` with `plan_adjustment`/`mood`/`bank_ledger`
  payload shapes; added `BADGE_TIERS`/`nextBadgeTier`/`currentTierFloor`
  helpers)
- app/bank/page.tsx (new) - screen 8, balance/badge/tier-progress from
  `GET /consumption-bank`, per-category Direncanakan/Terpakai/Tersimpan
  from `GET /plans/{date}` combined with Timeline-derived actuals
- app/riwayat/page.tsx (rewritten from placeholder) - screen 9, the
  merged Timeline grouped by date, using the real category/item names
  the backend now returns (TASK-016's enrichment) instead of raw ids
- app/wawasan/page.tsx (rewritten from placeholder) - screens 10 & 11,
  a Harian/Mingguan segmented toggle; daily shows `GET /insights/daily`
  as-is; weekly shows a hand-rolled bar comparison (planned vs actual
  for the last 7 days, computed from real `GET /plans/{date}` x7 +
  Timeline) plus `GET /insights/weekly`'s own text
- app/page.tsx (small change) - the Consumption Bank dashboard card now
  links to `/bank`

**Test results:**
- `tsc --noEmit` clean; `next build` compiles and statically generates
  all 13 routes (the original 12 plus `/bank`) with no errors.
- Live smoke test against the real running backend: `GET
  /consumption-bank` (balance 27000, full ledger with real
  auto-reconcile entries), `GET /insights/daily` and `/insights/weekly`
  (real rule-based text, including a mood-correlation note phrased as
  temporal-only and a most-frequent-category note), and `GET
  /plans/{date}` (3 real categories) all returned exactly the shapes
  the new pages expect. All 3 new routes (`/bank`, `/riwayat`,
  `/wawasan`) responded 200.

**Notes/caveats:**
- The weekly bar chart is intentionally hand-rolled (`div`/CSS height,
  no charting library) - the same "no unneeded technology" reasoning
  applied throughout this project.
- Weekly "Direncanakan" total is the sum of each of the last 7 days'
  own plan (0 for a day with none) - not a flat rate x7 - matching the
  same "sum what was actually recorded" interpretation as TASK-011's
  Calendar `cumulative_baseline`.
- Consumption Bank page's per-category "Tersimpan" is scoped to
  *today's* plan only (there's no per-category historical plan list
  endpoint) - consistent with the dashboard's own scope.
- Both dev servers left running: `http://localhost:3000` and backend
  on `:8080`.
