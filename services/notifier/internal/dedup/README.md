# dedup

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Pencatatan pasangan hazard_id/version yang sudah selesai diproses notifier.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `sqlite.go` | Koneksi SQLite lokal. |
| `repository.go` | Pemeriksaan dan pencatatan processed. |
| `schema.sql` | Tabel processed dengan kunci gabungan hazard_id/version. |
| `embed.go` | Embed skema lokal. |

## Kontrak dan alur

- Memenuhi DedupStore: Seen(ctx,id,version) dan Record(ctx,id,version).

## Dependensi

- application; driver SQLite murni Go.

## Aturan penting

- Simpan pada volume notifier sendiri.
- Catat setelah sender berhasil; jangan menandai sebelum kirim.
- Pencatatan idempoten memakai unique constraint.
- Dedup tidak menghilangkan celah crash kirim→catat; tidak mengklaim exactly-once.

## Langkah implementasi dan verifikasi

- Implementasikan query dan insert idempoten.
- Verifikasi duplikat setelah restart tidak dikirim ulang jika record processed sudah tersimpan.
