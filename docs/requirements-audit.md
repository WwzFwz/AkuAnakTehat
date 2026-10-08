# Audit persyaratan dan dokumentasi M1

Pemeriksaan dilakukan pada 8 Oktober 2026. Bukti regresi stack mengacu pada revisi `2353479`; perbaikan instrumentasi HTTP diverifikasi terpisah melalui tes module Client API. Acuan adalah dokumen M1 dan dokumen terpusat yang diberikan kelompok; identitas dokumennya tercatat pada [manifest sumber laporan](laporan/assets/input-documents.json). README rencana diperlakukan sebagai catatan desain awal, bukan tambahan persyaratan tugas.

**Kesimpulan audit ini** adalah jalur fungsional utama P1 hingga P5 tersedia dan memiliki bukti pengujian lokal. Tidak ditemukan komponen wajib yang hilang hanya karena nama file berbeda dari rancangan. Temuan instrumentasi retry HTTP client-api sudah ditutup melalui log per percobaan dan tes terarah. Kelengkapan penulisan serta pengumpulan belum selesai. Karena itu belum tepat menyatakan seluruh persyaratan sudah terpenuhi tanpa catatan.

## Ketentuan umum

| ID | Kewajiban | Implementasi dan bukti | Status |
| --- | --- | --- | --- |
| U1 | Komunikasi jaringan tanpa library bisnis bersama lintas service | Module per service, adapter HTTP, dan Kafka. [Regresi](evidence/reliability-2026-10-08/regression.txt) menjalankan rantai sumber, Aggregator, API, dan consumer. | Tersedia; jaringan diuji lokal |
| U2 | Satu pemilik tiap storage | PostgreSQL dimiliki Aggregator, Redis dimiliki Auth Service, SQLite lokal tiap consumer. [Compose](../docker-compose.yml) dan adapter client-api menunjukkan pembacaan kanonik melalui HTTP. | Tersedia; akses administratif test bukan akses service bisnis |
| U3 | Dockerfile per service dan satu orkestrasi | Delapan service mempunyai Dockerfile mandiri; Compose dan bootstrap dari [README utama](../README.md). | Tersedia |
| U4 | Service independen dan kegagalan dipicu nyata | TestPersistenceAndDependencyRecovery, TestIngestPipeline, TestEventPipeline, dan TestIndependentRebuild pada regresi. | Diuji lokal; demo sinkron tetap dilakukan kelompok |
| U5 | Kredensial instansi terpisah dan kredensial silang ditolak | Mock memakai hash kredensial berbeda dan format header berbeda; TestMockContracts memeriksa penolakan silang. | Diuji lokal |
| U6 | Secret dari konfigurasi yang tidak di-commit dan .env.example tersedia | [Generator](../scripts/secrets/generate.go), [.gitignore](../.gitignore), [.env.example](../.env.example), test log, dan scan histori. | Implementasi tersedia; scan revisi pengumpulan tetap mengikuti commit yang diperiksa |
| U7 | Health, log terstruktur, correlation ID, latency tiap panggilan keluar | Health tersedia pada tiap service; source HTTP, PostgreSQL, Redis, dan Kafka sudah memiliki instrumentasi. Client-api mencatat setiap percobaan HTTP dengan correlation ID, nomor percobaan, durasi, status, dan kategori hasil. | Tersedia; G1 ditutup dengan tes module, lihat bukti di bawah |

## Kontrak mock dan data

