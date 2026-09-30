# store

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penyimpanan in-memory BMKG, dimiliki mock dan aman untuk akses bersamaan.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `memory.go` | Koleksi data, mutex, dan snapshot respons. |
| `repository.go` | Operasi insert/update event dan warning serta query since. |
| `seed.go` | Load fixture historis dengan ID tetap. |

## Kontrak dan alur

- ListSeismicSince(since), ListWarningsSince(since), AddSeismic(event), dan SaveWarning(warning).

## Dependensi

- domain lokal; dipakai generator dan handler HTTP.

## Aturan penting

- since inklusif; urutan hasil deterministik.
- Jangan memegang mutex saat menunggu delay atau menulis jaringan.
- Jangan mengembalikan slice/map mutable yang dapat diubah generator secara bersamaan.
- Filter warning memakai internal_modified_at yang berubah saat eskalasi, bukan estimated_arrival.

## Langkah implementasi dan verifikasi

- Implementasikan koleksi, load seed, dan snapshot dengan lock singkat.
- Verifikasi query bersamaan dengan generator tidak menimbulkan data race.
