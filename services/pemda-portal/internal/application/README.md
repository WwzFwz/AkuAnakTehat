# application

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penerapan event ke view hazard terbaru milik consumer.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `updater.go` | Terima event dan terapkan hanya versi yang lebih baru. |
| `ports.go` | ViewStore untuk update atomik dan query view. |

## Kontrak dan alur

- Handle(ctx,event); ApplyIfNewer(ctx,event); List(ctx,limit/cursor).

## Dependensi

- contract; store SQLite mengimplementasikan port application.

## Aturan penting

- Update payload dan version dalam satu transaksi SQLite.
- Versi lebih kecil/sama diabaikan; jangan menimpa view baru dengan replay lama.
- Tidak meng-commit offset di application; pengaturan transport berada di consumer.

## Langkah implementasi dan verifikasi

- Implementasikan upsert bersyarat versi.
- Verifikasi crash sebelum/sesudah transaksi dan restart dengan volume tetap.
