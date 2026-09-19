---
id: solution-002
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [solution-001]
---

# Solution: Server-computed Aggregation for Views & Insight

Solution id: `sol-1789132202.393096` (for requirement `req-1789124277.430494`).

## Candidates evaluated
1. **Server-computed Aggregation** — chosen. Home/Timeline/Calendar dan
   Insight (daily/weekly/monthly) dihitung penuh di backend; reminder dan
   daily review juga dievaluasi server-side.
2. Client-side Aggregation from Raw Data Sync — ditolak, menduplikasi
   business logic penting (aturan Consumption Bank, definisi insight) di
   dua tempat, berisiko tidak konsisten antar versi aplikasi.

Approved by Charli Stiow, 2026-09-11.

## Functional requirement (Phase 2)
`fr-1789132539.593635` — agregasi Home/Timeline/Calendar, insight
rule-based, dan reminder/daily-review sepenuhnya dihitung di server.
