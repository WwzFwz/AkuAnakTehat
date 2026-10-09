# AkuAnakTehat

Platform koordinasi kebencanaan BNPB untuk **IF4031 Arsitektur Aplikasi Terdistribusi, Milestone 1**. Sistem menggabungkan informasi gempa dan potensi tsunami dari BMKG dengan laporan aktivitas vulkanik dari PVMBG, menyajikannya sesuai hak akses pengguna, serta menyalurkan perubahannya kepada consumer independen.

Kedua instansi berupa mock dengan data sintetis. Implementasi mencakup polling, pemetaan data kanonik, autentikasi, API baca, distribusi event, dan pengujian gangguan serta pemulihan. Ini merupakan PoC lokal pada satu host dengan delapan service aplikasi ketika Pemda Portal diaktifkan.

## Arsitektur

![Arsitektur sistem BNPB](docs/laporan/assets/figures/01-arsitektur.svg)

Setiap service berjalan dalam container terpisah dan mengakses penyimpanan miliknya sendiri. SQLite dan memori mock berada di dalam container service pemilik. Consumer membaca event dari Kafka secara independen.

Sistem memisahkan tiga alur agar kegagalan sumber tidak langsung menghambat pembacaan pengguna.

1. **Pengambilan data.** Worker Aggregator mengambil JSON dari BMKG dan PVMBG secara independen, memvalidasi field wajib, mempertahankan atribut tambahan, dan memetakan data ke `HazardEvent`. Warning tsunami dikorelasikan dengan gempa terkait. Hazard dan outbox disimpan atomik per record. Checkpoint maju setelah seluruh respons polling ditangani; retry memakai hash untuk menghindari duplikasi.
2. **Pembacaan data.** Client memperoleh JWT dari Auth Service, kemudian mengakses Client API. Setelah memeriksa token, hak akses, dan batas beban, Client API memanggil API internal Aggregator. Data dibaca dari Canonical Store sehingga request pengguna tidak menunggu HTTP sumber. Media memperoleh ringkasan; Tim Lapangan dan BNPB Pusat boleh membaca data mentah. Status sumber membantu pengguna menilai kebaruan data.
3. **Distribusi perubahan.** Relay mengirim snapshot outbox ke Kafka dan menandai publish setelah ACK. Dashboard Updater, Notifier, dan Pemda Portal memakai consumer group serta SQLite masing-masing. Consumer yang berhenti dapat melanjutkan pembacaan selama pesannya masih tersedia dalam retensi broker.

### Tanggung jawab service

| Service | Tanggung jawab | Penyimpanan | Akses default |
| --- | --- | --- | --- |
| [BMKG Mock](services/bmkg-mock/README.md) | Gempa, warning tsunami, seed, generator, dan API key | Memori lokal | `127.0.0.1:8081` |
| [PVMBG Mock](services/pvmbg-mock/README.md) | Vulkanik, perubahan skema, delay, dan outage | Memori lokal | `127.0.0.1:8082` |
| [Aggregator](services/aggregator/README.md) | Polling, pemetaan, korelasi, transaksi, query, dan relay | Pemilik PostgreSQL | `aggregator:9000`, internal |
| [Client API](services/client-api/README.md) | JWT, scope, proyeksi, filter, pagination, dan proteksi beban | Tidak persisten | `127.0.0.1:8080` |
| [Auth Service](services/auth-service/README.md) | Access token dan rotasi refresh token | Pemilik Redis | `127.0.0.1:8090` |
| [Dashboard Updater](services/dashboard-updater/README.md) | View hazard versi terbaru | SQLite sendiri | `127.0.0.1:8091/view` |
| [Notifier](services/notifier/README.md) | Simulasi alert SIAGA dan AWAS serta deduplikasi | SQLite sendiri | `127.0.0.1:8092/processed` |
| [Pemda Portal](services/pemda-portal/README.md) | Subscriber tambahan tanpa perubahan producer | SQLite sendiri | `127.0.0.1:8093/view`, profile `demo` |

Setiap service memiliki module Go dan Dockerfile sendiri. Client API tidak mengakses PostgreSQL langsung. Consumer membaca Kafka, sedangkan Auth Service mengelola sesi di Redis. Poller, mapper, query, dan relay tetap menjadi komponen Aggregator karena menggunakan data dan transaksi milik service tersebut.

