---
id: solution-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [rootcause-001]
---

# Solution: Flexible Baseline + Consumption Bank

Solution id: `sol-1789127578.796544` (for requirement `req-1789124277.430426`).

## Candidates evaluated
1. **Flexible Baseline + Psychological Buffer (Consumption Bank)** — chosen.
   Baseline harian sebagai target lunak, plan-vs-actual, Consumption Bank
   sebagai buffer psikologis dengan ledger, floor 0, cap 10%/transaksi.
2. Category-based Soft Envelope Budgeting — ditolak, terlalu mirip pola
   budgeting kaku yang jadi akar masalah di [[rootcause-001]].
3. Pure Passive Logging + Retrospective Analytics — ditolak, kehilangan
   elemen kesadaran-saat-itu-juga (plan-vs-actual real-time).

Approved by Charli Stiow, 2026-09-11.

## Functional requirement (Phase 2)
`fr-1789128519.964188` — mekanisme baseline lunak, plan-vs-actual, dan
Consumption Bank (floor 0, cap 10%, badge tier), dengan 4 acceptance
criteria (lihat requirements report Harness untuk teks lengkap).
