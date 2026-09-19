---
id: TASK-019
date: 2026-09-19
product_id: my-app
phase: 3
source: claude-code
status: submitted
related: [techstack-002, TASK-016, TASK-018]
---

# TASK-019: Web Profil/Settings - avatar upload, badge display, edit default budget

**Objective:** Build the full Profil tab: profile photo upload
(client-side resized, stored as base64 via the existing `PATCH
/me/profile` `avatar_data` field), a proper badge display (all 5
tiers, current tier highlighted), and an edit form for
`default_baseline_amount` - all against the existing, unmodified Go
backend.

**Depends on:** TASK-018 (approved). Final checkpoint of the
`techstack-002` pivot.

## Submission

**Files changed (`web/`, no backend changes needed):**
- app/profil/page.tsx (rewritten) - avatar tap-to-upload with
  client-side canvas resize (max 256x256, JPEG 85% quality) before
  sending, all 5 badge tiers displayed with the current one
  highlighted (from `profile.badge`, never recomputed client-side),
  inline edit for `default_baseline_amount`, sign-out kept from
  Checkpoint 1

**Test results:**
- `tsc --noEmit` clean; `next build` statically generates all 13 routes
  with no errors.
- Live smoke test against the real backend: `PATCH /me/profile` with
  both `avatar_data` (a real base64 JPEG payload) and
  `default_baseline_amount` in one call, confirmed both persisted via
  a follow-up `GET /me/profile`. Cleared the test avatar afterward so
  the product owner sees a clean upload prompt rather than my test
  image.

**Notes/caveats:**
- Avatar resize happens entirely client-side via `<canvas>` before the
  `PATCH` call - the backend never receives an oversized image,
  directly implementing the mitigation recorded against this risk in
  `techstack-002`.
- The 5-tier badge row always reflects the server's own `badge` string
  (exact match against `BADGE_TIERS` names) - no threshold math
  duplicated client-side, consistent with every other badge display in
  this project.
- Editing the default budget only ever calls `PATCH /me/profile`; it
  does not touch today's already-created plan (if any) - matches the
  "pre-fill, not auto-apply" design decided back in TASK-016.
- This completes all 4 checkpoints of the Next.js/HeroUI pivot
  (`techstack-002`): scaffold/auth/onboarding, core loop, Bank/
  Riwayat/Wawasan, and now Profil/Settings.