PostgreSQL, Redis, dan Kafka menggunakan jaringan internal tanpa memublikasikan port ke host. Port HTTP aplikasi hanya terikat ke loopback. Detail endpoint, model data, dan format event tersedia pada [kontrak API](docs/api/README.md).

## Teknologi dan alasan pemilihan

| Teknologi | Penggunaan dan alasan |
| --- | --- |
| Go 1.24.2 | Goroutine untuk worker independen dan `context` untuk deadline; module terpisah menjaga build setiap service. |
| HTTP, JSON, dan `net/http` | Kontrak mudah diperiksa; tolerant reader menerima atribut baru sambil memvalidasi field wajib. |
| PostgreSQL 16.8, JSONB, dan pgx | Hazard dan outbox disimpan atomik per record; checkpoint maju setelah seluruh respons ditangani. Kolom kanonik mendukung filter, sedangkan JSONB menyimpan atribut tambahan tanpa perubahan skema tabel. |
| golang-migrate | Migrasi tabel berversi yang di-embed dan dijalankan Aggregator. |
| Redis 7.4.2, AOF, dan Lua | TTL sesi, persistensi refresh token, dan rotasi state secara atomik di Redis. |
| JWT EdDSA dengan Ed25519 | Auth Service memegang private key; Client API memverifikasi dengan public key tanpa meminta auth pada setiap request. |
| Apache Kafka 3.9.1 dan franz-go | Retensi event dan offset per group mendukung subscriber independen serta replay. |
| SQLite melalui modernc | View dan deduplikasi lokal tanpa menambah server database untuk setiap consumer. |
| Docker Compose | Mengatur container, network, volume, serta health check bersama. |
| k6 0.57.0 | Mengukur latency, throughput, error, serta penolakan beban. Koneksi TCP diukur terpisah dari VU. |

Kafka memakai satu broker dengan replication factor 1. Pengiriman event bersifat at-least-once; notifier masih dapat mengirim ulang jika crash terjadi setelah pengiriman tetapi sebelum marker tersimpan. Notifikasi hanya disimulasikan melalui log. Atomisitas Redis tidak menjamin respons HTTP token baru sampai ke client. Sistem belum menggunakan TLS, koordinasi banyak instance, atau failover lintas host.

## Struktur repository

```text
AkuAnakTehat/
|-- services/
|   |-- aggregator/          # Ingest, query, PostgreSQL, outbox, Kafka
|   |-- auth-service/        # JWT, refresh token, Redis
|   |-- client-api/          # API publik, scope, proyeksi, proteksi beban
|   |-- bmkg-mock/           # Gempa dan warning tsunami sintetis
|   |-- pvmbg-mock/          # Vulkanik, schema toggle, delay, outage
|   |-- dashboard-updater/   # View event di SQLite
|   |-- notifier/            # Alert simulasi dan deduplikasi
|   `-- pemda-portal/        # Subscriber tambahan
|-- tools/field-cli/         # Client HTTP Tim Lapangan
|-- scripts/
|   |-- secrets/            # Bootstrap kredensial lokal
|   |-- check/              # Uji unit, integrasi, dan regresi
|   |-- demo/               # Panduan dan helper demo P1 hingga P5
|   |-- loadtest/           # Skenario k6 dan pengukuran TCP
|   `-- trace/              # Penelusuran correlation ID
|-- docs/
|   |-- api/                # Kontrak HTTP, storage, port, event
|   |-- demo/               # Pemahaman sistem dan tanya jawab
|   |-- evidence/           # Hasil pengujian
|   `-- laporan/            # Laporan dan diagram sistem
|-- infra/                  # Konfigurasi dan panduan infrastruktur
|-- env/                    # Konfigurasi lokal; secret diabaikan Git
|-- .github/workflows/      # Uji CI dan secret scanning
|-- docker-compose.yml      # Orkestrasi
|-- go.work                 # Workspace sembilan module Go
`-- Makefile                # Shortcut untuk shell POSIX
```

Pada setiap service, `cmd/` merakit proses dan dependensi, sedangkan `internal/` menampung domain, aplikasi, adapter, konfigurasi, dan observability sesuai kebutuhan service. Aggregator juga memiliki `migrations/` dan referensi gunung sintetis.

## Menjalankan sistem

