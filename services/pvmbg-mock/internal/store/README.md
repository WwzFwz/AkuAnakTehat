# store

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penyimpanan in-memory PVMBG, dimiliki mock dan aman untuk akses bersamaan.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `memory.go` | Koleksi data, mutex, dan snapshot respons. |
| `repository.go` | Operasi insert dan query laporan berdasarkan since. |
| `seed.go` | Load fixture historis dengan ID tetap. |

## Kontrak dan alur

- ListReportsSince(since) dan AddReport(report).

## Dependensi

- domain lokal; dipakai generator dan handler HTTP.

## Aturan penting

- since inklusif; urutan hasil deterministik.
- Jangan memegang mutex saat menunggu delay atau menulis jaringan.
- Jangan mengembalikan slice/map mutable yang dapat diubah generator secara bersamaan.
- Filter laporan memakai reported_at; schema toggle tidak mengubah semua seed lama.

## Langkah implementasi dan verifikasi

- Implementasikan koleksi, load seed, dan snapshot dengan lock singkat.
- Verifikasi query bersamaan dengan generator tidak menimbulkan data race.
