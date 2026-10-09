# Panduan demo P1 hingga P5

Panduan ini merupakan gladi lokal berdasarkan kriteria M1, bukan mekanisme pengujian resmi asisten. Dokumen spesifikasi menyatakan demo sinkron dilakukan setelah tenggat dan detailnya diumumkan paling lambat H+1. Sesuaikan urutan berikut dengan durasi yang diberikan penguji.

[README utama](../../README.md) | [Memahami sistem dan tanya jawab](../../docs/demo/pemahaman-sistem.md) | [Bukti gladi](../../docs/evidence/demo-2026-10-08/README.md)

## Persiapan sebelum presentasi

1. Aktifkan Docker Desktop dengan Linux containers. Siapkan Go 1.24.2 dan Python 3. Jalankan semua perintah dari direktori utama repository.
2. Build dan unduh dependensi sebelum sesi demo. Build pertama bergantung pada jaringan; jangan mengandalkan waktu demo untuk mengunduh image.
3. Buka diagram arsitektur, terminal perintah, dan laporan. Gunakan terminal kedua untuk `docker compose ps` atau trace.
4. Jalankan satu skenario yang mengubah state pada satu waktu. Pengujian dapat menghentikan sumber, database, atau broker sementara.
5. Pertahankan TTL access token 60 detik, interval generator 10 detik, dan port default. Kredensial dibaca helper dari file lokal dan tidak dicetak.

```powershell
powershell -NoProfile -File scripts/secrets/generate.ps1
docker compose config --quiet
docker compose --profile demo up -d --build --wait --wait-timeout 180
py scripts/demo/demo.py status
py scripts/demo/demo.py read --identity field-team --limit 2
```

Hasil normal berupa HTTP 200 pada health dan ready semua service, array hazard berisi data, dan status sumber HEALTHY. Helper status juga memeriksa Pemda Portal sehingga membutuhkan profile demo. Aggregator tidak dibuka ke host; pengujian kesiapan query dilakukan melalui Client API.

Jika data belum masuk atau status belum HEALTHY, tunggu polling selesai dan ulangi pembacaan. Jangan menafsirkan health 200 sebagai bukti bahwa data selalu baru. PostgreSQL dan volume lama tetap digunakan, sehingga record dari gladi sebelumnya dapat muncul.

## Rundown yang disarankan

| Bagian | Hal yang ditunjukkan | Pemeriksaan otomatis |
| --- | --- | --- |
| Pembuka | Alur ingest, baca, dan event pada diagram | `status` dan `read` |
| P1 | Field tambahan masuk tanpa perubahan kode atau restart | `verify p1` |
| P2 | Outage tidak menghentikan seismic; volcanic diberi status basi dan pulih | `verify p2`, lalu load test terpisah |
| P3 | Kredensial terpisah, raw Media ditolak, refresh setelah expiry alami | `verify p3` |
| P4 | Rebuild satu komponen dan evolusi JSONB tanpa DDL | `verify p4` |
| P5 | Consumer independen, catch-up, dan subscriber baru | `verify p5` |

Perintah `verify` menjalankan pengujian Go yang sudah tersedia, dengan assertion dan exit code bukan sekadar mencetak contoh hasil. Skenario saling beririsan; untuk gladi lengkap cukup jalankan `verify all` sekali, kemudian load test. Jangan menjumlahkan durasi setiap suite sebagai durasi demo resmi. Regresi lokal biasanya membutuhkan beberapa menit, sedangkan pemeriksaan expiry sendiri menunggu 65 detik.

## P1. Perubahan skema tanpa restart

Tujuan demonstrasi ialah memperlihatkan bentuk data lama dan baru tetap dapat dibaca.

```powershell
py scripts/demo/demo.py outage off
py scripts/demo/demo.py schema 1
Start-Sleep -Seconds 20
py scripts/demo/demo.py read --identity field-team --type volcanic --raw --limit 2
```

Catat `hazard_id` dari record baru yang tidak memiliki `confidence_level`. Jangan memakai data mentah Media untuk langkah ini. Setelah itu jalankan perintah berikut.

```powershell
py scripts/demo/demo.py schema 2
Start-Sleep -Seconds 20
py scripts/demo/demo.py read --identity field-team --type volcanic --raw --limit 2
# Ganti UUID dengan hazard_id lama yang tadi dicatat.
py scripts/demo/demo.py read --identity field-team --id <UUID-lama> --raw
py scripts/demo/demo.py verify p1
py scripts/demo/demo.py schema 1
```

Ganti seluruh `<UUID-lama>` sebelum menjalankan perintah. Dalam record baru, `attributes.confidence_level` muncul. Record lama tetap dapat dibaca tanpa field tersebut. Generator menambah record setiap 10 detik dan ingest memerlukan waktu polling; jika belum terlihat, ulangi pembacaan tanpa restart.

Suite otomatis memeriksa ingest, field baru melalui raw API, versi migrasi yang tetap, serta identitas container saat schema toggle. Suite ingest juga melakukan restart sebagai pengujian persistensi terpisah. Jangan menganggap restart tersebut sebagai syarat menerima field baru.

