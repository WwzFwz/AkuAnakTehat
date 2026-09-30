# loadtest

[Peta repository](../../README.md)

Konfigurasi k6 untuk pengukuran P2; belum ada hasil performa.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `seismic-only.js` | Skenario BMKG bersamaan PVMBG dengan delay3 s, tag endpoint terpisah. |
| `sustained.js` | Minimal50 koneksi paralel selama>= 60 s; rencana90 s. |
| `outage.js` | Request berkelanjutan ketika outage dinyalakan/dimatikan. |
| `auth.js` | Helper token test lokal dengan refresh serial per sesi. |

## Kontrak dan alur

- Laporkan throughput,p50/p95/p99,error rate tanpa penolakan terkontrol, dan jumlah429.
- Threshold BMKG-only p95< 300 ms; error rate di luar penolakan terkontrol< 1%.

## Dependensi

- Client-api dan auth-service yang sudah berjalan; tidak mengakses database langsung.

## Aturan penting

- Dokumentasikan VU, koneksi, pola request, durasi, mesin, dan versi tooling; jangan menganggap VU selalu identik dengan koneksi aktif.
- Jangan berbagi satu refresh token yang dirotasi paralel antarVU.
- Pisahkan login/refresh dan expected401 skenario expiry dari metrik request bisnis.
- Cache off menjadi baseline; cache on diuji hanya jika fitur tambahan sudah ada.

## Langkah implementasi dan verifikasi

- Siapkan label metrik dan definisi denominator error rate.
- Simpan konfigurasi serta output mentah tersanitasi sebagai bukti.
