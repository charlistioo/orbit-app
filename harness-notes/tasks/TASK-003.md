---
id: TASK-003
date: 2026-09-15
product_id: my-app
phase: 3
source: harness
status: approved
related: [security-001, TASK-002]
---

# TASK-003: Google Sign-In authentication

**Objective:** Implement POST /api/v1/auth/google (verify Google ID
token, create/fetch user, issue ORBIT session token) and the mobile
Google Sign-In flow.

**Depends on:** TASK-002 (approved)

**Acceptance criteria:**
- A new Google account signing in for the first time creates exactly
  one user profile.
- An invalid or expired ID token is rejected with 401.
- All subsequent endpoints require a valid ORBIT session token.

**Google OAuth Client ID used:**
`595683522917-al9clne9te05q8gb55a1t8v2t0kvf39u.apps.googleusercontent.com`
(Web application type, registered by the product owner in Google Cloud
Console; consent screen set to External/testing mode.)

## Submission

**Files changed:**
- backend/internal/auth/google.go, google_jwks.go, google_jwks_live_test.go
- backend/internal/auth/session.go, session_test.go
- backend/internal/auth/middleware.go, middleware_test.go
- backend/internal/store/users.go, db.go, users_integration_test.go
- backend/internal/api/auth_handler.go, auth_handler_test.go
- backend/main.go (wired routes, requires DATABASE_URL, GOOGLE_CLIENT_ID, SESSION_SECRET env vars)
- backend/go.mod, backend/go.sum
- mobile/LoginScreen.tsx (new)
- mobile/App.tsx (shows LoginScreen when signed out, calls protected /api/v1/me when signed in)
- mobile/package.json (expo-auth-session, expo-web-browser, expo-crypto)

**Test results:**
- 12 unit tests passing (`go test ./internal/...`): session token
  issue/verify round-trip, wrong-secret rejection, expired-token
  rejection, garbage-token rejection; middleware rejects missing/invalid
  Authorization header and passes through on a valid one; the sign-in
  handler (tested with a fake Google verifier and an in-memory fake user
  store, no real network/DB): a brand-new Google account creates exactly
  one user (Create called once), an existing google_id does not create a
  duplicate (Create called zero times), an invalid/expired token is
  rejected with 401 and never reaches user creation, a missing id_token
  is rejected with 400, an unexpected verifier error is a 500.
- Verified the JWKS-fetching code against Google's real public keys
  endpoint (network call, no fake data) - successfully fetched and
  parsed 2 real signing keys, confirming the fetch/parse logic matches
  Google's actual response shape.
- `go vet ./...` clean. Mobile: `tsc --noEmit` clean.

**Database integration test - now verified live:** once Docker was back
up, ran `internal/store/users_integration_test.go` against a real
PostgreSQL 16 instance with TASK-002's migrations applied fresh -
Create + FindByGoogleID round-trip correctly. Also smoke-tested the
actual running server (not just unit tests) with real env vars:
`POST /api/v1/auth/google` with a garbage token → 401, `GET /api/v1/me`
with no token → 401, with a garbage bearer token → 401. All as expected.

**Real end-to-end Google Sign-In - now confirmed by the product owner.**
First attempt failed twice, both fixed live:
1. `redirect_uri_mismatch` - Expo web's OAuth redirect URI
   (`http://localhost:8081`) wasn't registered in the Google Cloud OAuth
   client (only the `/api/v1/auth/google/callback` placeholder was).
   Product owner added it in Google Cloud Console.
2. `Failed to fetch` on the actual sign-in call - the CORS fix from
   TASK-001 only covered the health endpoint; `/auth/google` and `/me`
   had no CORS headers and no OPTIONS preflight handling. Fixed by
   adding a `withCORS` middleware in `main.go` wrapping every route.

After both fixes, sign-in with a real Google account
(`charlistiow@gmail.com`) succeeded end-to-end: session token issued,
and `GET /api/v1/me` with that token correctly returned the user's id.

Note: this CORS fix (`main.go`'s `withCORS` wrapper) was added *after*
the task was submitted to Harness, and Harness would not accept a
second submit call while status was still `submitted` (only
`picked_up` tasks can submit) - so Harness's stored test_results/notes
for this task predate the fix slightly. The code in this repo and this
file are the accurate, current record.

**Notes/caveats:**
- ORBIT session tokens are self-signed JWTs (HS256, 30-day expiry, no
  server-side session table) rather than a stored-session design - this
  keeps the already-approved database schema untouched (no sessions
  table was in the TASK-002 design) and matches the "server session
  token with limited validity" wording in the security design. Requires
  a `SESSION_SECRET` environment variable in production.
- Used Google's JWKS endpoint directly with the `golang-jwt` library
  already needed for session tokens, instead of the official
  `google.golang.org/api/idtoken` package - that package pulled in a
  large dependency tree and forced a Go toolchain upgrade (it requires
  Go >= 1.26, we're on 1.25.5), which conflicted with the MVP "no
  unneeded technology" principle. Verified against Google's real JWKS
  endpoint (see test results above).
- Client secret is not used anywhere (only the Client ID) - correct for
  ID token verification, which only ever needs the public signing keys.
- Added a minimal `GET /api/v1/me` protected endpoint (not in the
  original API design) purely to have something to prove the auth
  middleware actually gates access end-to-end - it's a reasonable stub
  that later tasks (Home, etc.) will likely reuse or replace.
