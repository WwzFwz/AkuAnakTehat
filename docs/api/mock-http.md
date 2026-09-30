# Kontrak HTTP mock sumber

Status: kontrak dan implementasi fondasi tersedia. Implementasi baru diperiksa kompilasinya; pengujian perilaku, race, serta integrasi Docker belum dilakukan. Tidak ada ingest BNPB pada tahap ini.

Kedua mock adalah module Go mandiri dan tidak berbagi DTO/business logic. Semua record adalah data demo sintetis. Setiap restart memuat ulang seed tetap; data runtime hanya ada di memori dan ID runtime memakai prefix acak per proses.

## Konvensi

- Endpoint data mengembalikan array JSON (hasil kosong `[]`), `Content-Type: application/json`.
- `since` opsional: RFC3339/RFC3339Nano dengan timezone, inklusif (`timestamp >= since`); jika dihilangkan, semua record dikembalikan. Format tidak valid menghasilkan 400. Encode karakter `+` pada timezone sebagai `%2B`.
- Urutan respons mengikuti urutan insert yang deterministik. Warning yang diperbarui tetap pada posisi awalnya; tidak menjanjikan urutan waktu modifikasi.
- Kesalahan handler berbentuk `{"error":"pesan"}`. Kredensial hilang, salah, atau milik domain lain menghasilkan 401; route/metode tak dikenal memakai respons standar `net/http` (404/405).
- `X-Correlation-ID` diteruskan ke respons dan log bila cocok dengan `[a-zA-Z0-9._:-]{1,128}`; nilai kosong/tidak valid diganti ID acak. Log JSON memuat method, path, status, latency_ms, correlation_id tanpa credential atau query string.
- `GET /health` memeriksa proses dan tidak memerlukan kredensial. `GET /ready` memeriksa kesiapan melayani data; PVMBG mengembalikan 503 selama simulasi outage, sedangkan `/health` tetap 200. Admin tetap dapat mematikan outage.

## BMKG

Default `HTTP_ADDR=:8081`. Endpoint data memerlukan `X-BMKG-Key: <credential>`; server mencocokkan SHA-256 credential dengan `BMKG_KEY_HASH` menggunakan perbandingan constant-time.

`GET /seismic-events?since=<timestamp>` memfilter `occurred_at`.

| Field | Tipe |
| --- | --- |
| event_id | string |
| magnitude, depth_km | number |
| epicenter_lat, epicenter_lon | number |
| region_name | string |
| occurred_at | timestamp RFC3339 |
| potential_tsunami | boolean |

`GET /tsunami-warnings?since=<timestamp>` memfilter waktu perubahan internal (`internal_modified_at`, tidak diserialisasi), termasuk saat severity warning lama meningkat. Field `estimated_arrival` bukan watermark. Aggregator perlu memakai waktu mulai request sebagai watermark dan overlap dalam asumsi jam host bersama.

| Field | Tipe |
| --- | --- |
| warning_id, related_event_id | string |
| threat_level | Waspada / Siaga / Awas |
| affected_zones | array of string |
| estimated_arrival | timestamp RFC3339 |

Seed: 20 gempa dan 4 warning terkait, pada 1 September 2026 UTC. Gunakan query tanpa `since` atau initial lookback yang mencakup seed. Generator menambahkan gempa setiap interval, termasuk skenario warning sebelum/sesudah event dan eskalasi dengan ID warning tetap. Warning hanya merujuk event berpotensi tsunami; pada skenario warning lebih awal, event terkait tersedia pada tick berikutnya.

Delay tetap pada endpoint data yang lolos autentikasi dan validasi: `BMKG_FIXED_DELAY=100ms`, rentang diizinkan 50–150ms.

## PVMBG

Default `HTTP_ADDR=:8082`. Endpoint data memerlukan `Authorization: Bearer pvmbg_<opaque>`. SHA-256 **seluruh token termasuk prefix** dicocokkan dengan `PVMBG_TOKEN_HASH`. Endpoint admin memerlukan `X-Admin-Key`, dicocokkan dengan `ADMIN_KEY_HASH`. Hash data dan admin harus berbeda. Kredensial BMKG, PVMBG, dan admin dibangkitkan terpisah oleh bootstrap.

`GET /volcanic-reports?since=<timestamp>` memfilter `reported_at`.

