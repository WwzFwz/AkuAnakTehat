# postgres

[Peta repository](../../README.md)

Konfigurasi server PostgreSQL; bukan lokasi skema aplikasi.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `postgresql.conf` | Konfigurasi server minimum bila diperlukan; hindari tuning tanpa pengukuran. |

## Kontrak dan alur

- Canonical-db hanya ada di store_net bersama Aggregator.

## Dependensi

- Docker Compose menyediakan volume dan kredensial; migrasi berada pada services/aggregator/migrations.

## Aturan penting

- Client-api dan consumer tidak menerima kredensial DB atau akses network store.
- Tidak menduplikasi schema.sql/migrasi di folder infra.
- max_connections harus cukup untuk total pool terbatas serta koneksi administrasi.

## Langkah implementasi dan verifikasi

- Pilih konfigurasi minimum yang diperlukan; file config boleh tidak dibuat jika default memadai.
- Verifikasi isolasi jaringan dan persistence.