Jelaskan bahwa field wajib tetap divalidasi. Field tambahan yang valid diteruskan ke attributes. Rename, hilangnya field wajib, atau perubahan tipe tidak otomatis didukung. Tunjukkan `services/aggregator/internal/application/canonicalize/input.go` dan `mapper.go` jika diminta.

## P2. Isolasi sumber, degradasi, dan beban

```powershell
py scripts/demo/demo.py outage on --mode error
Start-Sleep -Seconds 20
py scripts/demo/demo.py read --identity field-team --type seismic
py scripts/demo/demo.py read --identity field-team --type volcanic
py scripts/demo/demo.py outage off
Start-Sleep -Seconds 20
py scripts/demo/demo.py read --identity field-team --type volcanic
py scripts/demo/demo.py verify p2
```

Seismic harus tetap dapat dibaca dengan sumber BMKG HEALTHY. Volcanic tetap mengembalikan data tersimpan, tetapi sumber PVMBG berubah menjadi DEGRADED atau DOWN dan memiliki `stale_since`. Sesudah dipulihkan, tunggu polling serta cooldown breaker sampai status kembali HEALTHY. Jangan menuntut status berubah pada milidetik yang sama dengan endpoint admin.

Mode `hang` juga tersedia melalui `outage on --mode hang`. Selalu akhiri dengan `outage off`. Jika penguji meminta outage lebih lama, biarkan outage aktif selama durasi yang diminta dan ulangi pembacaan kedua jenis data; jangan menyebut pengujian 20 detik sebagai bukti outage 20 menit.

Untuk beban, jalankan runner berikut saat tidak ada skenario lain berjalan.

```powershell
# Jika memakai k6 native, arahkan ke k6 0.57.0 yang terpasang.
$env:K6_BINARY = 'C:\lokasi-k6\k6.exe'
py scripts/loadtest/run.py --output docs/evidence/demo-local/load
```

Path contoh harus diganti. Jika `K6_BINARY` tidak diset, runner memakai image k6 pada Compose. Pada lingkungan lokal ini pernah ditemukan durasi negatif dari k6 container; hasil demikian ditolak runner dan tidak sah untuk bukti performa. Versi native yang dipakai saat gladi tersimpan pada `.local/k6-native/k6-v0.57.0-windows-amd64/k6.exe` bila arsip lokal tersedia. Panduan umum tetap ada di [load test](../loadtest/README.md).

Tunjukkan p95 **seismic saja** di bawah 300 ms pada delay PVMBG tepat 3 detik, sampling sedikitnya 50 koneksi TCP selama minimal 60 detik, percentile, throughput sukses, serta error di luar 429 kurang dari 1%. Laporkan jumlah 429 dan bedakan dari respons sukses. Runner memulihkan konfigurasi delay PVMBG setelah selesai. Direktori output baru menjaga bukti 6 Oktober yang digunakan laporan.

## P3. Hak akses dan sesi autentikasi

```powershell
py scripts/demo/demo.py read --identity media
py scripts/demo/demo.py read --identity media --raw --expect 403
py scripts/demo/demo.py read --identity field-team --raw
py scripts/demo/demo.py read --identity bnpb-ops --raw
py scripts/demo/demo.py verify p3
```

Media hanya mendapat tujuh field ringkasan. Permintaan raw Media harus HTTP 403; karena itu helper memakai `--expect 403`. Tim Lapangan dan BNPB Pusat memperoleh koordinat, source reference ID, dan attributes.

Suite memeriksa kredensial lintas instansi, JWT invalid, pembatasan field, rotasi refresh, serta satu proses CLI yang tetap hidup selama 65 detik. Token awal memiliki TTL 60 detik. Token lama ditolak sebelum maupun setelah refresh, sedangkan token baru bekerja. Jangan memendekkan TTL atau mengubah jam mesin untuk demo expiry.

Helper `read` membuka sesi baru pada tiap pemanggilan sehingga beberapa pemanggilan `read` terpisah **bukan** bukti refresh otomatis. Gunakan suite expiry atau satu proses CLI dengan `--watch 65s --count 2`. Kredensial CLI dimuat dari file lokal mengikuti [panduan CLI](../../tools/field-cli/README.md), bukan ditulis ke slide atau argumen shell yang dipresentasikan.

Jelaskan bahwa JWT diverifikasi offline. Reuse refresh mencabut keluarga refresh, tetapi access token yang belum expired tetap dapat hidup sampai TTL-nya habis. Redis Lua tidak mencakup pengiriman respons HTTP.

## P4. Rebuild independen dan kepemilikan storage

```powershell
py scripts/demo/demo.py verify p4
```

Perhatikan keluaran `TestIndependentRebuild`. Notifier dihentikan, image-nya dibangun ulang, dan service dijalankan dengan `--no-deps`. Pembacaan Media tetap berhasil. Timestamp start Aggregator, Client API, Auth Service, mock, Dashboard Updater, serta Kafka tidak berubah.