Siapkan Git, Go 1.24.2, Docker Desktop atau Docker Engine dengan Compose v2, dan Python 3. Aktifkan Docker dengan Linux containers. Build pertama memerlukan akses internet untuk mengunduh image dan dependensi. Pastikan port host 8080–8082 dan 8090–8093 tersedia.

Jika repository belum tersedia, clone terlebih dahulu. Jika sudah, cukup buka terminal di direktori utamanya.

```powershell
git clone https://github.com/WwzFwz/AkuAnakTehat.git
cd AkuAnakTehat
```

Pada Windows PowerShell, jalankan berurutan. Lanjutkan hanya jika perintah sebelumnya berhasil.

```powershell
powershell -NoProfile -File scripts/secrets/generate.ps1
docker compose config --quiet
docker compose --profile demo up -d --build --wait --wait-timeout 240
py scripts/demo/demo.py status
py scripts/demo/demo.py read --identity media
```

Bootstrap membuat kredensial lokal di `env/` dan mempertahankan secret yang sudah ada. Tidak perlu menyalin `.env.example` menjadi `.env` atau mengisi password secara manual. Migrasi database dan pembuatan topic dijalankan otomatis saat startup. Profile `demo` menjalankan seluruh service, termasuk Pemda Portal yang diperiksa oleh perintah `status`. Jika Python tidak menyediakan launcher `py`, gunakan `python` sebagai penggantinya.

Pada Linux atau macOS, jalankan dari direktori utama repository:

```sh
export LOCAL_UID=$(id -u) LOCAL_GID=$(id -g)
sh scripts/secrets/generate.sh
docker compose config --quiet
docker compose --profile demo up -d --build --wait --wait-timeout 240
python3 scripts/demo/demo.py status
python3 scripts/demo/demo.py read --identity media
```

UID/GID memungkinkan service membaca bind mount secret. Tanda startup berhasil adalah container aktif berstatus `healthy`, hasil `status` menunjukkan HTTP 200, dan pembacaan API mengembalikan array `data`. Data awal memerlukan beberapa siklus polling; ulangi pembacaan sampai `sources` BMKG dan PVMBG berstatus `HEALTHY`. Generator selanjutnya menambahkan data sintetis setiap 10 detik.

Untuk memeriksa akses Tim Lapangan dan pembatasan akses Media:

```powershell
py scripts/demo/demo.py read --identity field-team --type volcanic --raw
py scripts/demo/demo.py read --identity media --raw --expect 403
```

Jika startup gagal, periksa `docker compose --profile demo ps -a` dan `docker compose logs --tail=100 <nama-service>`. Jika PowerShell menolak eksekusi berkas skrip, jalankan skrip yang sama dengan `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/secrets/generate.ps1`; pengaturan hanya berlaku untuk proses tersebut. Konfigurasi bootstrap yang tidak lengkap harus diperbaiki sebelum startup, bukan ditimpa sebagian dengan secret baru. [Panduan infrastruktur](infra/README.md) menjelaskan konfigurasi jaringan dan volume.

Untuk menghentikan seluruh stack dengan data tetap tersimpan, gunakan `docker compose --profile demo down`. Menjalankan kembali perintah `up` menggunakan volume yang sama; jangan menambahkan `--volumes` jika ingin mempertahankan data.

### Variabel konfigurasi

Compose membaca `env/<service>.env` yang dihasilkan bootstrap. Nilai opsional yang belum tertulis dapat ditambahkan ke berkas service terkait; tabel berikut mencantumkan nilai bawaan atau nilai hasil bootstrap. Durasi memakai satuan seperti `ms`, `s`, dan `h`. Secret pada tabel berarti nilai acak lokal, bukan nilai yang perlu disalin dari README. [.env.example](.env.example) hanya referensi sebagian variabel, bukan konfigurasi siap jalan.

