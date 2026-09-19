---
id: TASK-012
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-006, TASK-007, TASK-008]
---

# TASK-012: Rule-based insight engine

**Objective:** Implement GET /api/v1/insights/{daily,weekly,monthly}
generating template-based insight text from plan-vs-actual, Consumption
Bank usage, category frequency, and mood timing - never inventing
conclusions when data is insufficient.

**Depends on:** TASK-006, TASK-007, TASK-008 (all approved)

**Acceptance criteria:**
- A user with only 1 day of data receives no multi-day trend claim.
- A user with 3+ consecutive days of rising spend in one category
  receives a trend observation.
- No insight text ever states mood as the *cause* of spending - only
  temporal correlation.

## Submission

**Update (same task, before approval):** after the first submission, the
user asked to also add the Consumption Bank and category-frequency
insight lines that the initial submission had deliberately left out
(flagged as an open question) since the acceptance criteria didn't
explicitly require them. Both are now included, following the same
never-invent-a-conclusion discipline as the rest of the engine - see
below.

**Files changed:**
- backend/internal/insight/rules.go (new) - pure, DB-free rule engine:
  `Generate(Inputs) []string`. Now also emits a Consumption Bank
  credited/applied line whenever the period had non-zero ledger
  movement, and a "most frequently used category" line - but only when
  there are 2+ categories to compare **and** no tie for first place, so
  a lone category (or an exact tie) is never reported as "most
  frequent" without support.
- backend/internal/insight/rules_test.go (new) - 12 unit tests: the
  original 7 plus 5 new ones for Consumption Bank credited/applied,
  "no bank insight when nothing moved", "most frequent category
  reported", "no claim on a tie", "no claim with only one category"
- backend/internal/store/insight.go (rewritten) - fetches aggregates
  from Postgres (plan-vs-actual for a date range, distinct
  transaction-day count, 3-day category trend window, mood-before-
  spending, Consumption Bank movement for the range, category
  transaction-frequency for the range) and calls `insight.Generate`
- backend/internal/store/insight_integration_test.go (new) - 5
  end-to-end tests against real PostgreSQL: the original 3 (one per
  acceptance criterion) plus 2 new ones proving the Consumption Bank
  and category-frequency wiring
- backend/internal/api/insight_handler.go (new) - GetDaily/GetWeekly/
  GetMonthly HTTP handlers
- backend/internal/api/insight_handler_test.go (new) - handler unit
  tests with a fake store (auth passthrough, param passthrough, default
  date/month)
- backend/main.go (wired GET /api/v1/insights/daily, /weekly, /monthly)
- mobile/InsightsScreen.tsx (new)
- mobile/App.tsx (renders it post-login)
- backend/internal/store/consumption_bank_integration_test.go,
  plans_integration_test.go, transactions_integration_test.go (cleanup
  ordering bug fix, on request - see notes)

**Test results:**
- `internal/insight`: 12/12 unit tests pass, no database involved -
  the most direct proof of the 3 required acceptance criteria plus the
  2 added rules:
  `TestGenerate_NoTrendClaimWithOneDayOfHistory`,
  `TestGenerate_TrendObservationFor3ConsecutiveRisingDays`,
  `TestGenerate_MoodNeverStatedAsCause` (also checks the message
  explicitly frames itself as correlation, not causation, and rejects
  a list of causal phrasings), `TestGenerate_BankCreditedAndApplied`,
  `TestGenerate_NoBankInsightWhenNothingMoved`,
  `TestGenerate_MostFrequentCategoryReported`,
  `TestGenerate_NoMostFrequentCategoryOnTie`,
  `TestGenerate_NoMostFrequentCategoryWithOnlyOneCategory`, plus 4 more
  covering "not strictly rising", "no correlation, no mood insight",
  "no data at all", and "overspend framing".
- `internal/api`: 3/3 new handler unit tests pass (auth required,
  authenticated user/date passed through to the store, default month
  applied when omitted) using a fake `InsightStore`; full package
  (33 tests) still green.
