# Discovery Index (Phase 1 derived knowledge)

## Problems

| ID | Statement | Root cause status |
|---|---|---|
| prob-1789120596.494444 | User tidak menyadari pola konsumsi harian mereka sendiri | done — see [rootcause-001](rootcause-001-five-whys-session.md) |
| prob-1789120596.494511 | Aplikasi budgeting/accounting yang ada terlalu kaku dan punitif | done — see [rootcause-001](rootcause-001-five-whys-session.md) |
| prob-1789120596.494591 | User tidak punya cara memahami keterkaitan mood dengan pengeluaran | done — see [rootcause-001](rootcause-001-five-whys-session.md) |

## Assumptions

| ID | Assumption |
|---|---|
| assume-1789120596.494629 | Pendekatan non-punitif/fleksibel meningkatkan konsistensi pemakaian |
| assume-1789120596.494642 | Kategori/item user-defined tanpa normalisasi meningkatkan adopsi |
| assume-1789120596.494672 | Consumption Bank sebagai buffer psikologis meningkatkan fleksibilitas tanpa menyesatkan |

## Personas

| ID | Name |
|---|---|
| persona-1789124212.211663 | Rani - Mahasiswa Semester 5 |
| persona-1789124212.211689 | Dimas - Pekerja Muda Tahun Pertama |

## Needs

| ID | Statement | Persona |
|---|---|---|
| need-1789124221.396522 | Melihat pola pengeluaran tanpa wajib mencatat semua/dihukum | Rani |
| need-1789124221.396553 | Alat budgeting fleksibel, bukan kaku, agar termotivasi jangka panjang | Dimas |
| need-1789124221.396564 | Melihat kaitan temporal mood dan pengeluaran, bukan klaim sebab-akibat | Rani/Dimas |

## Features

| ID | Name |
|---|---|
| feat-1789124241.162749 | Flexible Daily Baseline & Plan vs Actual |
| feat-1789124241.162778 | Consumption Bank |
| feat-1789124241.162795 | Non-punitive Reminder & Daily Review |
| feat-1789124241.162801 | Mood & Spending Timeline |
| feat-1789124241.162809 | User-defined Categories & Quick Input |
| feat-1789124241.162819 | Rule-based Insight (Daily/Weekly/Monthly) |
| feat-1789124241.162829 | Home, Timeline, dan Calendar Views |
| feat-1789124241.16285 | Google Sign-In Account |

## Functional Requirements

| ID | Feature | Summary |
|---|---|---|
| req-1789124277.430426 | Flexible Daily Baseline & Plan vs Actual | Baseline harian sebagai target lunak, plan-vs-actual, histori immutable |
| req-1789124277.430456 | Consumption Bank | Ledger, floor 0, cap 10%/transaksi, badge tier |
| req-1789124277.430465 | Non-punitive Reminder & Daily Review | Reminder opsional, review akhir hari bisa dilewati |
| req-1789124277.430473 | Mood & Spending Timeline | Mood timestamped, korelasi temporal saja |
| req-1789124277.43048 | User-defined Categories & Quick Input | Tanpa fuzzy matching/normalisasi |
| req-1789124277.430488 | Rule-based Insight | Template-based, tidak menyimpulkan jika data kurang |
| req-1789124277.430494 | Home/Timeline/Calendar Views | Ringkasan cepat, riwayat kronologis, kalkulasi kalender otomatis |
| req-1789124277.430501 | Google Sign-In Account | Login wajib, satu profil per akun, IDR-only, timezone device |

All submitted 2026-09-11, analysis authored by claude-code based on
`orbit-app-concept.md` + user clarification rounds (see
[[interview-001]]).
