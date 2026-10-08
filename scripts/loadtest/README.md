# loadtest

[Peta repository](../../README.md)

Konfigurasi k6 untuk pengukuran P2. Hasil terbaru: [integrasi](../../docs/evidence/integration/README.md).

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** script menggunakan metrik terpisah untuk autentikasi, request bisnis, dan penolakan terkontrol. Runner menyimpan hasil nyata ke `docs/evidence/integration/load/`.

## Berkas dan tanggung jawab

| File | Tanggung jawab |
| --- | --- |
| `seismic-only.js` | Dua skenario seismic/volcanic berjalan bersamaan; p95 request seismic sukses <300 ms. |
| `sustained.js` | Default 50 VU selama 90 s dengan sesi token per VU. |
| `outage.js` | Aktifkan outage PVMBG, ukur stale/unavailable, pulihkan, lalu verifikasi HEALTHY. |
| `auth.js` | Helper token per VU, refresh serial, retry satu kali setelah 401, dan metrik error. |
| `run.py` | Runner Compose k6 0.57.0; PVMBG 3s, hasil JSON/text, dan pengukuran koneksi TCP nyata. |

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

Cara yang dapat diulang dari root dengan stack aktif: `py scripts/loadtest/run.py`
(Windows) atau `make load-check` (POSIX, Python 3). Tidak perlu memasang k6 pada
host. Runner memakai image `grafana/k6:0.57.0`, membaca kredensial lokal yang
diabaikan Git, dan menyimpan output/summary. PVMBG dibuat ulang sementara dengan
delay tepat 3s lalu konfigurasi awalnya dipulihkan. Jalankan terpisah dari suite
lain karena skenario outage mengubah keadaan mock.

Runner menolak hasil dengan durasi HTTP negatif. Bila lingkungan container
menghasilkan anomali timer, jangan memakai angka tersebut sebagai bukti latency.
Untuk menjalankan k6 native pada host yang sama, pasang k6 0.57.0 dari rilis resmi
dan set `K6_BINARY` ke path executable (misalnya `$env:K6_BINARY = 'C:\tools\k6.exe'`),
lalu jalankan runner yang sama. Target HTTP memakai port loopback stack Docker;
sampling koneksi dan override delay tetap identik. Versi aktual dicatat pada
`environment.txt`; secret diteruskan melalui environment/file lokal, bukan argumen.

Gunakan direktori keluaran baru untuk menjaga bukti yang telah dirujuk laporan:

```powershell
py scripts/loadtest/run.py --output docs/evidence/demo-local/load
```

Opsi ini berlaku untuk k6 native dan container. Tanpa `--output`, lokasi lama
`docs/evidence/integration/load/` tetap dipakai dan berkas dengan nama sama diperbarui.

Untuk sustained, runner menghitung koneksi TCP ESTABLISHED ke port 8080 melalui
`/proc/net/tcp{,6}` tiap sekitar 1s. Minimal 50 koneksi harus teramati terus-menerus
selama >=60s; jumlah VU saja tidak dianggap bukti jumlah koneksi. Semua koneksi
tidak selalu memiliki request aktif pada saat bersamaan. Ini load model closed
loop, bukan jaminan kapasitas untuk arrival rate arbitrer.

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
are counted as `controlled_429` and excluded from both numerator and denominator
of `system_error_rate`. Authentication has its own `auth_error_rate`.
Authentication throttling is reported separately as `auth_429`.
`business_requests` counts logical business responses, `successful_requests`
counts HTTP 200, and `business_latency` measures only HTTP 200; quick rejections
cannot improve its percentiles. HTTP totals still include auth, retries and control
requests and must not be reported as business throughput. `expired_access_retries`
counts the first 401 handled by refreshing. Tokens
and secrets are not placed in metric tags, checks, or output files. Do not use
`--http-debug=full`.

Metric semantics follow the [k6 metrics documentation](https://grafana.com/docs/k6/latest/using-k6/metrics/).

## Langkah implementasi dan verifikasi

- Siapkan label metrik dan definisi denominator error rate.
- Simpan konfigurasi serta output mentah tersanitasi sebagai bukti.
