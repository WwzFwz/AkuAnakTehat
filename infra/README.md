# Fondasi lokal

[Peta repository](../README.md)

Compose menyediakan PostgreSQL, Redis, Kafka KRaft, init topic, empat service fondasi, serta Aggregator ingest/query. Relay, dashboard-updater, dan notifier aktif; pemda-portal tersedia lewat profile demo.

Build image, health HTTP, serta persistence PostgreSQL/Redis/Kafka setelah restart telah diuji pada Docker Desktop lokal. Cakupan dan perintah ulang tersedia di [hasil pengujian fondasi](../docs/evidence/foundation/README.md); ini belum bukti demo P1–P5 lengkap.

## Menjalankan

Prasyarat: Go 1.24.2 dan Docker Engine dengan Compose v2. Linux/macOS dengan Make: `make up`.

Untuk Compose manual pada Linux/macOS, jalankan `export LOCAL_UID=$(id -u) LOCAL_GID=$(id -g)` sebelum `docker compose up`. Make dan CI mengaturnya otomatis. Auth/client-api memakai UID/GID pemilik file agar bisa membaca bind mount secret dengan permission 0600; tidak perlu membuka permission private key ke semua user. Windows Docker Desktop memakai default user non-root 10001.

Windows PowerShell, tanpa Make atau WSL:

```powershell
powershell -NoProfile -File scripts/secrets/generate.ps1
docker compose config --quiet
docker compose up -d --build
```

`docker compose --profile demo up -d --build pemda-portal` menambah subscriber ketiga. `docker compose --profile demo down` menghentikan seluruh profile yang aktif dan mempertahankan volume.

`docker compose down` menghentikan container dan mempertahankan named volume. Jangan menghapus volume hanya untuk menjalankan ulang sistem.

`docker compose up -d --build --wait --wait-timeout 180` menunggu healthcheck. Mock/client-api memakai liveness, auth-service readiness Redis, dan Aggregator `/ready` yang memeriksa database query. Aggregator tidak membuka port host; network source/store/edge/bus. Consumer memakai bus_net/consumer_net tanpa store_net, readiness memeriksa Kafka dan SQLite.

Profile `test` berisi container `ingest-test` yang membuat schema PostgreSQL sementara lalu menghapusnya. Jalankan `docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test`; flag env-file eksplisit diperlukan pada Compose lokal yang tidak meneruskan env_file saat `run`. Lihat [bukti ingest](../docs/evidence/ingest/README.md).

| Komponen | Akses |
| --- | --- |
| BMKG | `http://127.0.0.1:8081` |
| PVMBG | `http://127.0.0.1:8082` |
| Auth | `http://127.0.0.1:8090` |
| Client API | `http://127.0.0.1:8080` |
| Dashboard | `http://127.0.0.1:8091/view` |
| Notifier audit | `http://127.0.0.1:8092/processed` |
| Pemda (profile demo) | `http://127.0.0.1:8093/view` |
| PostgreSQL | `canonical-db:5432`, hanya `store_net` |
| Redis | `auth-store:6379`, hanya `auth_net` |
| Kafka | `kafka:9092`, hanya `bus_net` |

Port HTTP dipublikasikan pada loopback. Database, Redis, dan Kafka tidak memiliki port host. Aggregator memakai `source_net`, `store_net`, `edge_net`, dan `bus_net`; client-api tetap hanya di `edge_net`.

Kredensial per service di `env/*.env` dan kunci lokal di `env/keys/*.pem` diabaikan Git. `env/demo-clients.json` memuat kredensial demo plaintext lokal; jangan masukkan ke evidence/log. `env/clients.json` hanya berisi hash SHA-256. Generator tidak merotasi kredensial yang sudah ada.

Image dipin pada `postgres:16.8-alpine`, `redis:7.4.2-alpine`, dan `apache/kafka:3.9.1`. Topologi ini untuk demonstrasi lokal satu broker, bukan HA. PostgreSQL menyiapkan database kosong; skema milik migrasi Aggregator.

Referensi konfigurasi: [Compose startup order](https://docs.docker.com/compose/how-tos/startup-order/) dan [image resmi Kafka](https://hub.docker.com/r/apache/kafka/).
