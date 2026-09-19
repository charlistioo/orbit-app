---
id: techstack-002
date: 2026-09-17
product_id: my-app
phase: 2
source: claude-code
related: [techstack-001, TASK-015]
---

# Tech stack: Next.js + Tailwind + HeroUI (frontend_web)

Selection id: `techsel-1789664787.076312`, category `frontend_web`.
Approved by Charli Stiow, 2026-09-17.

## Why this exists

The product owner asked to use [HeroUI](https://heroui.com/) for a more
elegant UI. HeroUI is a web-only component library (React DOM + Tailwind
CSS) and cannot run inside React Native at all - there is no way to
satisfy that request while staying on the approved `techstack-001`
(React Native/Expo) selection. Presented as an explicit tradeoff via
AskUserQuestion; the product owner chose a full pivot to a web app over
keeping the native mobile build.

## Candidates evaluated

1. **Next.js + Tailwind CSS + HeroUI** - chosen. Reuses the existing Go/
   PostgreSQL REST API unchanged; gives the requested elegant, consistent
   UI via HeroUI's component set.
2. Keep React Native/Expo, restyle only - rejected, since HeroUI simply
   cannot run in that environment.

## Risks (pre-mortem, required before approval)

- Client-side `localStorage` session token is more exposed to XSS than a
  native app's storage. Mitigation: keep the existing stateless-JWT/
  30-day-expiry design (already limits blast radius), never render
  unsanitized user HTML; revisit httpOnly cookies later if needed.
- Avatar photos stored as base64 text directly on `users` (no dedicated
  file storage) could bloat rows for large images. Mitigation: resize/
  compress client-side before upload (cap ~256x256).
- Abandoning React Native/Expo means no native mobile build (APK/IPA)
  without a separate future effort. Mitigation: the Go backend and its
  full REST API are unchanged and reusable by any future native client;
  `mobile/` is left in place, not deleted.

## What does NOT change

The entire `backend/` Go + PostgreSQL API surface is reused as-is -
`/auth/google`, `/categories`, `/items`, `/plans`, `/transactions`,
`/consumption-bank`, `/consumption-bank/apply`, `/mood`, `/home`,
`/calendar`, `/insights/*`, `/reminders/today`, `/daily-review` all stay
exactly as built in TASK-001-014. Only small additive fields/endpoints
are needed for the new onboarding/profile screens (see TASK-016's Work
Order once proposed).

See `C:\Users\charl\.claude\plans\dapper-wandering-gem.md` for the full
implementation plan (screen-by-route mapping, sequencing checkpoints).
