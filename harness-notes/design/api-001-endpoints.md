---
id: api-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [database-001]
---

# API Design (REST, id apidesign-1789132120.21986)

19 endpoints, all auth-required, versioned `/api/v1`.

| Method | Path | Purpose |
|---|---|---|
| POST | /auth/google | Sign in with Google, create profile if new |
| GET | /categories | List categories, most recent first |
| POST | /categories | Create category |
| GET | /categories/{id}/items | List items under a category |
| POST | /categories/{id}/items | Create item |
| POST | /plans | Create/replace daily plan |
| PATCH | /plans/{plan_id}/categories/{plan_category_id} | Adjust planned amount (records plan_adjustment) |
| GET | /plans/{date} | Get plan for a date |
| POST | /transactions | Record a transaction (snapshots plan) |
| GET | /transactions | List transactions (Timeline) |
| GET | /consumption-bank | Get balance, badge, ledger |
| POST | /consumption-bank/apply | Apply bank toward an overspend |
| POST | /mood | Log mood entry |
| GET | /insights/daily | Today's insight |
| GET | /insights/weekly | This week's insight |
| GET | /insights/monthly | This month's insight |
| GET | /home | Home screen aggregate |
| GET | /calendar | Baseline vs actual per date + cumulative |
| POST | /daily-review | Submit end-of-day review outcome |