| Field | Tipe |
| --- | --- |
| report_id, volcano_id | string |
| alert_level | Normal / Waspada / Siaga / Awas |
| eruption_count_24h | integer |
| ash_column_height_m | number |
| reported_at | timestamp RFC3339 |
| confidence_level | number 0–1; hanya record baru ketika skema v2 aktif |

Seed: 20 laporan skema v1 pada 1 September 2026 UTC. ID gunung yang dipakai adalah `VOLCANO-DEMO-01` dan `VOLCANO-DEMO-02`. Pemetaan nama/koordinat menjadi tanggung jawab referensi statis Aggregator, bukan mock.

`POST /admin/schema-version`: body `{"version":2}` mengaktifkan field baru, `{"version":1}` menonaktifkan untuk laporan berikutnya. Alternatif `{"enabled":true/false}` diterima. Body kosong atau `{}` melakukan toggle. Kedua selector sekaligus ditolak. Record lama tidak diubah, jadi v1 dan v2 dapat berdampingan tanpa restart.

`POST /admin/outage`: body `{"enabled":true,"mode":"hang"}` atau `{"enabled":true,"mode":"error"}`. Mode error menghasilkan 503; mode hang menunggu perubahan state atau pembatalan request. `{"enabled":false}` memulihkan layanan tanpa restart, termasuk membangunkan request hang. Body kosong atau `{}` toggle enabled; mode awal `error`, mode yang dihilangkan mempertahankan nilai sebelumnya. Body eksplisit disarankan untuk script yang dapat diulang. Admin dan health tidak terkena delay/outage.

Respons kedua admin: `{"outage":false,"mode":"error","schema_version":1}` sesuai state setelah perubahan. Body dibatasi 4 KiB; unknown field, JSON rusak, nilai selector tidak valid, atau beberapa dokumen JSON menghasilkan 400.

Delay endpoint data diacak antara `PVMBG_DELAY_MIN=500ms` dan `PVMBG_DELAY_MAX=3s`; untuk skenario P2 gunakan keduanya `3s`. Konfigurasi mengizinkan `0 <= MIN <= MAX <= 20s` untuk pengujian. Perubahan outage membangunkan request yang sedang delay. Tidak ada lock storage yang ditahan selama delay/hang atau penulisan HTTP.

## Konfigurasi dan lifecycle

| Variabel | Service | Default/aturan |
| --- | --- | --- |
| HTTP_ADDR | keduanya | :8081 / :8082 |
| GENERATION_INTERVAL | keduanya | 10s; harus >0 dan <=10s |
| BMKG_KEY_HASH | BMKG | wajib; 64 digit SHA-256 hex lowercase |
| BMKG_FIXED_DELAY | BMKG | 100ms; 50–150ms |
| PVMBG_TOKEN_HASH | PVMBG | wajib; 64 digit SHA-256 hex lowercase |
| ADMIN_KEY_HASH | PVMBG | wajib; terpisah dari token data |
| PVMBG_DELAY_MIN | PVMBG | 500ms |
| PVMBG_DELAY_MAX | PVMBG | 3s |

Konfigurasi salah menggagalkan startup, tanpa credential default. Tidak ada dependensi PostgreSQL/Redis/Kafka. Seed JSON di-embed ke binary. SIGTERM/interrupt membatalkan generator dan request, lalu memberi HTTP shutdown maksimal 5 detik. HTTP read-header timeout 5s, read timeout 10s, write timeout 30s, idle timeout 60s. Pada mode hang, client wajib memakai deadline agar request cepat dibatalkan; server juga membatalkan request saat shutdown.

Build masing-masing dari direktori service: `go build ./cmd/bmkg-mock` / `go build ./cmd/pvmbg-mock`. Docker build context juga direktori service masing-masing. Go toolchain dipin ke 1.24.2; fondasi ini hanya memakai standard library.

## Batas tahap ini

- Kompilasi tidak membuktikan skenario P1–P5 lulus.
- Uji semantik filter/update warning, credential lintas domain, schema toggle, outage recovery/cancellation, konkurensi/race dan load belum dijalankan.
- Storage in-memory bertambah sepanjang proses dan belum memiliki retensi; gunakan untuk demo terbatas, bukan deployment produksi.
