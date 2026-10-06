# Review 3B dan verifikasi integrasi M1

Tanggal: 6 Oktober 2026. Branch `origin/feat/alfan` pada `29a01bc` diambil melalui
fetch dan fast-forward ke `main`, lalu diperbaiki dengan commit `fix(...)` per
komponen. Hasil di bawah berasal dari stack lokal yang benar-benar dijalankan.

## Temuan review yang diperbaiki

| Komponen | Masalah sebelumnya | Perbaikan dan verifikasi |
| --- | --- | --- |
| Aggregator query | `limit=0` berubah menjadi default; malformed query dapat diabaikan parser. Port repository juga belum menormalisasi filter. | Tolak batas HTTP di luar 1..500 dan query rusak, normalisasi di repository; unit test serta integrasi filter/cursor. |
| Client-api raw | Decode ke `any` memakai float64 sehingga integer besar dapat berubah. | `UseNumber`, tolak trailing JSON; test nilai tepat `9007199254740993`. |
| CLI build | Module CLI tidak memiliki `go.mod` dan tidak diperiksa CI. | Module mandiri, workspace, serta test/vet semua service dan tools. |
| CLI kredensial | Default flag secret dari env ikut tercetak pada help; redirect HTTP dapat meneruskan body login. | Secret dibaca sesudah parsing, redirect ditolak, URL dasar divalidasi; tes help dan redirect 307. |
| CLI sesi | 401 terlambat dapat membatalkan token baru; mode CLI satu request tidak membuktikan sesi melewati expiry. | Invalidasi hanya token yang digunakan request, refresh serial, mode `--watch 65s --count 2`; uji TTL alami. |
| Load test | Skenario seismic belum membebani volcanic bersamaan; 429 mengencerkan denominator error; jumlah VU belum membuktikan koneksi. | Skenario paralel, metrik auth/bisnis terpisah, hanya sukses untuk latency, sampling TCP ESTABLISHED. |
| Observability | Log HTTP Aggregator pada debug tidak muncul; penyelesaian consumer tidak mencatat trace. | Log HTTP info, latency outbound/publish, `record_completed`, serta helper pencarian correlation ID. |
| Consumer readiness | `franz-go v1.18.1` Ping dapat menunggu handshake melewati deadline, sehingga `/ready` timeout ketika Kafka mati. | Batasi waktu tunggu caller dan satu probe yang belum selesai per consumer. Test broker tanpa respons: deadline 50 ms selesai sekitar 50 ms; sebelumnya sekitar 10 s. |
| Pengujian | Ingest masih mengharapkan migration v1; subscriber lama dapat membuat replay tampak lulus tanpa menguji group baru. | Harapkan migration v2, uji group pemda unik dengan store kosong, lalu pulihkan konfigurasi normal. |

`AGENTS.md` tidak dilacak Git. Aturan ignore juga mencakup folder anak dan variasi
huruf besar/kecil. Secret hasil bootstrap tetap hanya berada pada file lokal yang
diabaikan Git.

## Bukti pengujian

| Pemeriksaan | Bukti | Cakupan |
| --- | --- | --- |
| Go unit dan vet | [modules.txt](modules.txt) | Bootstrap serta sembilan module: delapan service dan CLI. |
| Consumer setelah perbaikan deadline | [consumer-readiness.txt](consumer-readiness.txt) | Tiga consumer, retry/DLQ/commit, deadline broker yang tidak merespons, dan penolakan probe paralel. `go vet` juga dijalankan. |
| PostgreSQL nyata | [postgres.txt](postgres.txt) | Transaksi ingest, watermark/outbox, serta query dalam schema uji terisolasi. |
| Regresi seluruh stack | [regression.txt](regression.txt) | Fondasi, restart dependensi, ingest, event, query, expiry, dan rebuild; lulus 234,280 s. |
| Subscriber baru | [events-fresh-group.txt](events-fresh-group.txt) | Ulang jalur event dengan group unik dan store kosong; tidak mengganti Aggregator. |
| Query awal | [query-runtime.txt](query-runtime.txt) | Eksekusi awal suite query; seluruh tes lulus sebelum penguatan trace dan readiness. |
| Trace event nyata | [event-trace.jsonl](event-trace.jsonl) | Satu correlation ID pada mock BMKG, poller/outbox Aggregator, dan ketiga consumer. |
| Kondisi akhir | [services.txt](services.txt) | Status container setelah seluruh pengujian. |

