# aggregator

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Aggregator: pemilik Canonical Store, ingest, query internal, dan outbox.

**Pemilik utama:** A. **Port internal:** 9000. Jalur ingest A sudah berjalan: mock → tolerant reader → canonicalize/korelasi → transaksi PostgreSQL → pending outbox. Relay Kafka (C) sudah aktif; tiga consumer tersedia terpisah. Query/API internal (B) belum diimplementasikan. [Bukti pengujian](../../docs/evidence/ingest/README.md).

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`internal/domain/hazard`](internal/domain/hazard/README.md) | A | Inti/pendukung | Model HazardEvent dan aturan murni yang dipakai ingest serta query. |
| [`internal/domain/tsunami`](internal/domain/tsunami/README.md) | A | Inti/pendukung | Model warning dan aturan korelasi ke gempa, tanpa akses jaringan atau storage. |
| [`internal/application/canonicalize`](internal/application/canonicalize/README.md) | A | Inti/pendukung | Pemetaan data sumber ke HazardEvent; pemilik tipe input yang digunakan adapter dan ingest. |
| [`internal/application/ingest`](internal/application/ingest/README.md) | A | Inti/pendukung | Orkestrasi pemetaan, korelasi, deteksi perubahan, penulisan atomik, dan kemajuan polling. |
| [`internal/application/query`](internal/application/query/README.md) | B | Inti/pendukung | Use case baca Canonical Store melalui API internal, beserta filter, halaman, dan status sumber. |
| [`internal/adapter/inbound/http`](internal/adapter/inbound/http/README.md) | B | Inti/pendukung | Transport API internal Aggregator: routing, autentikasi internal, validasi input, dan serialisasi hasil query. |
| [`internal/adapter/outbound/bmkg`](internal/adapter/outbound/bmkg/README.md) | A | Inti/pendukung | HTTP client dan tolerant decoder untuk BMKG; tidak berisi aturan pemetaan severity. |
| [`internal/adapter/outbound/pvmbg`](internal/adapter/outbound/pvmbg/README.md) | A | Inti/pendukung | HTTP client dan tolerant decoder untuk PVMBG; tidak berisi aturan pemetaan severity. |
| [`internal/adapter/outbound/sourcehttp`](internal/adapter/outbound/sourcehttp/README.md) | A | Inti/pendukung | Transport HTTP lokal service, timeout, batas respons, dan penolakan redirect. |
| [`internal/adapter/outbound/postgres`](internal/adapter/outbound/postgres/README.md) | A/B/C | Inti/pendukung | Implementasi penyimpanan Aggregator, dengan pemisahan kepemilikan file untuk transaksi ingest, query, dan outbox. |
| [`internal/adapter/outbound/kafka`](internal/adapter/outbound/kafka/README.md) | C | Inti/pendukung | Producer ke topic event kanonik; tidak mengetahui daftar atau alamat consumer. |
| [`internal/worker/poller`](internal/worker/poller/README.md) | A | Inti/pendukung | Penjadwal polling independen per sumber, dengan satu jalur penulisan per sumber. |
| [`internal/worker/outbox`](internal/worker/outbox/README.md) | C | Inti/pendukung | Relay polling event tersimpan dan housekeeping baris terkirim; tetap di dalam Aggregator. |
| [`migrations`](migrations/README.md) | A | Inti/pendukung | Skema berversi Canonical Store; dimiliki dan dijalankan oleh Aggregator. |
| [`reference`](reference/README.md) | A | Inti/pendukung | Referensi statis nama serta koordinat gunung api milik BNPB. |
| [`cmd/aggregator`](cmd/aggregator/README.md) | A | Inti/pendukung | Composition root aggregator; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | A | Inti/pendukung | Konfigurasi lokal aggregator; nama variabel berikut adalah usulan yang perlu disepakati. |
| [`internal/observability`](internal/observability/README.md) | A | Inti/pendukung | Logging terstruktur, correlation ID, serta liveness/readiness milik service. |

## Berkas tingkat service yang direncanakan

| File | Tanggung jawab |
| --- | --- |
| `go.mod` | Satu module mandiri untuk service ini; versi Go dipilih dan dipin saat implementasi. |
| `Dockerfile` | Build dengan konteks folder service sendiri dan runtime minimal non-root. |
| `.dockerignore` | Mengecualikan secret, artefak lokal, dan berkas yang tidak diperlukan saat build. |

## Batas dan urutan kerja

- Tidak meng-import business logic atau DTO dari module service lain.
- Interface kecil didefinisikan oleh package pemakai; concrete adapter dirangkai melalui composition root.
- Sepakati kontrak → implementasikan jalur inti → hubungkan dependensi → jalankan skenario tugas → kumpulkan bukti.
- Timeout lokal, batas konkurensi yang relevan, health, log terstruktur, dan correlation ID termasuk baseline.
- Cache, LISTEN/NOTIFY, schema_observations, dan propagasi deadline lewat header adalah tambahan; jangan menjadikannya prasyarat fungsi inti.
- Service dapat dimulai sebagai proses sendiri; kesiapan dependensi dilaporkan oleh readiness. Jangan mengembalikan sukses palsu untuk fitur yang belum dibuat.

## Menjalankan dan melanjutkan

`docker compose up -d --build --wait aggregator` menjalankan migrasi embedded dan dua worker sumber. Jalankan mock melalui `make up` atau Compose root agar polling berhasil. Tidak ada port host Aggregator; koneksi hanya pada network source/store/edge/bus.

- `/health`: liveness. `/ready/ingest`: DB dapat diakses.
- `/ready` dan `/internal/hazards...` masih 503 `query_not_implemented`, sehingga client-api tidak menganggap query sudah siap.
- BMKG 2s, PVMBG 5s, jitter maksimum 10%, overlap 10s. HTTP timeout BMKG 1s/PVMBG 4s; transaksi DB 2s. Satu siklus tidak overlap dengan siklus berikutnya.
- Checkpoint kosong mengambil seluruh histori seed. Berikutnya memakai waktu mulai request sukses dikurangi overlap; mock warning harus menafsirkan since sebagai waktu perubahan dan jam sumber harus selaras.
- Deployment tahap ini satu instance Aggregator. Tidak ada koordinasi leader polling untuk banyak replica.
- `source_endpoint_status` menyimpan hasil setiap endpoint; satu endpoint BMKG sukses tidak menutupi kegagalan endpoint lain.

B dapat menggunakan model `hazard.Event`/`Record` dan skema migrasi untuk repository query. C membaca snapshot JSON immutable di `outbox.payload`, lalu mengisi `outbox.published_at` hanya setelah ACK broker. Relay berjalan dengan polling 1s, batch 100, pool DB terpisah 2 koneksi, dan housekeeping published >24h. Lihat [bukti event](../../docs/evidence/events/README.md).

Pengujian: `go test ./...` dan `go vet ./...` dari module ini; `make ingest-check` dari root pada POSIX, atau perintah Compose/Go pada [bukti ingest](../../docs/evidence/ingest/README.md).
