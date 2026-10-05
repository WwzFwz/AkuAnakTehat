# loadtest

[Peta repository](../../README.md)

Konfigurasi k6 untuk pengukuran P2; belum ada hasil performa.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** script k6 tersedia dan run runtime lokal lulus; output tersanitasi untuk evidence P2 belum disimpan.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `seismic-only.js` | Request seismic dengan tag endpoint dan threshold p95 <300 ms. |
| `sustained.js` | Default 50 VU selama 60 s dengan sesi token per VU. |
| `outage.js` | Aktifkan outage PVMBG, ukur stale/unavailable, pulihkan, lalu verifikasi HEALTHY. |
| `auth.js` | Helper token per VU, refresh serial, retry satu kali setelah 401, dan metrik error. |

## Kontrak dan alur

- Laporkan throughput, p50/p95/p99, error rate tanpa penolakan terkontrol, dan jumlah 429.
- Threshold BMKG-only p95 < 300 ms; error rate di luar penolakan terkontrol < 1%.

## Dependensi

- Client-api dan auth-service yang sudah berjalan; tidak mengakses database langsung.

## Aturan penting

- Dokumentasikan VU, koneksi, pola request, durasi, mesin, dan versi tooling; jangan menganggap VU selalu identik dengan koneksi aktif.
- Jangan berbagi satu refresh token yang dirotasi paralel antarVU.
- Pisahkan login/refresh dan expected401 skenario expiry dari metrik request bisnis.
- Cache off menjadi baseline; cache on diuji hanya jika fitur tambahan sudah ada.

## Menjalankan

Pastikan Compose aktif dan sediakan secret lokal tanpa menuliskannya ke output:

```powershell
$demo = Get-Content -Raw .\env\demo.env | ConvertFrom-StringData
$clients = Get-Content -Raw .\env\demo-clients.json | ConvertFrom-Json
$env:FIELD_CLI_CLIENT_ID = "field-team"
$env:FIELD_CLI_CLIENT_SECRET = ($clients | Where-Object client_id -eq "field-team").client_secret
$env:PVMBG_ADMIN_KEY = $demo.PVMBG_ADMIN_KEY
```

Run from the repository root:

```powershell
k6 run .\scripts\loadtest\seismic-only.js
k6 run .\scripts\loadtest\sustained.js
k6 run .\scripts\loadtest\outage.js
```

Override `VUS`, `DURATION`, `SLEEP`, `OUTAGE_SECONDS`, and `RECOVERY_SECONDS`
when needed. For a deterministic three-second PVMBG delay, start the mock with
`PVMBG_DELAY_MIN=3s` and `PVMBG_DELAY_MAX=3s` before running `seismic-only.js`.

The scripts report p50/p95/p99 through k6 Trend metrics. HTTP `429` responses
are counted as `controlled_429` and excluded from `system_error_rate`. Tokens
and secrets are not placed in metric tags, checks, or output files. Do not use
`--http-debug=full`.

## Langkah implementasi dan verifikasi

- Siapkan label metrik dan definisi denominator error rate.
- Simpan konfigurasi serta output mentah tersanitasi sebagai bukti.
