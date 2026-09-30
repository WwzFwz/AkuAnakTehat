# Fondasi lokal

[Peta repository](../README.md)

Compose menyediakan PostgreSQL, Redis, Kafka KRaft, inisialisasi topic, dan empat service fondasi: `bmkg-mock`, `pvmbg-mock`, `auth-service`, `client-api`. Aggregator dan consumer belum menjadi service Compose. Query client-api dengan token valid dapat menghasilkan `503` sampai Aggregator tersedia.

Konfigurasi Compose sudah divalidasi secara statis. Build image, health runtime, dan persistence setelah restart masih perlu diuji; ini belum bukti demo P1–P5.

## Menjalankan

Prasyarat: Go 1.24.2 dan Docker Engine dengan Compose v2. Linux/macOS dengan Make: `make up`.

Windows PowerShell, tanpa Make atau WSL:

```powershell
powershell -NoProfile -File scripts/secrets/generate.ps1
docker compose config --quiet
docker compose up -d --build
```

`docker compose down` menghentikan container dan mempertahankan named volume. Jangan menghapus volume hanya untuk menjalankan ulang sistem.

| Komponen | Akses |
| --- | --- |
| BMKG | `http://127.0.0.1:8081` |
| PVMBG | `http://127.0.0.1:8082` |
| Auth | `http://127.0.0.1:8090` |
| Client API | `http://127.0.0.1:8080` |
| PostgreSQL | `canonical-db:5432`, hanya `store_net` |
| Redis | `auth-store:6379`, hanya `auth_net` |
| Kafka | `kafka:9092`, hanya `bus_net` |

Port HTTP dipublikasikan pada loopback. Database, Redis, dan Kafka tidak memiliki port host. Aggregator kelak memerlukan `source_net`, `store_net`, `edge_net`, dan `bus_net`; client-api tetap hanya di `edge_net`.

Kredensial per service di `env/*.env` dan kunci lokal di `env/keys/*.pem` diabaikan Git. `env/demo-clients.json` memuat kredensial demo plaintext lokal; jangan masukkan ke evidence/log. `env/clients.json` hanya berisi hash SHA-256. Generator tidak merotasi kredensial yang sudah ada.

Image dipin pada `postgres:16.8-alpine`, `redis:7.4.2-alpine`, dan `apache/kafka:3.9.1`. Topologi ini untuk demonstrasi lokal satu broker, bukan HA. PostgreSQL menyiapkan database kosong; skema milik migrasi Aggregator.

Referensi konfigurasi: [Compose startup order](https://docs.docker.com/compose/how-tos/startup-order/) dan [image resmi Kafka](https://hub.docker.com/r/apache/kafka/).