| Lokasi | Variabel dan nilai | Kegunaan |
| --- | --- | --- |
| Semua service HTTP | `HTTP_ADDR` sesuai port pada tabel service | Alamat listen di dalam container. Mengubah port host dilakukan pada `ports` di Compose. |
| `env/postgres.env` | `POSTGRES_DB=bnpb`, `POSTGRES_USER=bnpb`, `POSTGRES_PASSWORD` berupa secret | Database kanonik milik Aggregator. |
| `env/redis.env` | `REDIS_PASSWORD` berupa secret | Kredensial Redis, diselaraskan bootstrap dengan Auth Service. |
| `env/bmkg-mock.env` | `BMKG_KEY_HASH` berupa hash kredensial, `GENERATION_INTERVAL=10s`, `BMKG_FIXED_DELAY=100ms` | Autentikasi, interval generator, dan delay BMKG. |
| `env/pvmbg-mock.env` | `PVMBG_TOKEN_HASH`, `ADMIN_KEY_HASH` berupa hash kredensial | Autentikasi pembacaan dan kontrol demo PVMBG. |
| `env/pvmbg-mock.env` | `GENERATION_INTERVAL=10s`, `PVMBG_DELAY_MIN=500ms`, `PVMBG_DELAY_MAX=3s` | Interval generator dan rentang delay sumber. |
| Auth Service dan Client API | `JWT_ISSUER=bnpb-auth`, `JWT_AUDIENCE=bnpb-api` | Claim yang harus cocok antara penerbit dan pemeriksa token. |
| `env/auth-service.env` | `JWT_PRIVATE_KEY_FILE=/run/keys/jwt-private.pem`, `CLIENTS_FILE=/run/config/clients.json` | Lokasi kunci privat dan daftar identitas di container. |
| `env/auth-service.env` | `REDIS_ADDR=auth-store:6379`, `REDIS_PASSWORD` berupa secret, `REDIS_TIMEOUT=200ms` | Koneksi penyimpanan sesi. |
| `env/auth-service.env` | `ACCESS_TOKEN_TTL=60s`, `REFRESH_TOKEN_TTL=8h`, `TOKEN_RATE_LIMIT=20`, `TOKEN_RATE_BURST=40` | Masa berlaku token dan pembatasan request token. |
| `env/client-api.env` | `JWT_PUBLIC_KEY_FILE=/run/keys/jwt-public.pem`, `AGGREGATOR_URL=http://aggregator:9000`, `INTERNAL_KEY` berupa secret | Verifikasi JWT dan akses API internal. |
| `env/client-api.env` | `AGGREGATOR_TIMEOUT=1500ms` | Batas total panggilan Aggregator; harus positif dan tidak melebihi 1500 ms. |
| `env/client-api.env` | `MAX_CONCURRENT=100`, `RATE_LIMIT_RPS=100`, `RATE_LIMIT_BURST=200` | Batas request aktif serta rate limit per identitas. |
| `env/client-api.env` | `PAGE_DEFAULT=100`, `PAGE_MAX=500` | Ukuran halaman; default tidak boleh melebihi maksimum, maksimum tidak boleh melebihi 500. |
| `env/aggregator.env` | `DATABASE_URL` berisi koneksi PostgreSQL dan secret, `INTERNAL_KEY` berupa secret | Akses storage dan autentikasi API internal; key harus cocok dengan Client API. |
| `env/aggregator.env` | `BMKG_URL=http://bmkg-mock:8081`, `PVMBG_URL=http://pvmbg-mock:8082`, `BMKG_API_KEY`, `PVMBG_TOKEN` | Alamat sumber serta kredensial hasil bootstrap. |
| `env/aggregator.env` | `BMKG_POLL_INTERVAL=2s`, `PVMBG_POLL_INTERVAL=5s`, `POLL_OVERLAP=10s` | Interval pengambilan data dan overlap watermark. |
| `env/aggregator.env` | `BMKG_TIMEOUT=1s`, `PVMBG_TIMEOUT=4s`, `DB_TIMEOUT=2s` | Batas waktu panggilan sumber serta operasi database. |
| `env/aggregator.env` | `DB_POOL_SIZE=5`, `QUERY_DB_POOL_SIZE=3`, `BREAKER_FAILURES=3`, `BREAKER_COOLDOWN=10s` | Pool ingest/query serta circuit breaker sumber. |
| `env/aggregator.env` | `KAFKA_BROKERS=kafka:9092`, `KAFKA_TOPIC=bnpb.hazard-events.v1`, `KAFKA_PUBLISH_TIMEOUT=5s` | Tujuan dan timeout publish event. |
| `env/aggregator.env` | `OUTBOX_POLL_INTERVAL=1s`, `OUTBOX_RETENTION=24h` | Interval relay dan retensi record published; pending tidak dihapus oleh retensi ini. |
| `environment` pada masing-masing consumer di Compose | `KAFKA_BROKERS=kafka:9092`, `KAFKA_TOPIC=bnpb.hazard-events.v1`, `KAFKA_DLQ_TOPIC=bnpb.hazard-events.v1.dlq` | Broker, topic sumber, dan DLQ. |
| `environment` pada masing-masing consumer di Compose | `KAFKA_GROUP_ID` mengikuti nama service, `PROCESS_TIMEOUT=2s`, `MAX_ATTEMPTS=3` | Subscription independen, timeout pemrosesan, dan jumlah percobaan. Jangan menyamakan group ketiga consumer. |
| `environment` pada masing-masing consumer di Compose | `SQLITE_PATH=/data/view.db`; Notifier memakai `/data/processed.db` | Penyimpanan lokal pada volume masing-masing consumer. |
| Shell Linux/macOS | `LOCAL_UID`, `LOCAL_GID` mengikuti pengguna host | Izin baca berkas secret; Windows memakai default 10001. |

