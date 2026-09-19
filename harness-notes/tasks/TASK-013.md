---
id: TASK-013
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-006]
---

# TASK-013: Reminder & daily review flow

**Objective:** Implement server-side detection of planned categories
without a recorded transaction, and POST /api/v1/daily-review to
finalize a day (including skip-without-penalty).

**Depends on:** TASK-006 (approved)

**Acceptance criteria:**
- Selecting "Lewati" during daily review still marks the day as valid
  with no penalty flag.
- Reminder text is neutral and never uses guilt-inducing language.

## Submission

**Files changed:**
- backend/migrations/0011_create_daily_reviews.up.sql, .down.sql (new)
  - `daily_reviews` table: `user_id`, `review_date`, `status` (CHECK
    IN 'completed'/'skipped'), unique on `(user_id, review_date)` so
    re-finalizing a date overwrites rather than duplicates.
- backend/internal/reminder/message.go (new) - pure, DB-free reminder
  text builder: `Message([]CategoryName) string`
- backend/internal/reminder/message_test.go (new) - 3 unit tests,
  including the direct proof of the "neutral language" acceptance
  criterion
- backend/internal/store/daily_review.go (new) - `PendingCategories`
  (planned categories for a date with no matching transaction) and
  `Finalize`/`GetByDate` for the review record itself
- backend/internal/store/daily_review_integration_test.go (new) - 2
  end-to-end tests against real PostgreSQL
- backend/internal/api/daily_review_handler.go (new) -
  `GetTodayReminder` (GET /api/v1/reminders/today) and
  `FinalizeDailyReview` (POST /api/v1/daily-review)
- backend/internal/api/daily_review_handler_test.go (new) - 7 handler
  unit tests with a fake store, including the direct proof of the
  "skip is valid, no penalty" acceptance criterion
- backend/main.go (wired both new routes)
- mobile/DailyReviewScreen.tsx (new)
- mobile/App.tsx (renders it post-login)

**Test results:**
- `internal/reminder`: 3/3 unit tests pass, no database involved.
  `TestMessage_NeverUsesGuiltInducingLanguage` is the direct proof of
  the second acceptance criterion - checks the generated text against
  a list of guilt/blame/urgency phrases ("kamu lupa", "seharusnya",
  "jangan lupa", "wajib", "terlambat", "gagal", etc.) and against "!"
  punctuation, and fails if any appear.
- `internal/api`: 7/7 new handler unit tests pass with a fake store.
  `TestFinalizeDailyReview_SkipIsValidNoPenalty` is the direct proof
  of the first acceptance criterion - asserts the response for a
  "skip" action has `status: "skipped"` and `valid: true` (there is no
  separate penalized/invalid state in the response shape at all).
  Also covers: no-pending -> empty message, pending -> non-empty
  message with category names, auth required on both endpoints,
  "complete" action, invalid action -> 400 without calling the store,
  missing date -> 400. Full package (40 tests) green.
- `internal/store`: 2/2 new integration tests against real PostgreSQL
  pass: a category with a same-day transaction is excluded from
  `PendingCategories` while one without is still listed; skipping a
  date persists status "skipped", and re-finalizing the same date as
  "completed" overwrites the same row (same id) rather than creating a
  second one. Full package (19 tests) green across 3 consecutive runs
  with 0 leftover rows afterwards.
- Live smoke test: rebuilt binary, real backend, real session token.
  `GET /reminders/today` returned an empty message (nothing pending
  for the live account that day). `POST /daily-review` with
  `action: "skip"` returned `{"status":"skipped","valid":true}`;
  immediately re-posting with `action: "complete"` for the same date
  returned `{"status":"completed","valid":true}` (confirms the
  overwrite behavior live, not just in the integration test);
  `action: "bogus"` correctly returned 400.
- `go build ./...`, `go vet ./...`, `npx tsc --noEmit` all clean.

**Notes/caveats:**
- "Penalty" isn't a field or state anywhere in the schema or API
  response - there was never a penalized status to avoid setting.
  `skipped` and `completed` are two equally-valid, equally-final review
  outcomes for a date; the response always reports `valid: true`
  regardless of which one was chosen, which is what the acceptance
  criterion actually asks for ("still marks the day as valid").
- The reminder message is intentionally the *only* place the wording
  is generated (`internal/reminder`, no DB dependency) so the
  no-guilt-language rule is enforced by a pure function anyone can unit
  test without touching a database - mirrors the same reasoning behind
  splitting `internal/insight` out in TASK-012.
- A category counts as "pending" only if it has a plan for that date
  and no transaction recorded for that date/category - a category with
  no plan at all isn't pending (nothing was expected of it, so there's
  nothing to remind about).
- `daily_reviews` has no relationship to Consumption Bank, badges, or
  any other reward/penalty mechanism in the schema - finalizing a day
  (either way) doesn't trigger any side effects elsewhere; it's purely
  a record of "the user acknowledged this date."