`TestDynamicSchemaAndStaleAPI` membuktikan record lama dan baru berdampingan tanpa versi migrasi berubah. Tunjukkan JSONB pada migrasi Aggregator dan adapter HTTP Client API jika ditanya. Client API dan consumer tidak memiliki jalur aplikasi langsung ke Canonical Store.

Akses database oleh test atau operator Docker merupakan pemeriksaan administratif. Jangan mengklaim network Compose melindungi dari administrator host. Jangan menghapus volume untuk membuat pengujian tampak bersih.

## P5. Consumer independen dan penambahan subscriber

```powershell
py scripts/demo/demo.py view dashboard-updater
py scripts/demo/demo.py view notifier
py scripts/demo/demo.py view pemda-portal
py scripts/demo/demo.py verify p5
```

Suite menampilkan tahapan aliran mock sampai Kafka dan dashboard, replay dengan group Pemda baru serta SQLite kosong, deduplikasi, DLQ, catch-up dashboard, dan pemulihan broker. Data fixture diberi ID unik. Producer tidak diganti ketika subscriber baru ditambahkan.

Untuk memperlihatkan penghentian consumer secara langsung, jalankan langkah berikut secara terpisah dari suite otomatis.

```powershell
docker compose stop dashboard-updater
Start-Sleep -Seconds 20
py scripts/demo/demo.py view notifier
py scripts/demo/demo.py view pemda-portal
docker compose up -d --no-deps --wait --wait-timeout 90 dashboard-updater
py scripts/demo/demo.py verify p5
```

View default diurutkan menurut ID, bukan waktu kedatangan. Dua record yang tampil sama tidak membuktikan consumer berhenti. Suite memeriksa ID dan version tertentu sehingga menjadi bukti catch-up yang lebih tepat. Untuk inspeksi manual halaman lain, gunakan `view <service> --after <next_cursor>`.

Dashboard dan Pemda menyimpan versi terbaru. Notifier memakai marker hazard/version dan mengirim alert sebagai log simulasi. Crash sesudah send tetapi sebelum marker masih dapat menghasilkan duplikasi. Jangan menyebutnya exactly-once. Group lama mengikuti committed offset, sedangkan group baru membaca dari earliest yang masih tersedia pada retensi.

## Trace dan bukti

Helper `read` menampilkan `correlation_id` pada output tanpa token. Salin nilainya ke helper trace berikut.

```powershell
py scripts/trace/trace.py <correlation-id>
```

Query memiliki trace sendiri; query bukan penyebab ingest. Untuk alur ingest sampai consumer, cari correlation ID dari log `outbox_published`, kemudian telusuri melalui helper yang sama. Jangan menggunakan `docker inspect` penuh, `docker compose config` penuh, atau debug HTTP yang dapat mencetak secret sebagai screenshot.

Untuk menyimpan satu suite pada PowerShell dan tetap melihat hasilnya, gunakan perintah berikut. Periksa `$LASTEXITCODE` sebelum menyebut suite lulus.

```powershell
New-Item -ItemType Directory -Force docs/evidence/demo-local | Out-Null
py scripts/demo/demo.py verify all 2>&1 | Tee-Object docs/evidence/demo-local/regression.txt
if ($LASTEXITCODE -ne 0) { throw 'Regresi gagal; periksa output.' }
```

## Pemulihan jika demo terputus

```powershell
py scripts/demo/demo.py restore
py scripts/demo/demo.py status
Start-Sleep -Seconds 20
py scripts/demo/demo.py read --identity field-team --type volcanic
```

`restore` mengembalikan konfigurasi Compose normal, menjalankan ulang dependensi dan consumer, menonaktifkan outage, serta memilih schema 1 untuk record berikutnya. Perintah ini **bukan** pemulihan konfigurasi kustom sebelum demo. Record lama, state SQLite, outbox, group uji, dan volume tetap ada. Jika startup belum pernah dilakukan, jalankan bootstrap dan `up --build` terlebih dahulu.

Jika Docker tidak dapat diakses, aktifkan Docker Desktop dan periksa `docker info`. Jika pembacaan HTTP 401, pastikan secret lokal dan container berasal dari bootstrap yang sama. Jika HTTP 503, periksa readiness dependensi dan status sumber. Jika port sudah terpakai, hentikan proses yang memang memakai port demo; jangan mengganti seluruh port tanpa memperbarui kontrak pengujian.

## Sebelum dikumpulkan

- Baca [panduan pemahaman](../../docs/demo/pemahaman-sistem.md) dan lakukan latihan menjelaskan diagram tanpa membaca script.
- Konfirmasi hasil workflow final pada tab Actions GitHub. Run foundation #21 berhasil; hasil lokal tetap tidak menggantikan verifikasi hosted CI.
- Pastikan tidak ada placeholder laporan yang tertinggal sebelum pengumpulan.
- Setelah kelompok meninjau versi final, buat tag `milestone-1` dan isi formulir sesuai spesifikasi. Panduan ini tidak membuat tag atau mengirim tugas secara otomatis.
