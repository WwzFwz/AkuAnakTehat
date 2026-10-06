# AkuAnakTehat

Rancangan repository platform koordinasi kebencanaan BNPB untuk IF4031 Milestone 1.

**Status saat ini: jalur 3A, 3B, dan 3C sudah terhubung.** Query internal, client-api, CLI Tim Lapangan, ingest, outbox, serta tiga consumer sudah memiliki pengujian unit dan integrasi. Bukti runtime dan load test terbaru ada di [verifikasi integrasi](docs/evidence/integration/README.md); hasil ini menjadi bahan laporan dan demo M1.

## Cara membaca

1. Pilih service pada tabel berikut.
2. Buka README komponen seperti `internal/store/`, `application/ingest/`, atau `authn/`.
3. Baca tujuan, rencana file, kontrak, dependensi, aturan, dan langkah verifikasinya sebelum coding.
4. Buat file kode ketika komponen mulai dikerjakan; jangan membuat seluruh placeholder kode sekaligus.

README hanya ditempatkan pada komponen yang punya tanggung jawab dan pada indeks service. Folder pengelompokan seperti `internal/`, `adapter/`, dan `outbound/` tidak memerlukan README tersendiri.

## Service

| Service | Pemilik utama | Tanggung jawab |
| --- | --- | --- |
| [aggregator](services/aggregator/README.md) | A | Aggregator: pemilik Canonical Store, ingest, query internal, dan outbox. |
| [client-api](services/client-api/README.md) | B | API downstream dengan verifikasi JWT, otorisasi, dan proyeksi. |
| [auth-service](services/auth-service/README.md) | B | Penerbit token dan pemilik auth-store. |
| [bmkg-mock](services/bmkg-mock/README.md) | A | Mock gempa dan warning tsunami dengan kontrak independen. |
| [pvmbg-mock](services/pvmbg-mock/README.md) | A | Mock vulkanik dengan delay, outage, dan schema evolution. |
| [dashboard-updater](services/dashboard-updater/README.md) | C | Consumer view hazard terbaru dengan SQLite persisten. |
| [notifier](services/notifier/README.md) | C | Consumer notifikasi SIAGA/AWAS dengan dedup persisten. |
| [pemda-portal](services/pemda-portal/README.md) | C | Consumer ketiga untuk demo subscription baru tanpa perubahan producer. |

A/B/C adalah pembagian kerja rancangan, belum nama anggota. Di Aggregator, A memegang ingest dan unit of work, B query/HTTP, dan C outbox/Kafka. Detail pemilik file ada di README komponennya.

## Folder pendukung

| Folder | Panduan |
| --- | --- |
| `docs/api/` | [Kontrak baseline untuk review bersama](docs/api/README.md) |
| `docs/evidence/` | Bukti [P1](docs/evidence/p1/README.md), [P2](docs/evidence/p2/README.md), [P3](docs/evidence/p3/README.md), [P4](docs/evidence/p4/README.md), [P5](docs/evidence/p5/README.md) |
| `infra/` | [Cara menjalankan fondasi](infra/README.md), [Kafka](infra/kafka/README.md), [PostgreSQL](infra/postgres/README.md), [Redis](infra/redis/README.md) |
| `scripts/` | [Secret bootstrap](scripts/secrets/README.md), [demo](scripts/demo/README.md), [load test](scripts/loadtest/README.md), [trace](scripts/trace/README.md) |
| `tools/field-cli/` | CLI HTTP mandiri untuk Tim Lapangan |
| `env/keys/` | [Penyimpanan kunci lokal](env/keys/README.md); secret hasil bootstrap diabaikan Git |
| `.github/workflows/` | [CI fondasi](.github/workflows/README.md) |

## Batas arsitektur

- Satu service = satu module Go dan container; antarservice berkomunikasi melalui jaringan.
- Canonical Store hanya diakses langsung Aggregator. Client-api membaca lewat API internal Aggregator.
- Poller, mapper, transaksi ingest, API internal, dan relay outbox tetap berada dalam Aggregator.
- Setiap consumer memiliki group Kafka serta penyimpanan SQLite sendiri.
- Auth-service memiliki Redis dan private key. Client-api hanya memiliki public key untuk verifikasi.
- Tidak ada shared business package atau database bersama antarservice.

## Empat kontrak sebelum jalur inti

