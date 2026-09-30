# postgres

[Peta repository](../../README.md)

Konfigurasi server PostgreSQL; bukan lokasi skema aplikasi.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** PostgreSQL 16.8, env per service, named volume, dan healthcheck tersedia di Compose. Persistence record setelah restart telah diverifikasi; tabel uji dibersihkan. Skema Aggregator belum dibuat. Lihat [hasil fondasi](../../docs/evidence/foundation/README.md).

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `postgresql.conf` | Konfigurasi server minimum bila diperlukan; hindari tuning tanpa pengukuran. |

## Kontrak dan alur

- Canonical-db hanya ada di store_net. Aggregator akan ditambahkan pada tahap jalur inti.

## Dependensi

- Docker Compose menyediakan volume dan kredensial; migrasi berada pada services/aggregator/migrations.

## Aturan penting

- Client-api dan consumer tidak menerima kredensial DB atau akses network store.
- Tidak menduplikasi schema.sql/migrasi di folder infra.
- max_connections harus cukup untuk total pool terbatas serta koneksi administrasi.

## Langkah implementasi dan verifikasi

- Pilih konfigurasi minimum yang diperlukan; file config boleh tidak dibuat jika default memadai.
- Verifikasi isolasi jaringan dan persistence.
