# auth

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Validasi kredensial domain PVMBG.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `credentials.go` | Load hash kredensial dari konfigurasi. |
| `middleware.go` | Validasi bearer token pada data dan X-Admin-Key pada admin. |

## Kontrak dan alur

- Data memakai Authorization: Bearer pvmbg_<token>; admin memakai X-Admin-Key.

## Dependensi

- config lokal dan HTTP handler.

## Aturan penting

- Perbandingan hash dilakukan constant-time; kredensial tidak dipakai lintas domain.
- Kredensial salah/missing menghasilkan401/403 sesuai kontrak.
- Jangan mencatat nilai secret pada log atau pesan error.

## Langkah implementasi dan verifikasi

- Uji kredensial valid, salah, kosong, dan kredensial instansi lain dalam format header endpoint tujuan.