- `internal/store`: 5/5 new integration tests against real PostgreSQL
  pass, each directly proving one behavior end-to-end (real user, real
  category, real transactions/mood/plan/ledger rows, real SQL
  aggregation, real `insight.Generate` call): 1-day history produces
  no trend claim; 3 consecutive days of strictly rising spend in
  "Kopi" (Rp10.000 -> Rp20.000 -> Rp30.000) produces a trend
  observation naming "Kopi"; a mood entry logged before a same-day
  transaction produces a mood insight that never uses a causal phrase;
  an auto-reconcile credit from underspending Rp8.000 against a
  Rp20.000 plan is surfaced as a Consumption Bank insight; 3 Kopi
  transactions vs 1 Makanan transaction in the same week reports Kopi
  as the most frequently used category. Full package (17 tests) green.
- Live smoke test: rebuilt binary, real backend process against the
  shared test Postgres container, real session token for the live test
  account (`f9e75a8b-...`, Kopi category). `/insights/daily` and
  `/weekly` both responded correctly with the "belum ada data" message
  (that account has no transactions in the current date/week) -
  confirms the expanded engine didn't break the no-data path or crash
  the server.
- `go build ./...`, `go vet ./...`, `npx tsc --noEmit` all clean.

**Notes/caveats:**
- The rule logic lives in its own DB-free package
  (`internal/insight`), separate from `internal/store`, specifically so
  the 3 acceptance criteria could be unit tested as pure functions
  (construct `Inputs`, call `Generate`, assert on strings) without
  needing a database or even a fake store - the integration tests then
  separately prove the SQL aggregation feeding that engine is correct.
  This is a slightly different shape than earlier tasks (which put all
  logic straight in `store`), but I believe it's the more testable
  design for "never invent a conclusion" rules specifically, since the
  rules themselves - not just the SQL - are what the acceptance
  criteria are actually about.
- "3+ consecutive days of rising spend" is checked per-category over
  the 3 most recent distinct calendar days up to the report's
  reference date (today for daily, the week's end date for weekly, the
  month's last day for monthly) - not the whole reporting window. A
  weekly or monthly report still won't claim a trend unless the
  account has at least 3 total distinct days of transaction history
  anywhere, satisfying the "not enough data" rule even for longer
  periods.
- Mood is never treated as a cause anywhere in the generated text - the
  only mood-related insight is phrased as "a mood entry was logged
  before some transactions" (temporal ordering only), and only appears
  at all when that ordering is actually true for the period.
- Consumption Bank credited/applied is summed by joining each ledger
  row to its related transaction's `occurred_at` (not the ledger row's
  own `created_at`) - a reconciliation can legitimately be written
  later than the transaction it's for, and it's the transaction's date
  the report is organized around. Caught this via a failing integration
  test before it shipped.
- "Most frequently used category" requires at least 2 categories with
  transactions in the period, and refuses to name one on an exact tie -
  both cases where naming a "most frequent" category wouldn't actually
  be supported by the data.
- **Found and fixed, on explicit request (out of TASK-012's own
  scope):** while repeatedly running the full suite I found that 3
  pre-existing integration test files -
  `consumption_bank_integration_test.go`, `plans_integration_test.go`,
  `transactions_integration_test.go` - never deleted the
  `plan_categories` row that `PlanStore.CreateOrReplace` creates
  implicitly (and, in the bank tests, deleted transactions *before*
  the `consumption_bank_ledger` rows that reference them). Their
  cleanup used `defer db.ExecContext(...)` with the error discarded,
  so once a child row still referenced a parent being deleted, the
  delete failed silently and the test still reported PASS while
  leaving a `users` row behind - which then broke the *next* run of
  those same tests with a duplicate-key error. Same bug class already
  fixed in `timeline_integration_test.go` during TASK-010, just never
  applied to these three older files.

  After I reported this, you asked me to fix it, so I did: switched
  every `defer` in these 3 files to `t.Cleanup`, added the missing
  `plan_categories` (and, in `plans_integration_test.go`,
  `plan_adjustments`) deletes, and reordered registration so cleanup
  always runs child-before-parent, `consumption_bank_ledger` before
  the transactions it references, `db.Close()` last. Verified with 5
  consecutive full runs of `internal/store` (all green) and confirmed
  0 leftover `integration-%` rows in the database afterwards - the
  tests now genuinely clean up after themselves instead of merely
  reporting PASS while leaking rows.
- The `cmd/minttoken` tool used for the live smoke test was deleted
  again before this submission, per the usual convention.
