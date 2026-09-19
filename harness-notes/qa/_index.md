# QA Index (Phase 4)

Started 2026-09-16. Test cases written against the 2 approved functional
requirements' acceptance criteria plus the security threat list from
the technical design's testing scope manifest.

| Test | Title | Requirement | Result |
|---|---|---|---|
| TEST-API-001 | Overspend recorded in full | fr-1789128519.964188 | pass |
| TEST-API-002 | Plan snapshot immutable | fr-1789128519.964188 | pass |
| TEST-API-003 | Consumption Bank apply cap/floor | fr-1789128519.964188 | pass |
| TEST-API-004 | Badge tier per threshold | fr-1789128519.964188 | pass |
| TEST-API-005 | Home fully computed server-side | fr-1789132539.593635 | pass |
| TEST-API-006 | No trend claim with 1 day history | fr-1789132539.593635 | pass |
| TEST-API-007 | Reminder pending categories | fr-1789132539.593635 | pass |
| SEC-API-001 | Cross-user data isolation | (security) | pass |
| SEC-API-002 | Forged/expired Google token rejected | (security) | pass |
| SEC-API-003 | Protected endpoints require valid session | (security) | pass |
| SEC-API-004 | Transactions/ledger have no mutation route | (security) | pass |

**Coverage:** 2/2 functional requirements verified.
**Release readiness:** requirement_status 2/2 verified, test_status
11/11 passed, security 4 test cases tracked, 0 open bugs.
`release_status: release_ready` - human acceptance recorded by Charli
Stiow on 2026-09-16.

See [phase-4-qa-pass.md](phase-4-qa-pass.md) for the full write-up.
