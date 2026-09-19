---
id: techstack-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [architecture-001]
---

# Tech Stack

All three approved by Charli Stiow, 2026-09-11.

| Category | Chosen | Alternatives considered | Why chosen |
|---|---|---|---|
| Mobile | React Native (Expo) + TypeScript | Flutter, Native (Kotlin+Swift) | Cross-platform satu codebase, cepat untuk MVP, sudah ditentukan pemilik produk |
| Backend | Go | Node.js (Express/Fastify), Python (FastAPI) | Performa baik, deployment sederhana, cocok untuk business logic yang jelas (Consumption Bank, insight) |
| Database | PostgreSQL | MySQL, MongoDB | Data ORBIT sangat relasional (user-plan-transaksi-ledger), butuh jaminan transaksi ACID untuk ledger |

## Failure modes & mitigations (required before approval)

| Category | Failure mode | Mitigation |
|---|---|---|
| Backend | Transaksi tercatat dua kali jika koneksi mobile putus-nyambung saat retry. | Idempotency key per transaksi dari client; backend mengabaikan percobaan ulang dengan key yang sama. |
| Database | Dua transaksi hampir bersamaan sama-sama lolos cek batas 10% Consumption Bank sebelum salah satu tersimpan (race condition). | Row lock (`SELECT ... FOR UPDATE`) saat membaca saldo untuk validasi, sebelum entri ledger baru disimpan dalam satu transaksi database. |
| Mobile | Pencatatan transaksi mengandalkan koneksi internet, padahal sering dipakai di tempat bersinyal buruk. | Draft transaksi disimpan lokal saat offline, dikirim ulang otomatis begitu koneksi kembali, memakai idempotency key yang sama. |
