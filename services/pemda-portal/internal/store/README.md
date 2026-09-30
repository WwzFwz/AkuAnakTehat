# store

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penyimpanan view SQLite persisten milik pemda-portal.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `sqlite.go` | Buka koneksi, konfigurasi timeout, dan lifecycle SQLite. |
| `repository.go` | ApplyIfNewer serta query view. |
| `schema.sql` | Tabel hazard_view berisi hazard_id, version, payload, updated_at. |
| `embed.go` | Embed skema lokal untuk initialization. |

## Kontrak dan alur

- Implementasi ViewStore milik application.
- ApplyIfNewer(ctx,event) atomik dan memberi tahu apakah versi diterapkan; List membaca view lokal.

## Dependensi

- application/contract serta driver SQLite murni Go yang dipilih saat implementasi.

## Aturan penting

- File SQLite berada pada named volume khusus service; tidak dipakai bersama consumer lain.
- Payload dan version diperbarui atomik sebelum consumer commit offset.
- Gunakan query berparameter dan batas hasil.
- Jika volume hilang atau retention terlewati, strategi rebuild diperlukan; earliest saja tidak memulihkan semua keadaan.

## Langkah implementasi dan verifikasi

- Definisikan skema minimum dan path database.
- Verifikasi view tetap ada setelah proses/container restart.
