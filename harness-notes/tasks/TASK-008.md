---
id: TASK-008
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-003]
---

# TASK-008: Mood tracking

**Objective:** Implement POST /api/v1/mood to log a timestamped mood
entry, and the mobile mood-selection UI.

**Depends on:** TASK-003 (approved)

**Acceptance criteria:**
- Multiple mood entries can be logged in the same day, each with its
  own timestamp.

## Submission

**Files changed:**
- backend/internal/store/mood.go, mood_integration_test.go
- backend/internal/api/mood_handler.go, mood_handler_test.go
- backend/main.go (wired POST /api/v1/mood)
- mobile/MoodSelector.tsx (new)
- mobile/App.tsx (embeds the mood selector post-login)

**Test results:**
- 3 unit tests with fakes: a valid mood is accepted and stored, an
  invalid mood value (not one of the 5 allowed) is rejected 400 before
  reaching the store, three moods logged in a row all create separate
  entries.
- Integration test against real PostgreSQL: three mood entries logged
  for the same user at different times of the same day, confirmed as 3
  distinct rows with distinct ids and timestamps.
- Live smoke test against the running server with a real session
  token: logged good/neutral/stressed in sequence, confirmed all 3 rows
  in the database with correct timestamps; an invalid mood value was
  correctly rejected with 400.
- `go vet` and `tsc --noEmit` clean.

**Notes/caveats:**
- No uniqueness constraint on (user, date) - mood can be logged any
  number of times per day, matching "mood is optional, can be updated
  whenever the user's state changes" from the product spec. This task
  only implements *logging*; using mood data in insight generation is
  a later task (rule-based insight engine).
- **Incident during live testing (my mistake, not a code issue):**
  while re-using the real Google account's user id from earlier
  sessions to mint a test token, I discovered it no longer existed - a
  `TRUNCATE ... CASCADE` I ran during TASK-007 to clean up leftover
  integration-test data accidentally wiped the entire local test
  database's `users` table, including the real account created via
  actual Google Sign-In and its "Food" category from TASK-004's live
  test. A new profile for the same Google account now exists (presumably
  re-created by a subsequent real sign-in), but that specific "Food"
  category is gone from this local test database. This is entirely
  contained to the disposable local Docker test database used for our
  manual verification - not production data - but flagging it plainly
  since it affects what you'll see if you open the mobile app again
  (you may need to sign in again and re-create categories to keep
  testing).
- Backend and mobile web app left running for you to try live if you
  want, same as the last few tasks.
