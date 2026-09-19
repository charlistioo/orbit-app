---
id: TASK-001
date: 2026-09-12
product_id: my-app
phase: 3
source: harness
status: approved
related: [architecture-001, techstack-001, devplan-001]
---

# TASK-001: Project scaffolding

**Objective:** Set up Go backend project structure and React Native (Expo)
+ TypeScript mobile app skeleton, both able to run locally and talk to
each other over REST.

**Depends on:** none

**Acceptance criteria:**
- Backend runs locally and responds to a health check endpoint.
- Mobile app runs in Expo and can successfully call the backend health
  check.

**Approved technology:** backend = Go, database = PostgreSQL (from
architecture), mobile = React Native (Expo) + TypeScript (per
[[techstack-001]] and the task objective text).

**Security notes (from Work Order, apply going forward across all
tasks):**
- All queries for user-owned entities must filter by user_id from the
  session token, never a client-supplied parameter.
- Google ID token verified server-side on sign-in; all other endpoints
  validate an ORBIT session token.
- transactions and consumption_bank_ledger are create/read only - no
  update/delete endpoints.
- OAuth client secret only in backend env vars, never shipped to mobile.

## Submission

**Files changed:**
- backend/go.mod
- backend/main.go
- backend/.gitignore
- mobile/package.json
- mobile/app.json
- mobile/tsconfig.json
- mobile/index.ts
- mobile/App.tsx
- mobile/.gitignore

**Test results:**
- Backend builds (`go build`) and responds to `GET /api/v1/health` with
  `{"status":"healthy"}`.
- Mobile app type-checks cleanly (`tsc --noEmit`), Expo web bundler starts
  and serves the app (HTTP 200, JS bundle compiles ~1.7MB).
- End-to-end: a fetch call identical to the one in `App.tsx` (calling
  `http://localhost:8080/api/v1/health`) successfully retrieves the
  backend's health status while both processes run locally.

**Notes/caveats:**
- `create-expo-app` (the official scaffolding CLI) failed in this
  environment due to an npm `pack --dry-run` JSON-parsing error unrelated
  to the project itself; scaffolded the equivalent file set by hand
  instead (same template structure: App.tsx, index.ts, app.json,
  tsconfig.json, package.json).
- Added a permissive CORS header to the health endpoint to support local
  web-based testing of the mobile client; harmless for the native
  (iOS/Android) target platform.
- `react-native` version was pinned to the Expo-recommended version via
  `expo install` after an initial mismatch warning.

## Post-submission incident (live testing by product owner)

Product owner tested the app manually and got "Failed to reach backend"
even though both processes were individually healthy. Root cause: the
mobile web preview (`localhost:8081`) calling the backend (`localhost:8080`)
is a cross-origin request, and the browser enforces CORS - the backend
needed an `Access-Control-Allow-Origin` header, which was already in the
submitted source but wasn't active in the running process because `pkill`
(a Unix command) silently failed to terminate the old Windows process, so
a stale pre-CORS binary kept serving port 8080. Fixed by force-killing via
PowerShell (`Stop-Process -Force`) and rebuilding. Confirmed working by
the product owner after the fix - no code change beyond what was already
submitted was needed, since the CORS header was already in the
originally-submitted `main.go`.