Consumer memakai default di kode dan tidak membaca `env/<consumer>.env` secara otomatis. Konfigurasi broker seperti listener, retensi, dan ukuran pesan terdapat pada service `kafka` di [docker-compose.yml](docker-compose.yml); pembuatan topic terdapat pada [init-topics.sh](infra/kafka/init-topics.sh).

Helper demo membaca `BMKG_API_KEY`, `PVMBG_TOKEN`, `PVMBG_ADMIN_KEY`, dan `INTERNAL_KEY` dari `env/demo.env`, serta kredensial client dari `env/demo-clients.json`. Auth Service memakai versi hash pada `env/clients.json`. Kunci JWT berada pada `env/keys/`. Seluruh berkas tersebut dibuat bootstrap dan diabaikan Git.

Untuk mengubah konfigurasi, edit berkas milik service terkait, kemudian buat ulang service tersebut. Contoh setelah mengubah batas beban pada `env/client-api.env`:

```powershell
docker compose config --quiet
docker compose up -d --no-deps --force-recreate --wait --wait-timeout 180 client-api
```

`docker compose restart` saja tidak memuat ulang environment. Perubahan secret yang dipakai bersama harus diselaraskan pada pihak pemakai dan penyimpannya; menjalankan generator kembali mempertahankan konfigurasi yang sudah ada.

### Konfigurasi load test

Variabel berikut diatur pada shell yang menjalankan `scripts/loadtest/run.py`. Kredensial dibaca dari berkas lokal hasil bootstrap.

| Variabel | Nilai bawaan dan fungsi |
| --- | --- |
| `K6_BINARY` | Path executable k6 native. Jika tidak diatur, runner memakai container `grafana/k6:0.57.0`. |
| `VUS` | Default 10 pada masing-masing skenario seismic/volcanic paralel dan 50 pada sustained. Override berlaku untuk kedua script. |
| `DURATION` | Default `60s` untuk skenario paralel dan `90s` untuk sustained. |
| `SLEEP` | Jeda iterasi; default 0,1 s pada paralel/sustained dan 0,2 s pada outage. |
| `OUTAGE_SECONDS`, `RECOVERY_SECONDS` | Default 20 dan 30 detik. |
| `OUTAGE_VUS`, `RECOVERY_VUS` | Default masing-masing 5; ada satu VU tambahan untuk kontrol pemulihan. |
| `K6_HTTP_TIMEOUT` | Default `5s` untuk request HTTP k6. |
| `K6_RUNNER_TIMEOUT` | Opsional, misalnya `15m`. Jika diisi harus mencakup durasi skenario dan margin 60 s; outage juga menghitung offset recovery 2 s. Tanpa override, deadline minimal 240 s. |

## Demo dan pengujian

Jalankan pemeriksaan berikut dari root repository. Uji unit tidak membutuhkan stack, sedangkan regresi dan k6 memerlukan seluruh service pada profile `demo` sudah aktif. Pada Linux/macOS gunakan `sh scripts/check/check.sh` untuk pemeriksaan Go dan ganti `py` dengan `python3`.

```powershell
powershell -NoProfile -File scripts/check/check.ps1
py scripts/demo/demo.py verify all
```