| Persyaratan | Lokasi implementasi | Pemeriksaan |
| --- | --- | --- |
| Endpoint BMKG terpisah untuk gempa dan warning | [Router BMKG](../services/bmkg-mock/internal/http/router.go), [adapter BMKG](../services/aggregator/internal/adapter/outbound/bmkg/client.go) | TestMockContracts dan TestIngestPipeline |
| Seed minimal 20 record dan generator paling lambat tiap 10 s | `seed/`, `internal/generator/`, dan `internal/config/` masing-masing mock | Suite fondasi memeriksa seed dan pertambahan data; config menolak interval lebih dari 10 s |
| BMKG delay tetap 50 sampai 150 ms | [Konfigurasi BMKG](../services/bmkg-mock/internal/config/config.go) | Default 100 ms dan batas konfigurasi eksplisit |
| PVMBG delay acak, schema v2, dan outage runtime | [Konfigurasi](../services/pvmbg-mock/internal/config/config.go), [router](../services/pvmbg-mock/internal/http/router.go), [state](../services/pvmbg-mock/internal/simulation/state.go) | Default 500 ms sampai 3 s; test admin, drift, serta outage error dan hang tersedia |
| Pemetaan kanonik, ambang magnitude, warning, dan referensi volcano | [Mapper](../services/aggregator/internal/application/canonicalize/mapper.go), [uji mapper](../services/aggregator/internal/application/canonicalize/mapper_test.go), [referensi](../services/aggregator/reference/volcanoes.json) | Uji unit dan transaksi PostgreSQL mencakup korelasi warning serta perubahan severity |
| Field tambahan tetap tersedia sebagai attributes | [Decoder](../services/aggregator/internal/application/canonicalize/input.go), mapper, dan JSONB | TestDynamicSchemaAndStaleAPI memeriksa record lama dan baru tanpa restart atau DDL |
| Tiga client berbeda, read-only, TTL default 60 s, refresh otomatis | Generator secret, Auth Service, client-api, dan CLI | TestQueryIntegration dan TestNaturalExpiryAndFieldCLI |

Nilai referensi gunung api sintetis, interpretasi `since` warning sebagai waktu perubahan, batas ukuran payload, dan keterlambatan di luar overlap tetap merupakan asumsi implementasi. Lihat [kontrak ketahanan](api/reliability.md). Skenario data valid di luar asumsi tersebut tidak otomatis tercakup oleh kelulusan dataset demo.

## Problem P1 hingga P5

| Kriteria | Bukti implementasi dan pengujian | Batas kesimpulan |
| --- | --- | --- |
| P1.1 sampai P1.5 | Decoder, mapper, JSONB, log schema_drift; TestIngestPipeline dan TestDynamicSchemaAndStaleAPI | Field aditif didukung; penghapusan atau perubahan tipe field wajib tidak otomatis didukung |
| P2.1 | [Load terbaru](evidence/reliability-2026-10-08/README.md), p95 seismic 10,92 ms ketika request volcanic bersamaan dan PVMBG delay 3 s | Berlaku pada host dan pola beban yang dicatat |
| P2.2 | 50 TCP selama 88,53 s, error non-429 0%, tidak ada crash; throughput dan p50/p95/p99 dilaporkan | 78,91% request sustained ditolak dengan 429; throughput sukses sekitar 100,20/s |
| P2.3 | Seismic 480/480 sehat, volcanic 480/480 dengan data basi, recovery HEALTHY sebelum pengembalian konfigurasi mock | Outage k6 20 s dan recovery 30 s; belum menjadi bukti outage 10 sampai 20 menit |
| P2.4 | [Bagian konkurensi laporan](laporan/sections/07-p2-konkurensi.tex), pool terpisah, timeout, breaker, limiter dan runner | Pemulihan tidak instan; kapasitas maksimum belum diukur |
| P3.1 dan P3.2 | TestMockContracts, TestTokenRotationAndAuthorization, TestQueryIntegration | Pembatasan field dilakukan server, bukan sekadar UI |
| P3.3 | TestNaturalExpiryAndFieldCLI menunggu TTL alami dalam sesi 65 s dan memeriksa token lama ditolak | Access JWT yang belum kedaluwarsa tidak dicabut seketika hanya karena refresh; ini batas desain yang dinyatakan |
| P3.4 dan P3.5 | Konfigurasi lokal, scan histori, serta [pembahasan keamanan](laporan/sections/08-p3-autentikasi.tex) | TLS internal, Kafka SASL, dan revocation access token instan tidak diklaim tersedia |
| P4.1 sampai P4.4 | TestIndependentRebuild, TestDynamicSchemaAndStaleAPI, ownership Compose, serta perbandingan storage dalam laporan | Satu host, satu penulis Aggregator, dan satu broker; belum HA atau koordinasi banyak replica |
| P5.1 sampai P5.5 | TestEventPipeline menguji publish broker, consumer independen, catch-up, dan subscriber baru dengan group serta store kosong | Retensi Kafka membatasi replay; notifier mempunyai celah duplikasi antara pengiriman dan marker |

