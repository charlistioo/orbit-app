---
id: interview-001
date: 2026-09-11
product_id: my-app
phase: 1
source: user
related: []
---

# Interview 001 — ORBIT concept doc + clarification round

## Raw notes

Target user kami adalah mahasiswa, pelajar, dan pekerja muda yang memiliki
penghasilan atau uang saku tetapi kurang sadar terhadap pola konsumsi harian
mereka, terutama pengeluaran kecil yang berulang seperti kopi, jajan,
makanan, dan self-reward. Masalahnya: mereka tidak tahu berapa banyak uang
yang mereka keluarkan untuk konsumsi sehari-hari, konsumsi apa yang paling
sering mereka lakukan, kapan pengeluaran mereka meningkat, seberapa sering
mereka melakukan self-reward, dan apakah pengeluaran mereka masih sesuai
dengan kemampuan finansial yang mereka tentukan sendiri. Secara individual
setiap transaksi kecil ini terlihat remeh, tapi jika terus dilakukan tanpa
disadari, akumulasinya menjadi kebocoran finansial yang signifikan.

Aplikasi budgeting atau accounting yang ada saat ini biasanya terlalu kaku,
mengunci pengeluaran, memberi penalti atau notifikasi yang menghakimi
(seperti "kamu boros"), atau mewajibkan pencatatan setiap transaksi.
Pendekatan seperti ini membuat banyak orang berhenti memakai aplikasi
tersebut karena terasa seperti mengerjakan administrasi keuangan, bukan
sesuatu yang ringan dan berkelanjutan untuk dipakai sehari-hari.

User tetap ingin bebas membelanjakan uangnya sesuai keinginan dan mood
mereka, tanpa merasa dikunci atau dihakimi, tetapi mereka juga ingin
memiliki kesadaran (awareness) terhadap ke mana uang mereka pergi dan
bagaimana pola konsumsi mereka berubah dari waktu ke waktu, termasuk
kaitannya dengan perubahan mood atau kondisi harian mereka.

### Klarifikasi tambahan (round 2)

1. Consumption Bank tidak boleh bersaldo negatif — floor di 0. Overload
   tetap dicatat penuh apa adanya (tidak didiskon); pemakaian saldo bank
   adalah entri ledger terpisah, bukan pengurang nilai transaksi. Realisasi
   di bawah rencana cukup masuk insight harian biasa. Ditambahkan fitur
   badge/tier dari saldo bank saat ini: Frugal (100rb), Thrifty (200rb),
   Economical (300rb), Collector (400rb), Master (500rb+).
2. IDR-only, timezone mengikuti device user (bukan fixed server timezone).
3. Wajib login via Google Sign-In. Satu akun Google = satu profil data,
   tanpa fitur sync multi-device eksplisit.
4. Insight untuk MVP bersifat rule-based/template, bukan LLM-generated.

## My analysis (problems & assumptions submitted to Harness)

Submitted and accepted by Harness on 2026-09-11.

Problems:
- `prob-1789120596.494444` — user tidak menyadari pola konsumsi harian
  mereka (pengeluaran kecil berulang menjadi kebocoran finansial)
- `prob-1789120596.494511` — aplikasi budgeting/accounting yang ada
  terlalu kaku/punitif sehingga user berhenti memakainya
- `prob-1789120596.494591` — user tidak punya cara memahami keterkaitan
  mood dengan perubahan pola pengeluaran tanpa klaim kausalitas berlebihan

Assumptions:
- `assume-1789120596.494629` — pendekatan non-punitif/fleksibel akan
  membuat user lebih konsisten memakai aplikasi
- `assume-1789120596.494642` — kategori/item user-defined tanpa
  normalisasi otomatis akan meningkatkan adopsi
- `assume-1789120596.494672` — Consumption Bank sebagai buffer psikologis
  akan meningkatkan fleksibilitas tanpa menyesatkan kondisi keuangan
  aktual user

Next: 5 Whys per problem to find root cause (see `harness-notes/discovery/`
once complete).

Interviewer: Charli Stiow
