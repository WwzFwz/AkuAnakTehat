# api

[Peta repository](../../README.md)

Tempat menyepakati kontrak integrasi sebelum anggota mengimplementasikan komponen masing-masing.

**Pemilik rencana:** A/B/C. **Tahap:** Baseline / pendukung baseline.

**Status:** empat kontrak baseline tersedia untuk review bersama. Kontrak Aggregator, event, port, dan storage mendefinisikan batas tahap jalur inti; keberadaan dokumen tersebut bukan bukti persetujuan anggota atau server/producer/consumer sudah berjalan. Kontrak mock, token, dan client menyertai implementasi awal fondasinya.

## Dokumen kontrak

| File | Tanggung jawab |
| --- | --- |
| [aggregator-http.md](aggregator-http.md) | Endpoint/filter/cursor/DTO/error API internal dan autentikasinya. |
| [consumers.md](consumers.md) | Group, SQLite, retry/DLQ, endpoint internal dan konfigurasi consumer. |
| [hazard-event.md](hazard-event.md) | Field envelope event, semantik version/event_id/correlation_id, dan contoh payload lengkap. |
| [aggregator-ports.md](aggregator-ports.md) | UnitOfWork/Tx, HazardQuery, dan OutboxStore beserta semantik error. |
| [storage.md](storage.md) | Kepemilikan data, kolom, constraint, dan indeks acuan migrasi. |
| [client-http.md](client-http.md) | Kontrak API downstream, proyeksi dan raw-field rejection. |
| [token-http.md](token-http.md) | Request/response token dan refresh pada fondasi. |
| [mock-http.md](mock-http.md) | Kontrak sumber, admin, kredensial, dan semantik since. |
| [reliability.md](reliability.md) | Batas ukuran, karantina, retry storage, checkpoint, freshness, dan trace outbound. |

## Kontrak dan alur

- Empat titik temu utama: skema Aggregator, envelope Kafka, HTTP internal, dan port di dalam Aggregator.
- Ingest, query, relay, tiga consumer, serta API downstream telah diimplementasikan. Bukti pengujian berada di `docs/evidence/`; kontrak saja tidak menggantikan verifikasi runtime.

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
