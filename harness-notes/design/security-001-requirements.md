---
id: security-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [api-001]
---

# Security Requirements (id secdesign-1789132172.386202)

1. **Cross-user data access** — every query for user-owned entities must
   filter by user_id from the session token, never from a client-supplied
   parameter.
2. **Forged/expired Google ID token** — backend verifies the ID token with
   Google on every sign-in, issues its own short-lived session token for
   subsequent requests.
3. **Retroactive modification of historical data** — transactions and
   consumption_bank_ledger are append-only at the API level; no
   update/delete endpoints exist for either.
4. **Leaked OAuth secret** — Google OAuth Client Secret lives only in
   backend environment variables, never shipped to the mobile client.

## Integration
Google Sign-In (OAuth 2.0) — the only external dependency for MVP, id
`integdesign-1789132156.728536`. ID token verified server-side; no ORBIT
data is ever sent to Google.

## Non-functional requirements
- Usability: recording a transaction takes at most 4 interaction steps
  (category, item, amount, save), using recency-based quick-select.
- Security: 100% of endpoints touching user-owned data enforce
  session-derived user_id scoping, verified by code review and an
  automated test per endpoint.
