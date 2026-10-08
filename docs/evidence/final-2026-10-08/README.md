# Hasil verifikasi final

Pengujian lokal 8 Oktober 2026 terhadap source `a465139364001705907e810ca3242f21b924208e`, yang sudah mencakup log latency per percobaan HTTP. Stack dibangun ulang sebelum pengujian. Regresi, k6, outage panjang, dan bootstrap terisolasi dijalankan berurutan.

## Ringkasan

| Pemeriksaan | Hasil | Bukti |
| --- | --- | --- |
| Build stack | Berhasil; service siap | [build.txt](build.txt) |
| Integrasi PostgreSQL | Lulus, termasuk envelope 4 MiB, pagination byte, dan pemulihan backlog 600 record | [postgres.txt](postgres.txt) |
| Regresi stack | Lulus, 306,263 detik | [regression.txt](regression.txt) |
| K6 seismic dan volcanic bersamaan | P95 seismic 12,51 ms, di bawah 300 ms; error non-429 0% | [seismic-only.json](load/seismic-only.json) |
| K6 sustained | 50 VU; sedikitnya 50 TCP selama 87,78 detik; error non-429 0% | [sustained.json](load/sustained.json), [connections.json](load/connections.json) |
| Outage singkat dan recovery | Lulus; seismic 480/480 sehat, volcanic pulih pada fase recovery | [outage.json](load/outage.json) |
| Outage 600 detik dan recovery 90 detik | Lulus tanpa restart atau penggantian container | [result.json](long-outage/result.json), [summary.json](long-outage/summary.json) |
| Bootstrap dari volume kosong | Lulus dengan source bersih, kredensial baru, dan enam volume terpisah | [result.json](bootstrap/result.json) |
| Pemulihan stack utama | Volume, kredensial, dan data acuan dipertahankan; seluruh endpoint health/ready yang diperiksa siap | [final-health.txt](final-health.txt), [final-sources.json](final-sources.json) |

## Lingkungan dan metode beban

Docker Desktop 28.0.4 menyediakan 16 CPU dan 7.936.241.664 byte RAM. K6 0.57.0 berjalan native Windows. Versi tool lengkap terdapat pada [environment.json](environment.json) dan [environment.txt](load/environment.txt).

PVMBG diberi delay tetap 3 detik untuk tiga skenario standar. Skenario paralel memakai 10 VU seismic dan 10 VU volcanic selama 60 detik. Sustained memakai 50 VU selama 90 detik dengan jeda 0,1 detik per iterasi. Outage singkat memakai 5 VU per fase, outage 20 detik, recovery 30 detik, dan satu VU kontrol. Runner mengembalikan konfigurasi mock sesudah skenario standar.

Pada sustained terdapat 42.511 respons bisnis, terdiri dari 8.793 HTTP 200 dan 33.718 HTTP 429. Throughput seluruh respons 471,80/s; throughput sukses 97,59/s. Sebanyak 79,32% respons ditolak secara terkontrol oleh limit satu identitas yang dipakai bersama. Sembilan auth 429 dicatat terpisah; error autentikasi setelah retry 0%. Error bisnis non-429 sebesar 0/8.793, bukan dibagi jumlah respons yang mencakup 429.

P50/p95/p99 respons sukses sustained masing-masing 6,89/13,93/27,11 ms. Sampling TCP menghasilkan 67 sampel, dengan sedikitnya 50 koneksi selama 87,78 detik berturut-turut. Koneksi TCP tidak berarti seluruh request sedang diproses bersamaan. Hasil memenuhi threshold skenario lokal; ini bukan pengukuran kapasitas maksimum atau jaminan performa deployment lain.

Outage singkat menghasilkan 435/480 respons volcanic dengan status basi. Sebanyak 45 respons awal masih berada sebelum kegagalan sumber terdeteksi. Pada recovery, 625/730 respons sudah HEALTHY dan 105 masih dalam pemulihan; seluruhnya HTTP 200. Ringkasan metrik tersedia pada [analysis.json](analysis.json).

## Outage panjang

[Runner outage panjang](../../../scripts/check/long_outage.py) menggunakan script k6 yang sama dengan outage standar, dengan outage 600 detik dan recovery 90 detik. Tiap fase memakai 5 VU; satu pemeriksa tambahan mengambil sampel sekitar setiap 5 detik. Delay mock memakai konfigurasi normal setelah runner standar selesai.

