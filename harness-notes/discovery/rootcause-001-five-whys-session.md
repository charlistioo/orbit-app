---
id: rootcause-001
date: 2026-09-11
product_id: my-app
phase: 1
source: claude-code
related: [interview-001]
---

# Root cause session — all 3 problems

## Problem 1 — prob-1789120596.494444
User tidak menyadari pola konsumsi harian mereka sendiri.

Why chain (mix of live Q&A with user + own reasoning):
1. Kadang tidak mencatat; kalau mencatat, tidak punya batasan pengeluaran.
2. Sebagian tidak kepikiran menetapkan batasan; sebagian mencoba tapi terasa kaku sehingga ditinggalkan.
3. Batasan berupa angka tetap yang tidak boleh dilanggar sama sekali.
4. Angka tetap itu ditetapkan sendiri oleh user, bukan karena tool memaksa.
5. Itu satu-satunya paradigma budgeting yang mereka tahu.

**Root cause:** Satu-satunya paradigma budgeting yang diketahui user adalah
batas angka tetap yang tidak boleh dilanggar sama sekali ("disiplin = never
exceed"). Begitu batas itu dilanggar, user merasa gagal dan meninggalkan
seluruh upaya pencatatan — bukan hanya melonggarkan batasnya. Ini langsung
memvalidasi filosofi baseline-fleksibel ORBIT.

## Problem 2 — prob-1789120596.494511
Aplikasi budgeting yang ada terlalu kaku dan punitif.

Why chain (reasoned independently, no ambiguity requiring user input):
1. Filosofi desain aplikasi finansial umum menyamakan kontrol ketat dengan
   kesuksesan produk — sejalan dengan paradigma kaku di Problem 1.
2. Aplikasi finansial secara historis dibangun oleh tim yang mengukur
   keberhasilan dari metrik kepatuhan/kontrol (compliance-first), bukan
   dari retensi jangka panjang berbasis pengalaman emosional pengguna.

**Root cause:** Aplikasi finansial yang ada dibangun compliance-first,
sehingga mekanisme retensi punitif (lock, streak wajib, notifikasi
menyalahkan) dipinjam dari pola aplikasi habit-tracking lain tanpa
penyesuaian ke domain pengeluaran uang yang sensitif secara emosional.
Tekanan semacam itu membuat user menghindari konfrontasi dengan data
mereka sendiri dan berhenti memakai aplikasi sepenuhnya.

## Problem 3 — prob-1789120596.494591
User tidak punya cara memahami keterkaitan mood dengan pengeluaran.

Why chain (reasoned independently):
1. Expense tracking dan mood tracking berkembang sebagai dua kategori
   produk terpisah dengan tujuan dan metrik keberhasilan berbeda.
2. Tidak ada insentif produk untuk menggabungkan keduanya - aplikasi
   finansial dinilai dari akurasi pencatatan, aplikasi mood dari kedalaman
   refleksi personal, sehingga tidak ada tim/produk yang menjembataninya.

**Root cause:** Tidak ada produk yang mencatat keduanya dalam satu timeline
yang sama karena tidak ada insentif produk untuk menggabungkannya, sehingga
user tidak pernah punya data gabungan untuk melihat korelasi temporal —
dan tanpa alat yang menyajikan ini sebagai korelasi (bukan kausalitas),
mereka juga berisiko salah menyimpulkan hubungan yang belum tentu benar.

**Catatan proses:** Harness ternyata mensyaratkan why-chain minimal 2
langkah sebelum status root cause berubah dari `unvalidated` ke
`validated` (baru diketahui setelah `validate` gagal meski root cause
sudah terisi di why-1 untuk Problem 2 dan 3) - sudah diperbaiki dengan
menambah satu why lagi untuk keduanya.

## Note on process
Problem 1 root cause digali lewat percakapan langsung dengan user (sebelum
kebijakan batching berlaku). Problem 2 dan 3 dinalar sepenuhnya sendiri dari
konteks yang sudah ada di `orbit-app-concept.md` dan Problem 1 — tidak ada
titik yang genuinely ambigu sehingga tidak perlu menunggu jawaban user.