Hasil gagal pertama dipertahankan pada
[regression-before-readiness-fix.txt](regression-before-readiness-fix.txt):
`TestEventPipeline` timeout pada `/ready` dashboard ketika broker dihentikan.
Kasus itu diperbaiki, lalu seluruh regresi dijalankan ulang; hasil gagal tersebut
bukan status kode akhir. File output dari PowerShell dikonversi ke UTF-8 tanpa
mengubah hasil pengujian. Wrapper stderr PowerShell pada output build PostgreSQL
tidak menunjukkan kegagalan test; lihat `PASS` dan hasil test pada akhir file.

Pengukuran awal k6 di Docker juga dipertahankan pada
[load-docker-invalid/](load-docker-invalid/). Walaupun threshold k6 tidak gagal,
output mencatat durasi HTTP negatif (contohnya minimum latency bisnis -541,10 ms
pada seismic/volcanic). **Output tersebut tidak dipakai untuk klaim performa.**
Penyebab lingkungan belum diisolasi secara pasti. Runner sekarang menolak durasi
negatif, dan pengukuran diulang menggunakan k6 0.57.0 native Windows terhadap
stack Docker yang sama. Arsip binary berasal dari rilis resmi dan SHA256-nya
dicocokkan dengan checksum rilis. Binary disimpan lokal dan tidak masuk Git.

## Hasil load test final

Load generator: k6 0.57.0 native Windows, Go 1.23.6. Target: Docker Desktop 28.0.4,
16 CPU tersedia dan RAM container 7.936.237.568 byte. Host: Ryzen 9 6900HS,
16 logical processor, RAM 16.388.050.944 byte, Windows 11 build 26300.
Lihat [lingkungan k6/Docker](load/environment.txt) dan [host.json](host.json).

PVMBG diatur delay tepat 3 s selama pengujian, lalu dikembalikan ke konfigurasi
normal. Seluruh threshold lulus dan tidak ada minimum durasi HTTP negatif.

| Skenario | Konfigurasi | Request bisnis /s | Sukses 200 /s | p50 / p95 / p99 sukses (ms) | Bisnis 429 | Auth 429 | Error bisnis di luar 429 |
| --- | --- | ---: | ---: | --- | ---: | ---: | --- |
| [seismic-only](load/seismic-only.json) | 10 VU seismic + 10 volcanic, 60 s | 191.47 | 109.44 | 5.06 / 8.41 / 10.52 | 4928 | 0 | 0.00% (0/6574) |
| [sustained](load/sustained.json) | 50 VU, 90 s | 473.04 | 107.54 | 5.89 / 12.49 / 22.82 | 32930 | 10 | 0.00% (0/9689) |
| [outage](load/outage.json) | 5 VU per fase, outage 20 s + recovery 30 s; 1 VU kontrol | 23.01 | 23.01 | 5.57 / 14.60 / 36.02 | 0 | 0 | 0.00% (0/1201) |

p95 **seismic saja 8.62 ms**, di bawah 300 ms saat volcanic berjalan
bersamaan. Tabel menampilkan percentile gabungan per skenario; throughput sukses
dibedakan dari semua respons bisnis, termasuk penolakan 429.

[Sampling TCP](load/connections.json) menunjukkan sedikitnya 50 koneksi ESTABLISHED
berturut-turut selama **90.46 s**. Pada sustained: 42619 respons bisnis,
9689 HTTP 200, dan 32930 penolakan bisnis 429.
429 cukup banyak karena seluruh VU memakai satu identitas dengan rate limit
bersama; angka ini tidak boleh disembunyikan atau dihitung sebagai sukses.

Pada fase outage, 459/474 respons (96,83%) sudah menampilkan status sumber
bermasalah. Pada fase recovery, 457/727 (62,86%) sudah kembali HEALTHY; sisanya
masih pada masa deteksi/pemulihan polling dan breaker. Respons bisnis tetap HTTP
200 dengan data tersimpan. Threshold recovery >50% lulus.