1. A: [skema Canonical Store](docs/api/storage.md), acuan DDL pada [migrations](services/aggregator/migrations/README.md).
2. C: [envelope event Kafka](docs/api/hazard-event.md).
3. B: [HTTP internal Aggregator](docs/api/aggregator-http.md): route, filter, cursor, response, dan error.
4. A/B/C: [port internal Aggregator](docs/api/aggregator-ports.md): UnitOfWork/Tx, HazardQuery, dan OutboxStore.

Dokumen tersedia di [docs/api](docs/api/README.md), termasuk signature dan contoh payload. Implementasi port/skema A, query B, dan Store/Publisher C sudah tersedia. Perubahan kontrak perlu diselaraskan dengan seluruh pemakai.

## Urutan implementasi

1. Sepakati kontrak, buat module/service entrypoint, Compose, konfigurasi, health, dan log.
2. Bangun mock → polling → ingest → database → API, termasuk timeout lokal dan proteksi dasar P2.
3. Implementasikan identitas client, JWT, pembatasan field Media, dan refresh otomatis.
4. Bangun outbox polling → Kafka → consumer persisten, termasuk restart, catch-up, dan DLQ.
5. Jalankan seluruh skenario P1–P5 dan isi bukti nyata.
6. Jika waktu cukup, tambahkan singleflight/micro-cache, LISTEN/NOTIFY, schema_observations, dan propagasi deadline melalui header.

Fitur tambahan diberi label dalam README terkait. Timeout lokal bukan fitur tambahan.

## Menjalankan dan memeriksa fondasi

Prasyarat: Go 1.24.2, Docker dengan Compose v2, dan Python 3 untuk runner load test. Di PowerShell:

```powershell
powershell -NoProfile -File scripts/secrets/generate.ps1
docker compose config --quiet
docker compose up -d --build
```

Linux/macOS dengan Make dapat memakai `make up`. Detail port dan kredensial lokal ada di [panduan infra](infra/README.md). Client-api membaca query internal Aggregator melalui network Compose. Ingest Aggregator dan jalur baca berjalan independen dari mock saat membaca data tersimpan.

Pemeriksaan lokal: `powershell -NoProfile -File scripts/check/check.ps1` atau `make check` pada shell POSIX.

Setelah stack aktif, jalankan `go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m` dari root (`make smoke` pada POSIX). Suite ini mengubah simulasi mock sementara dan me-restart infrastruktur; jalankan saat tidak ada demo lain. Detail cakupan dan hasil ada di [bukti fondasi](docs/evidence/foundation/README.md).

| File | Isi saat ini |
| --- | --- |
| `go.work` | Sembilan module: delapan service dan CLI Tim Lapangan. |
| `docker-compose.yml` | Delapan aplikasi (pemda pada profile demo), PostgreSQL, Redis, Kafka, init topic, serta profile test/tools. |
| `Makefile` | Shortcut bootstrap, up/down, unit/vet, serta pemeriksaan fondasi, ingest, event, query, dan load. |
| `.env.example` | Variabel konfigurasi beserta placeholder non-secret. |
| `.gitignore` | Abaikan env/kunci/runtime artefak; wajib tersedia sebelum generator secret dijalankan. |

README komponen yang belum diimplementasikan tetap memuat rencana file dan konfigurasi.

## Status verifikasi

Hasil aktual dan batas cakupan dicatat di [verifikasi integrasi](docs/evidence/integration/README.md), termasuk perbaikan review branch `feat/alfan`, pengujian seluruh module, transaksi PostgreSQL, alur event/query, expiry token alami, rebuild komponen independen, dan metrik k6. Jalankan suite runtime **berurutan** karena beberapa test mengubah simulasi sumber atau menghentikan dependensi sementara. Perintah tersedia di [panduan pemeriksaan](scripts/check/README.md).

Eksekusi GitHub Actions di server belum dikonfirmasi dari sesi ini. Penyusunan laporan akhir, gladi demo, tag `milestone-1`, dan pengumpulan tetap pekerjaan terpisah. Fitur tambahan tetap belum menjadi bagian baseline.

## Dasar dan bantuan penulisan

Panduan mengikuti rancangan arsitektur serta struktur repository M1 yang telah direview dalam percakapan, termasuk koreksi terakhir pengguna. Dokumentasi awal dibantu OpenAI Codex; tim tetap perlu meninjau kontrak, memahami implementasi, dan melengkapi deklarasi penggunaan LLM pada laporan.
