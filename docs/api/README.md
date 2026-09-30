# api

[Peta repository](../../README.md)

Tempat menyepakati kontrak integrasi sebelum anggota mengimplementasikan komponen masing-masing.

**Pemilik rencana:** A/B/C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `aggregator-http.md` | Endpoint/filter/cursor/DTO/error API internal dan autentikasinya. |
| `hazard-event.md` | Field envelope event, semantik version/event_id/correlation_id, dan contoh payload lengkap. |
| `aggregator-ports.md` | UnitOfWork/Tx, HazardQuery, dan OutboxStore beserta semantik error. |
| `storage.md` | Peta kepemilikan data dan rujukan ke migrasi; bukan salinan DDL. |
| `client-http.md` | Kontrak API downstream, proyeksi dan raw-field rejection. |
| `token-http.md` | Request/response token dan refresh yang dipakai demo. |
| `mock-http.md` | Kontrak sumber, admin, kredensial, dan semantik since. |

## Kontrak dan alur

- Empat titik temu utama: skema Aggregator, envelope Kafka, HTTP internal, dan port di dalam Aggregator.
- README package menyatakan rencana; file kontrak rinci di atas belum dibuat.

## Dependensi

- A/B/C menyepakati bersama; DDL sebenarnya tetap milik migrations Aggregator.

## Aturan penting

- Setiap nama field/status harus konsisten lintas producer/client/consumer.
- Pisahkan kontrak publik dari metadata internal storage.
- Unknown field policy additive-only dinyatakan; perubahan destruktif perlu versi baru.
- Jangan memakai docs/api sebagai tempat shared business code.

## Langkah implementasi dan verifikasi

- Finalisasi empat kontrak utama sebelum parallel coding.
- Catat keputusan error, pagination, data freshness, dan transaksi yang belum final.
