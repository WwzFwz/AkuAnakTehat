# loadtest

[Peta repository](../../README.md)

Konfigurasi k6 untuk pengukuran P2. Hasil terbaru: [pengujian final](../../docs/evidence/final-2026-10-08/README.md).

**Status:** script menggunakan metrik terpisah untuk autentikasi, request bisnis, dan penolakan terkontrol. Runner menyimpan hasil nyata ke `docs/evidence/integration/load/`.

## Berkas dan tanggung jawab

| File | Tanggung jawab |
| --- | --- |
| `seismic-only.js` | Dua skenario seismic/volcanic berjalan bersamaan; p95 request seismic sukses <300 ms. |
| `sustained.js` | Default 50 VU selama 90 s dengan sesi token per VU. |
| `outage.js` | Aktifkan outage PVMBG, verifikasi seismic tetap tersedia dengan BMKG sehat, ukur volcanic stale/unavailable, lalu verifikasi pemulihan. |
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

Override `VUS`, `DURATION`, `SLEEP`, `OUTAGE_SECONDS`, `RECOVERY_SECONDS`,
`OUTAGE_VUS`, `RECOVERY_VUS`, and `K6_HTTP_TIMEOUT` through the environment when
needed. The runner forwards only this allowlist to both native k6 and the Docker
container; arbitrary environment variables, especially credentials, are not
forwarded as scenario options.

The runner deadline is dynamic. It remains at least 240 seconds for the default
scenarios and grows to the selected scenario duration plus a 60-second margin.
The outage deadline includes both `OUTAGE_SECONDS` and `RECOVERY_SECONDS`. Set
`K6_RUNNER_TIMEOUT` with a k6 duration such as `15m` to choose an explicit
deadline for a special run. The explicit deadline must cover the scenario and
the 60-second margin; shorter values are rejected before writing evidence or
changing the mock. For outage, the calculation also includes the two-second
offset before recovery starts. A `duration` argument passed directly to `run()`
takes precedence over `DURATION` for both k6 and the runner deadline.

Example short smoke configuration:

```powershell
$env:VUS = '1'
$env:DURATION = '5s'
$env:SLEEP = '0.2'
$env:OUTAGE_SECONDS = '20'
$env:RECOVERY_SECONDS = '30'
$env:OUTAGE_VUS = '1'
$env:RECOVERY_VUS = '1'
$env:K6_RUNNER_TIMEOUT = '2m'
py scripts/loadtest/run.py --output docs/evidence/loadtest-smoke --skip-connection-check
```

`--skip-connection-check` is only for short smoke runs. Do not use it for final
P2 evidence because the sustained 50-TCP requirement is then not measured.
Remove temporary overrides before a normal run. For a deterministic three-second
PVMBG delay, start the mock with `PVMBG_DELAY_MIN=3s` and
`PVMBG_DELAY_MAX=3s` before running `seismic-only.js`.

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

- Verifikasi forwarding override dan perhitungan deadline tanpa Docker:

  ```powershell
  py -m unittest discover -s scripts/loadtest -p "test_*.py" -v
  ```

- Siapkan label metrik dan definisi denominator error rate.
- Simpan konfigurasi serta output mentah tersanitasi sebagai bukti.