Pemeriksaan Go berhasil jika seluruh tes dan `go vet` selesai dengan exit code 0. `verify all` menjalankan regresi tanpa cache hasil Go dan harus berakhir dengan `PASS`. Untuk menjalankan satu kelompok skenario, pilih perintah berikut; tidak perlu menjalankan semuanya lagi setelah `verify all` berhasil.

| Skenario | Perintah | Perilaku yang diperiksa dan tanda berhasil |
| --- | --- | --- |
| P1, interoperabilitas | `py scripts/demo/demo.py verify p1` | Ingest, korelasi, dan perubahan skema; record lama dan baru tetap terbaca tanpa perubahan DDL, tes berakhir `PASS`. |
| P2, degradasi | `py scripts/demo/demo.py verify p2` | Saat PVMBG terganggu, seismic tetap sehat, volcanic menampilkan status basi, lalu pulih; tes berakhir `PASS`. Ukuran beban dan latency diperiksa melalui k6 di bawah. |
| P3, autentikasi | `py scripts/demo/demo.py verify p3` | Kredensial silang dan raw Media ditolak, expiry alami dan refresh CLI berhasil; tes berakhir `PASS`. Sesi CLI menunggu 65 detik untuk melewati TTL default 60 detik. |
| P4, isolasi komponen | `py scripts/demo/demo.py verify p4` | Rebuild Notifier tidak mengganti service lain, pembacaan tetap berjalan, dan atribut baru tersimpan tanpa DDL; tes berakhir `PASS`. |
| P5, pub/sub | `py scripts/demo/demo.py verify p5` | Consumer independen, replay subscriber baru, dedup, DLQ, pemulihan broker, serta event besar; tes berakhir `PASS`. |

Untuk mengamati perubahan secara manual, jalankan satu perintah setiap kali dan beri waktu beberapa siklus polling sebelum membaca data:

```powershell
py scripts/demo/demo.py schema 2
py scripts/demo/demo.py read --identity field-team --type volcanic --raw
py scripts/demo/demo.py outage on --mode error
py scripts/demo/demo.py read --identity media --type seismic
py scripts/demo/demo.py read --identity field-team --type volcanic
py scripts/demo/demo.py outage off
py scripts/demo/demo.py schema 1
```

Skema 2 menambahkan field pada record baru; record lama tidak otomatis berubah. Status sumber tidak berubah seketika setelah outage diaktifkan atau dimatikan. [Panduan demo](scripts/demo/README.md) menjelaskan langkah inspeksi setiap skenario. Pengujian tambahan outage 600 detik dan bootstrap terisolasi tersedia pada [panduan pemeriksaan](scripts/check/README.md).

### Load test P2

