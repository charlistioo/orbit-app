---
id: TASK-016
date: 2026-09-17
product_id: my-app
phase: 3
source: claude-code
status: approved
related: [techstack-002, TASK-003]
---

# TASK-016: Web app scaffold - Next.js/Tailwind/HeroUI + auth + onboarding shell

**Objective:** Create the `web/` Next.js app (Tailwind + HeroUI +
lucide-react icons, no emoji), wire Google Sign-In via Google Identity
Services against the existing `/api/v1/auth/google` endpoint, build the
4-tab shell (Beranda/Riwayat/Wawasan/Profil), and the onboarding flow
(value-prop screen + Nama/Umur/Budget profile-setup screen for
first-time users). Add the small backend additions this needs.

**Depends on:** TASK-003 (approved) - reuses its `/auth/google` endpoint
unchanged.

## Submission

**Files changed:**
- backend/migrations/0012_add_user_profile_fields.up/down.sql (new)
- backend/internal/store/users.go (rewritten - `age`, `avatar_data`,
  `default_baseline_amount`, `onboarding_completed` fields;
  `GetByID`, `UpdateProfile` methods)
- backend/internal/store/users_integration_test.go (new test:
  `TestUserStore_UpdateProfileIsPartialAndCompletesOnboardingOnce`)
- backend/internal/api/profile_handler.go (new) - `GET`/`PATCH
  /api/v1/me/profile`
- backend/internal/api/profile_handler_test.go (new) - 6 unit tests
- backend/internal/api/auth_handler.go (`Me` is now a method with
  `onboarding_completed` in its response) + auth_handler_test.go (2 new
  tests)
- backend/main.go (wired the 2 new routes, `Me` now a method)
- web/ (new Next.js 16 / React 19 / TypeScript app):
  - app/providers.tsx, layout.tsx, globals.css (HeroUI + warm
    cream/terracotta/sage palette, per the new reference screens)
  - lib/api.ts, lib/auth.tsx (fetch wrapper + auth context, session
    token in localStorage)
  - components/AppShell.tsx (route-gating: unauthenticated -> onboarding,
    authenticated-but-not-onboarded -> profile setup, else the app),
    components/BottomNav.tsx (4 tabs)
  - app/onboarding/page.tsx (screen 1, value prop)
  - app/login/page.tsx (Google Sign-In)
  - app/onboarding/profile/page.tsx (screen 2a, Nama/Umur/Budget)
  - app/page.tsx, app/riwayat/page.tsx, app/wawasan/page.tsx,
    app/profil/page.tsx (Checkpoint 1 placeholders + a working sign-out
    on Profil - full dashboard/timeline/insight content is Checkpoints
    2-4)

**Test results:**
- Backend: `go build ./...`, `go vet ./...` clean; full `go test ./...`
  green (all packages, including 8 new profile/me tests + the new
  integration test against real PostgreSQL).
- Web: `tsc --noEmit` clean; `next build` compiles and statically
  generates all 8 routes with no errors.
- Live smoke test against the rebuilt, running backend with a real
  session token: `GET /me` correctly reports
  `onboarding_completed: false` for a fresh account; `GET /me/profile`
  returns the full profile shape; `PATCH /me/profile` with
  `display_name`/`age`/`default_baseline_amount` saves correctly and
  flips `onboarding_completed` to `true`; `GET /me` afterward reflects
  it. Reset the live test account back to `onboarding_completed: false`
  afterward so the product owner can walk through the real onboarding
  flow themselves.
- The Next.js pages themselves could not be click-through verified by
  me the way I verify a Go API (no browser automation available) - `curl`
  only shows the pre-hydration loading spinner, since routing/auth
  gating happens client-side. `next build`'s static generation succeeding
  for every route is the strongest automated signal available; visual/
  interactive confirmation needs the product owner opening a real
  browser, same limitation noted for Google Sign-In back in TASK-003.

**Notes/caveats:**
- **Action needed from you:** Google Identity Services enforces
  "Authorized JavaScript origins" per OAuth client, same as the Expo-web
  CORS/origin issue from TASK-001/003. The web app runs on
  `http://localhost:3000` - please add that as an Authorized JavaScript
  origin on the same Google Cloud Console OAuth client used before, or
  Sign-In will fail with a redirect/origin error.
- `default_baseline_amount` is a **pre-fill**, not an auto-created plan -
  it only feeds the (not-yet-built) daily "Rencanakan harimu" screen's
  suggested total; `POST /plans` still requires the user's confirmation
  each day, unchanged from TASK-005.
- `mobile/` (the React Native/Expo app) is untouched, still present,
  not deleted - it's simply no longer the active development target
  per the approved `techstack-002` pivot.
- Avatar upload UI doesn't exist yet (Checkpoint 4) - `avatar_data` is
  wired end-to-end on the backend (column + GET/PATCH) but nothing
  writes to it yet.
- Both dev servers are left running for you to try: backend on
  `:8080`, Next.js on `:3000` (`http://localhost:3000/onboarding`).