[Regresi terbaru](evidence/reliability-2026-10-08/regression.txt) lulus dalam 306,951 s. [Pengujian PostgreSQL](evidence/reliability-2026-10-08/postgres.txt) memeriksa batas envelope dan backlog. [Pengujian module](evidence/reliability-2026-10-08/modules.txt) meliputi unit test serta vet. Transkrip tersebut berasal dari pengujian perubahan sebelum commit dan dipertahankan sebagai bukti, bukan dijalankan ulang ketika README diperbarui.

## Status temuan audit

### G1. Latency tiap percobaan HTTP client-api

**Ditutup.** [Client.fetchAttempt](../services/client-api/internal/adapter/outbound/aggregator/client.go) menghasilkan satu log `upstream_request` untuk setiap pemanggilan `HTTP.Do` oleh adapter, termasuk pembacaan dan decode body pada percobaan itu. Log memuat correlation ID, nomor percobaan, durasi, status HTTP atau 0 bila tidak ada respons, serta kategori hasil tetap. URL, kredensial, isi respons, dan pesan error mentah tidak dicatat.

[Pengujian retry](../services/client-api/internal/adapter/outbound/aggregator/attempt_test.go) menggagalkan percobaan pertama lalu mengizinkan percobaan kedua berhasil. Dua log, deadline bersama, sanitasi, batas retry, pembatalan, timeout, dan penutupan body diperiksa. Seluruh tes module Client API dan go vet lulus. [Bukti U7](evidence/http-attempts-2026-10-08/README.md) mencatat perintah dan cakupannya; regresi stack serta k6 tidak dijalankan ulang untuk perubahan instrumentasi ini.

### G2. Informasi laporan dan penyelesaian kelompok

Kontribusi penulisan kedua anggota telah dikonfirmasi mencakup keseluruhan laporan. Deklarasi penggunaan LLM disediakan sebagai placeholder untuk diisi sendiri oleh anggota. Status hosted CI, demo sinkron, revisi atau tag pengumpulan, dan formulir belum boleh dinyatakan selesai hanya berdasarkan tes lokal. Daftar terperinci tetap berada di [review laporan](laporan/REVIEW.md). Placeholder ini merupakan informasi yang belum tersedia, bukan komponen Go yang belum dibuat.

## Fitur opsional dan perubahan pembagian file

Singleflight, micro-cache, LISTEN/NOTIFY, schema_observations, serta propagasi deadline melalui header belum diimplementasikan. Spesifikasi tidak mewajibkan mekanisme tersebut; timeout lokal, polling relay, dan log drift sudah menjalankan jalur dasar. `internal/cache` secara eksplisit merupakan folder dokumentasi opsi pengembangan, bukan package runtime.

Nama file rencana seperti `batch.go`, `event_envelope.go`, `hazard_service.go`, dan `claims.go` tidak diwajibkan oleh spesifikasi. Batch tersedia di canonicalize/input.go; envelope di ingest/service.go; use case dan port client-api di application/service.go; claims di authn/verifier.go. README komponen kini menunjuk berkas aktual dan menjelaskan tanggung jawabnya. Pemecahan file baru hanya diperlukan bila membantu pemeliharaan, bukan untuk mengejar daftar rencana lama.
