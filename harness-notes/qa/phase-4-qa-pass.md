---
id: phase-4-qa-pass
date: 2026-09-16
product_id: my-app
phase: 4
source: claude-code
related: [TASK-001, TASK-006, TASK-007, TASK-009, TASK-012, TASK-013, TASK-014]
---

# Phase 4 QA pass

Harness does not generate or execute test cases itself - these were
written by hand, grounded in the 2 approved functional requirements'
acceptance criteria (`fr-1789128519.964188`, `fr-1789132539.593635`)
and the security threat list from the technical design's testing scope
manifest, then submitted and recorded via the QA API.

## API test cases (7) - fr-1789128519.964188 (Consumption Bank / Flexible Baseline)

- **TEST-API-001** - Overspend transaction recorded in full, never
  blocked or reduced. Evidence: `TestTransactionStore_PlanSnapshotIsImmutable`
  (real PostgreSQL), `TestCreateTransaction_ExceedsPlan_StillRecordedInFull`.
- **TEST-API-002** - Plan snapshot immutable after a later plan edit.
  Evidence: `TestTransactionStore_PlanSnapshotIsImmutable`.
- **TEST-API-003** - Consumption Bank apply capped at 10% of balance,
  floor 0. Evidence: `TestConsumptionBank_ApplyRespectsCapAndFloor`.
- **TEST-API-004** - Badge tier correct at every threshold boundary.
  Evidence: `TestBadgeTierFor` plus the TASK-014 live smoke test
  (balance pushed through all 6 states: none/Frugal/Thrifty/
  Economical/Collector/Master).

## API test cases (3) - fr-1789132539.593635 (server-side computation)

- **TEST-API-005** - Home endpoint fully computed, no client math
  needed. Evidence: `TestGetHome_FullyComputedSummary` plus a fresh
  live `GET /home` call made during this QA pass.
- **TEST-API-006** - No multi-day trend claim with only 1 day of
  history. Evidence: `TestGenerate_NoTrendClaimWithOneDayOfHistory`,
  `TestInsightStore_NoTrendClaimWithOneDayOfHistory`, plus a fresh live
  `GET /insights/daily` call.
- **TEST-API-007** - Reminder pending-categories computed server-side.
  Evidence: `TestDailyReviewStore_PendingCategoriesExcludesRecordedOnes`.

## Security test cases (4)

- **SEC-API-001** - Cross-user isolation. **Freshly executed live**
  during this QA pass (not just cited from Phase 3): created two real
  throwaway accounts (A, B) against the shared test database; B was
  given a category, item, and transaction. As A: `POST
  /categories/{B's category_id}/items` returned 404 before any row was
  created. `GET /categories`, `/transactions`, and `/home` as A all
  returned empty/zeroed results despite B's real data existing in the
  same database at the same time. Test data cleaned up afterward with
  targeted deletes.
- **SEC-API-002** - Forged/expired Google ID token rejected. Evidence:
  `TestGoogleSignIn_InvalidTokenRejectedWith401` plus the TASK-003 live
  smoke test. Noted limitation: an actually-expired *real* Google
  token wasn't obtainable for this pass (needs a live interactive
  sign-in that has since expired) - expiry validation itself is
  exercised via golang-jwt's standard `exp` claim check against
  Google's real JWKS (`TestRealGoogleVerifier_CanFetchRealJWKS`).
- **SEC-API-003** - Protected endpoints reject missing/garbage
  session tokens. **Freshly executed live**: `GET /me`, `/home`,
  `/categories` each tested with no `Authorization` header and with
  `Authorization: Bearer garbage-not-a-jwt` - all 6 combinations
  returned 401.
- **SEC-API-004** - No PATCH/PUT/DELETE exists for transactions or the
  Consumption Bank ledger. **Freshly executed live**: `PATCH`/`DELETE`
  on `/transactions/{id}` returned 404 (no such route pattern
  registered at all); `PATCH`/`DELETE` on `/consumption-bank` returned
  405 (route exists for GET only). Confirmed against `backend/main.go`'s
  route table directly as well.

## Result

- Coverage: both functional requirements `verified`.
- Release readiness: `requirement_status: 2/2 verified`,
  `test_status: 11/11 passed`, `security_status: 4 security test
  case(s) tracked`, `bug_status: 0 open`.
- `release_status: release_ready` - human acceptance recorded by
  Charli Stiow on 2026-09-16, in direct response to this QA write-up.

## Known limitations carried over from implementation

See the implementation completion report
(`GET /qa/my-app/readiness`) for the full per-task list - most notably:
a real end-to-end Google Sign-In with a live account was never
verifiable by the agent itself (needs an interactive login - the
product owner tested this manually during TASK-003), and this project
has no JS/TS test runner configured for the mobile app yet (mobile
verification has consistently relied on `tsc --noEmit` plus manual/live
smoke testing rather than automated component tests).
