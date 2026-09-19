---
id: TASK-007
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-006]
---

# TASK-007: Consumption Bank engine

**Objective:** Implement the Consumption Bank ledger: auto-credit on
underspend, opt-in debit toward overspend (GET/POST /consumption-bank,
/consumption-bank/apply), floor at 0, cap at 10% of balance per use,
and badge tier calculation.

**Depends on:** TASK-006 (approved)

**Acceptance criteria:**
- Applying the bank never reduces balance below 0.
- Applied amount never exceeds 10% of the balance at the time of the
  request.
- The recorded transaction amount is unchanged by applying the bank;
  the application appears only as a separate ledger entry.
- The correct badge tier is returned for a given balance.

## Design decision confirmed with product owner before building

The objective's "auto-credit on underspend" needed a timing decision
not settled by the existing design: a category is only truly
"underspent" once you sum *all* its transactions for the day and
compare to the plan, not any single transaction in isolation - but
there's no "close the day" step yet (that's TASK-013). Asked the
product owner: real-time (recompute and adjust as each transaction
comes in) vs. wait for Daily Review. **Chose real-time.**

Implementation: each transaction event recomputes "unspent room" for
its category+date = `max(planned - cumulative_actual, 0)` and appends
only the *incremental change* from the previous such computation as one
new ledger entry (positive credit, or negative reversal if a later
transaction eats into an earlier day's leftover) - the ledger stays
append-only (no entry is ever edited), and multiple transactions in the
same category/day never double-credit.

## Submission

**Files changed:**
- backend/internal/store/consumption_bank.go, consumption_bank_integration_test.go
- backend/internal/store/transactions.go (UTC fix, see below)
- backend/internal/api/consumption_bank_handler.go, consumption_bank_handler_test.go
- backend/internal/api/transactions_handler.go (wired bank reconciliation + UTC fix)
- backend/main.go (wired 2 new routes)

**Test results:**
- 8 unit tests with fakes: badge tier boundaries (all 5 tiers + below
  first tier), GET returns balance/badge/ledger correctly, apply
  rejects a missing transaction_id (400) and "nothing to apply" (422),
  apply succeeds and returns the ledger entry.
- Integration tests against real PostgreSQL (the real proof):
  - **Multi-transaction reconciliation**: plan=16000; tx1=12000 credits
    4000; tx2=+3000 (cumulative 15000, still under) reduces credit to
    1000 via a *new* -3000 entry (first entry untouched); tx3=+5000
    (cumulative 20000, now over plan) reverses credit fully to 0 via a
    third entry. 3 ledger rows total, summing to exactly 0 - proves no
    double-crediting across multiple same-day transactions.
  - **Apply cap/floor**: underspend credits 10000; an overspend
    transaction requests apply with no explicit amount -> capped at
    10% of balance (1000, not the full remaining overspend); a second
    apply requesting a huge amount is still capped at 10% of the *new*
    balance (900); the overspend transaction's own recorded amount
    (30000) is verified unchanged in the database throughout.
- Live smoke test against the running server with a real session
  token, reproducing both scenarios above end-to-end (real HTTP calls +
  direct DB checks) - this is where a real timezone bug was caught and
  fixed (see below).
- `go vet` clean.

**Bug found and fixed during live testing:**
`tx.OccurredAt` read back from the database via the driver comes back
in the local time zone (not UTC), so computing a plan date from it near
a local-midnight boundary silently picked the *wrong calendar day* and
skipped reconciliation against the correct plan. Fixed by (1) using the
original, already-UTC `occurredAt` request value in the handler instead
of the DB round-tripped one, and (2) defensively calling `.UTC()` before
formatting a date anywhere a transaction's `occurred_at` is turned into
a plan-lookup key, in both `transactions.go` and `consumption_bank.go`.
Caught by live testing with real HTTP timestamps, not by the unit or
integration tests (which happened to only use pre-UTC Go values
directly) - flagging this as a gap in test coverage going forward.

**Notes/caveats:**
- Concurrent writes to one user's ledger are serialized with a
  PostgreSQL advisory lock (`pg_advisory_xact_lock`) inside each
  reconcile/apply transaction - this is the mitigation for the race
  condition risk documented in the Phase 2 tech-stack review (two
  near-simultaneous requests both reading a stale balance).
- Badge tiers are a pure function of the current balance
  (`badgeTierFor`) - no separate storage needed.
- If Consumption Bank reconciliation fails right after a transaction is
  recorded, the transaction itself is still returned as successfully
  created (only logged server-side) - a bank bookkeeping problem must
  never make the user think their spending wasn't recorded.
- Backend left running for you to try live if you want, same as the
  last few tasks.
