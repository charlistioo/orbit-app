---
id: TASK-004
date: 2026-09-16
product_id: my-app
phase: 3
source: harness
status: approved
related: [database-001, api-001, TASK-003]
---

# TASK-004: Categories & items CRUD with quick-select

**Objective:** Implement category/item creation and listing (most
recent first), with no fuzzy matching or normalization, plus the
mobile quick-select UI.

**Depends on:** TASK-003 (approved)

**Acceptance criteria:**
- A category typed by the user is stored and returned exactly as typed.
- Previously created categories/items appear as quick-select options on
  the next transaction entry.

## Submission

**Files changed:**
- backend/internal/store/categories.go, items.go
- backend/internal/store/categories_integration_test.go
- backend/internal/api/categories_handler.go, categories_handler_test.go
- backend/internal/auth/middleware.go (added `ContextWithUserID` test helper)
- backend/main.go (wired 4 new routes)
- mobile/CategoryQuickSelect.tsx (new)
- mobile/App.tsx (embeds the quick-select component post-login)

**Test results:**
- 5 new unit tests with fakes: category name stored exactly as typed
  (including leading/trailing whitespace and punctuation - no
  trimming/normalization), categories listed most-recent-first, a
  user's category list is scoped to them (never another user's data),
  creating an item under a category id owned by a different user is
  rejected with 404 before any item is created, items listed
  most-recent-first.
- Integration test against a real PostgreSQL database (with TASK-002's
  schema): create user -> create category (weird name preserved
  exactly) -> verify ownership -> list -> create item -> list items,
  all round-tripping correctly.
- Live smoke test against the actual running server with a real
  session token (for the real Google account from TASK-003):
  `POST /categories` with `"  Kopi Susu!! "` stored and returned with
  that exact whitespace/punctuation intact; item created and listed
  under it; requesting items under a made-up category id correctly
  returned 404.
- `go vet` and `tsc --noEmit` clean.

**Notes/caveats:**
- Added a `CategoryStore.Exists` ownership check (not explicitly asked
  for in the acceptance criteria, but required by the Work Order's
  standing security note: never trust a client-supplied id without
  scoping it to the authenticated user) - without it, a user could
  reference another user's `category_id` in the URL and either read an
  empty list or, worse, create items silently attached to someone
  else's category.
- The quick-select UI is embedded directly in the post-login screen for
  now, since there's no Home/transaction-entry screen yet (that's a
  later task) - it's a self-contained component
  (`CategoryQuickSelect.tsx`) that a future task can drop into the real
  entry flow without rework.
- Left the backend and mobile web app running so the product owner can
  try the quick-select UI live before approving, same pattern as the
  last two tasks.