- Semua 14.281 pembacaan seismic selama outage sukses dengan BMKG HEALTHY.
- Sebanyak 14.226/14.281 pembacaan volcanic menampilkan status basi; 55 pembacaan awal mendahului deteksi gangguan.
- Semua 30.753 respons bisnis pada kedua fase HTTP 200; tidak ada 429 atau error bisnis. Error autentikasi 0%.
- Recovery menghasilkan 2.151/2.191 respons HEALTHY. Sampel pertama yang sudah sehat tercatat pada detik ke-605,313 sejak runner mulai; ini observasi sampling, bukan SLA.
- Sebanyak 136 sampel disimpan, termasuk 117 ketika outage aktif. Sampel aktif membentang dari detik 5,266 sampai 600,156; pemulihan dipicu oleh skenario kontrol setelah durasi outage selesai.
- Identitas container, waktu mulai, dan jumlah restart tetap sama sebelum dan sesudah pengujian.

Lihat [konfigurasi](long-outage/configuration.json), [sampel](long-outage/samples.jsonl), [hasil pemeriksaan](long-outage/result.json), dan [keluaran k6](long-outage/k6.txt).

## Bootstrap terisolasi

[Runner bootstrap](../../../scripts/check/clean_bootstrap.py) mengekstrak commit ke direktori baru pada `.local/`, tanpa konfigurasi lokal lama. Generator membuat 12 berkas konfigurasi/kunci. Pemanggilan kedua mempertahankannya secara idempoten. Project `bnpb-bootstrap-90d16f34` memakai network sendiri dan enam volume baru; daftar volume project kosong sebelum startup.

Migrasi mencapai versi 3 dengan status dirty false. API menyediakan 42 hazard dan kedua sumber HEALTHY. Media ditolak dengan 403 ketika meminta raw. Dashboard dan Pemda masing-masing menyimpan 42 snapshot; Notifier mencatat 46 pemrosesan karena satu hazard dapat mempunyai beberapa versi. Ketiga consumer mencakup seluruh ID hazard yang diperiksa dari Canonical Store.

Stack utama dihentikan sementara selama bootstrap. Sesudah uji, hanya volume project uji dihapus dan stack utama dinyalakan kembali. Nama serta waktu pembuatan volume utama, isi berkas kredensial, dan satu hazard acuan diperiksa tetap tersedia. Salinan source dan kredensial uji berada di `.local/` yang diabaikan Git. Bootstrap memakai Docker dan cache host yang tersedia; hasil ini tidak membuktikan instalasi Docker pada host baru atau operasi tanpa jaringan/cache.

Percobaan runner pertama meminta limit consumer 500, sedangkan kontrak endpoint membatasi 200. Parameter runner diperbaiki dan pengujian diulang dari project/volume kosong baru. Artefak percobaan awal dipertahankan pada [bootstrap-initial](bootstrap-initial/result.json) dan [keluaran awal](bootstrap-initial-runner.txt); hasil final menggunakan direktori `bootstrap/`. Tidak ada perubahan kode runtime untuk koreksi parameter runner ini.

## Mengulang pengujian

Gunakan direktori output baru. Jalankan dari root repository secara berurutan, dengan Docker Linux containers aktif dan `K6_BINARY` menunjuk k6 native 0.57.0.

```powershell
docker compose --profile demo up -d --build --wait --wait-timeout 240
docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test
python scripts/demo/demo.py verify all
python scripts/loadtest/run.py --output docs/evidence/run-baru/load
python scripts/check/long_outage.py --output docs/evidence/run-baru/long-outage --seconds 600 --recovery-seconds 90
python scripts/check/clean_bootstrap.py --output docs/evidence/run-baru/bootstrap
```

Transkrip disimpan sebagai UTF-8 dengan akhir baris LF; kode warna terminal dan spasi akhir baris dinormalisasi. Nilai metrik tidak diubah. `manifest.json` mencatat hash artefak dan runner. Bukti ini melengkapi [tes instrumentasi U7](../http-attempts-2026-10-08/README.md). Hasil hosted CI, scan secret seluruh histori revisi pengumpulan, demo sinkron, dan pengumpulan tidak disimpulkan dari tes lokal ini.