Bukti performa laporan memakai **k6 0.57.0 native Windows** dengan seluruh service aplikasi tetap di Docker. Untuk memakai alat yang sama, unduh arsip yang sesuai sistem operasi dari [rilis resmi k6 v0.57.0](https://github.com/grafana/k6/releases/tag/v0.57.0), ekstrak, lalu atur `K6_BINARY` ke path absolut executable. Jika k6 sudah berada di `PATH`, gunakan perintah berikut pada PowerShell:

```powershell
$env:K6_BINARY = (Get-Command k6 -ErrorAction Stop).Source
& $env:K6_BINARY version
$loadOutput = 'docs/evidence/demo-local/load-' + (Get-Date -Format 'yyyyMMdd-HHmmss')
py scripts/loadtest/run.py --output $loadOutput
```

Pastikan keluaran versi menunjukkan `v0.57.0`. Jika belum ada di `PATH`, ganti baris pertama dengan `$env:K6_BINARY = 'C:\lokasi-ekstraksi\k6.exe'` menggunakan lokasi executable yang nyata. Pada Linux/macOS gunakan `export K6_BINARY="$(command -v k6)"`, periksa `"$K6_BINARY" version`, lalu jalankan `python3 scripts/loadtest/run.py --output "docs/evidence/demo-local/load-$(date +%Y%m%d-%H%M%S)"`.

Alternatif tanpa instalasi k6 pada host adalah mode Docker. Pada PowerShell hapus pilihan native, kemudian jalankan runner:

```powershell
Remove-Item Env:K6_BINARY -ErrorAction SilentlyContinue
$loadOutput = 'docs/evidence/demo-local/load-docker-' + (Get-Date -Format 'yyyyMMdd-HHmmss')
py scripts/loadtest/run.py --output $loadOutput
```

Pada Linux/macOS gunakan `unset K6_BINARY`. Lingkungan Docker Desktop lokal pernah menghasilkan timing negatif pada k6 container. Runner menolak hasil seperti itu; bila terjadi, ulangi dengan k6 native dan direktori keluaran baru. Hasil negatif tidak dipakai sebagai bukti performa. Pilihan native atau Docker mengubah lokasi pembangkit beban, bukan lokasi service aplikasi.

Runner sementara mengatur delay PVMBG tepat 3 detik dan menjalankan tiga skenario berurutan. Keberhasilan ditandai `PASS seismic-only`, `PASS sustained`, `PASS outage`, dan exit code 0. Pemeriksaan mencakup p95 seismic di bawah 300 ms, sedikitnya 50 koneksi TCP selama minimal 60 detik, error bisnis di luar 429 kurang dari 1%, serta kondisi sumber saat outage dan recovery. JSON, transkrip, dan sampel koneksi tersimpan di direktori `--output`. Respons 429 tetap dilaporkan terpisah dari respons sukses.

Untuk memakai konfigurasi beban standar, gunakan shell yang tidak berisi override skenario dari percobaan sebelumnya. Jangan memakai `--skip-connection-check` untuk bukti persyaratan 50 TCP. [Panduan load test](scripts/loadtest/README.md) menjelaskan opsi smoke test dan definisi metrik lebih lanjut.

### Pemulihan setelah pengujian

Regresi, load test, dan demo yang mengubah state harus berjalan **berurutan** karena dapat menghentikan dependensi sementara. Hentikan urutan jika suatu perintah gagal dan periksa lognya. Setelah selesai, atau jika demo terputus, pulihkan kondisi normal:

```powershell
py scripts/demo/demo.py restore
py scripts/demo/demo.py status
```

`restore` menjalankan service dengan Compose normal, menonaktifkan outage, dan mengembalikan generator PVMBG ke skema 1. Data sintetis serta volume tetap dipertahankan. Status sumber memerlukan beberapa siklus polling untuk kembali sehat.

Hasil pengujian akhir pada lingkungan lokal menunjukkan p95 seismic **12,51 ms**, sedikitnya **50 koneksi TCP selama 87,78 detik**, serta **error bisnis di luar 429 sebesar 0%**. Pada beban sustained, 8.793 respons berhasil dan 33.718 respons ditolak dengan 429 sesuai pembatasan beban. Throughput sukses sebesar 97,59 respons per detik; respons 429 tidak dihitung sebagai throughput sukses.

Pengujian regresi mencakup pengiriman event 4 MiB, penolakan event yang melebihi batas, pemulihan backlog, serta pemrosesan ulang setelah gangguan penyimpanan consumer. Uji outage PVMBG selama 10 menit berhasil mempertahankan akses seismic dan memulihkan data volcanic tanpa restart. Bootstrap dengan kredensial baru dan enam volume terisolasi juga berhasil sampai data diterima ketiga consumer. Konfigurasi, hasil lengkap, dan batas interpretasi tersedia pada [bukti pengujian akhir](docs/evidence/final-2026-10-08/README.md). [Audit persyaratan](docs/requirements-audit.md) memuat pemetaan implementasi ke spesifikasi beserta temuan yang masih terbuka.

## Kontribusi kelompok

| Anggota | Perancangan dan validasi | Implementasi |
| --- | --- | --- |
| **Dzaky Aurelia Fawwaz (13523065)** | Merancang arsitektur dan mekanisme penyelesaian P1 hingga P5 | Mock BMKG dan PVMBG; Auth Service; ingest, pemetaan, korelasi tsunami, transaksi, dan watermark Aggregator; relay dan producer Kafka; ketiga consumer; infrastruktur, integrasi, serta pengujian lintas komponen. |
| **Muhammad Alfansya (13523005)** | Membantu verifikasi dan validasi kesesuaian rancangan dengan spesifikasi | Query dan API internal Aggregator; Client API beserta proteksi beban, pagination, dan galat; CLI Tim Lapangan; script load test k6. |

Kedua anggota berkontribusi pada keseluruhan penulisan laporan.
