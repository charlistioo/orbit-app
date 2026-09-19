---
id: architecture-001
date: 2026-09-11
product_id: my-app
phase: 2
source: claude-code
related: [solution-001, solution-002]
---

# Architecture: Modular Monolith

Style: `modular_monolith` (id `arch-1789128556.826899`).

Justification: pemilik produk eksplisit meminta tidak menggunakan
microservices untuk MVP; memprioritaskan business logic jelas dan data
bersih. Modular monolith memberi pemisahan tanggung jawab jelas tanpa
overhead operasional layanan terdistribusi.

## Components
1. **Auth & Identity** — Google Sign-In, sesi, satu akun = satu profil.
2. **Consumption Tracking** — baseline, daily plan, plan-adjustment,
   transaksi, kategori/item.
3. **Consumption Bank** — ledger, floor 0, cap 10%, badge tier.
4. **Mood Tracking** — mood entries dengan timestamp.
5. **Insight Engine** — insight rule-based harian/mingguan/bulanan.
6. **Reporting & Views** — Home, Timeline, Calendar.
7. **API Layer** — permukaan REST yang dikonsumsi mobile app.

## Architecture Decision Record
ADR-001 (id `adr-1789128566.219679`): pilih modular monolith bukan
microservices. Trade-off: tidak bisa di-scale/deploy independen per modul
tanpa refactor di masa depan, tapi dianggap dapat diterima untuk kebutuhan
MVP dan skala pengguna awal yang masih kecil.