Error rate memakai denominator respons bisnis **selain 429**. Autentikasi
memiliki metrik sendiri; `auth_error_rate` final 0% di semua skenario. Latency
hanya memakai HTTP 200 dan mengikuti `response.timings.duration`, sehingga
penolakan cepat tidak memperbaiki p95. Durasi ini tidak mencakup pembentukan
koneksi; lihat [definisi metrik k6](https://grafana.com/docs/k6/latest/using-k6/metrics/reference/).

Pada JSON ekspor k6, nilai `thresholds: false` berarti threshold **tidak dilanggar**.
Pada `system_error_rate`, `passes` menghitung sampel boolean true (error),
bukan jumlah test lulus. Output teks tersedia di [seismic](load/seismic-only.txt),
[sustained](load/sustained.txt), dan [outage](load/outage.txt).

## Keterkaitan P1-P5

- **P1:** mapper/tolerant reader diuji unit; schema v2 masuk saat runtime, field
  `confidence_level` terbaca melalui raw API, dan schema drift dicatat. Record
  lama dan baru tetap dapat dibaca.
- **P2:** seismic dan volcanic dibaca bersamaan saat mock PVMBG delay 3 s; sustained
  50 koneksi diukur. Outage PVMBG menghasilkan penanda stale, seismic tetap sehat,
  dan recovery berlangsung tanpa restart aplikasi. Outage Kafka tidak menghentikan
  ingest; backlog outbox dikirim setelah broker pulih.
- **P3:** tiga identitas memakai kredensial berbeda. Media ditolak ketika meminta
  raw/field terlarang. Satu sesi CLI melewati TTL access 60 s secara alami, melakukan
  refresh, dan token lama tetap ditolak. Tidak ada pengubahan jam sistem.
- **P4:** notifier di-stop, dibangun ulang, dan dijalankan sendiri; baca API tetap
  berhasil dan timestamp start container lain tidak berubah. Evolusi JSONB tidak
  mengubah versi migrasi atau mengganti container. Akses langsung Canonical Store
  tetap hanya tersedia bagi Aggregator pada konfigurasi aplikasi Compose.
- **P5:** consumer memiliki group dan SQLite terpisah. Replay, dedup, DLQ, stop dan
  catch-up diverifikasi. Group pemda baru membaca histori dari store kosong tanpa
  perubahan producer. Test meninggalkan group uji unik dan event sintetis sebagai
  artefak demo; group serta volume pemda reguler dipulihkan.

## Mengulang pengujian

Ikuti [panduan pemeriksaan](../../../scripts/check/README.md) dan
[panduan load test](../../../scripts/loadtest/README.md). Gunakan Go 1.24.2,
bootstrap default TTL 60 s, dan jalankan suite secara berurutan. Tes mengubah
simulasi mock serta menghentikan/memulai dependensi sementara; jangan dijalankan
bersamaan dengan demo lain. Volume aplikasi tidak dihapus.

## Batas hasil

- Lingkungan tunggal Docker Desktop; Kafka satu broker dengan replication factor
  1. Hasil ini tidak membuktikan high availability atau kapasitas multi-host.
- Load memakai pola closed loop dan satu identitas Tim Lapangan dengan rate limit
  bersama. Koneksi ESTABLISHED tidak berarti semua request aktif bersamaan.
- Notifier menggunakan pengiriman log simulasi. Dedup persisten diuji, tetapi crash
  tepat setelah pengiriman dan sebelum pencatatan dapat menghasilkan pengiriman
  ulang; tidak ada klaim exactly-once untuk side effect eksternal.
- Crash pada batas publish/ACK dan DLQ diuji melalui unit serta outage/replay;
  bukan fault injection di setiap instruksi proses.
- Hasil GitHub Actions hosted belum dikonfirmasi. Laporan akhir/PDF, gladi demo,
  tag `milestone-1`, dan pengumpulan masih pekerjaan tersendiri.
- Singleflight/micro-cache, `LISTEN/NOTIFY`, `schema_observations`, dan propagasi
  deadline melalui header tetap fitur tambahan yang belum diimplementasikan.
