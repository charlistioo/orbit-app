---
id: database-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [techstack-001]
---

# Database Schema (PostgreSQL)

9 entities, id `dbdesign-1789132089.811476`.

1. **users** — id, google_id (unique), email, display_name, created_at.
2. **categories** — id, user_id (fk), name, created_at. Index (user_id, name).
3. **items** — id, user_id (fk), category_id (fk), name, created_at.
4. **daily_plans** — id, user_id (fk), plan_date, baseline_amount, created_at. Unique (user_id, plan_date).
5. **plan_categories** — id, daily_plan_id (fk), category_id (fk), planned_amount, updated_at.
6. **plan_adjustments** — id, plan_category_id (fk), old_amount, new_amount, changed_at. (Preserves plan history — immutability requirement.)
7. **transactions** — id, user_id (fk), category_id (fk), item_id (fk, nullable), plan_category_id (fk, nullable), amount, plan_amount_snapshot (nullable — plan value in effect at record time), occurred_at, created_at. Index (user_id, occurred_at). Append-only (no update/delete endpoints).
8. **consumption_bank_ledger** — id, user_id (fk), delta_amount, reason, related_transaction_id (fk, nullable), balance_after (check >= 0), created_at. Index (user_id, created_at). Ledger is source of truth, not a single balance field.
9. **mood_entries** — id, user_id (fk), mood (very_good/good/neutral/stressed/sad), recorded_at. Index (user_id, recorded_at).

Data lifecycle: transactions and ledger entries are append-only/immutable —
plan changes never rewrite historical rows.
