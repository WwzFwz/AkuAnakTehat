# auth-service

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Penerbit token dan pemilik auth-store.

**Port default:** 8090. Implementasi terhubung melalui Compose. Cakupan dan batas verifikasi tersedia pada [audit persyaratan](../../docs/requirements-audit.md).

## Komponen

| Folder | Tanggung jawab |
| --- | --- |
| [`internal/domain`](internal/domain/README.md) | Identitas client, scope server, dan lifecycle keluarga refresh token. |
| [`internal/application`](internal/application/README.md) | Use case pertukaran kredensial dan refresh tanpa login ulang pada alur normal. |
| [`internal/token`](internal/token/README.md) | Penandatangan JWT Ed25519 dan pembangkit token acak. |
| [`internal/adapter/outbound/redis`](internal/adapter/outbound/redis/README.md) | Persistensi refresh token dan status keluarga dengan TTL. |
| [`internal/adapter/inbound/http`](internal/adapter/inbound/http/README.md) | Transport penerbitan token dan proteksi endpoint sensitif. |
| [`cmd/auth-service`](cmd/auth-service/README.md) | Composition root auth-service; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | Konfigurasi lokal auth-service; nama variabel mengikuti config.go dan .env.example. |
| [`internal/observability`](internal/observability/README.md) | Logging terstruktur, correlation ID, serta liveness/readiness milik service. |

## Berkas tingkat service

| File | Tanggung jawab |
| --- | --- |
| `go.mod` | Satu module mandiri untuk service ini; versi Go dipin ke 1.24.2. |
| `Dockerfile` | Build dengan konteks folder service sendiri dan runtime minimal non-root. |
| `.dockerignore` | Mengecualikan secret, artefak lokal, dan berkas yang tidak diperlukan saat build. |

## Batas komponen

- Tidak meng-import business logic atau DTO dari module service lain.
- Interface kecil didefinisikan oleh package pemakai; concrete adapter dirangkai melalui composition root.
- Timeout lokal, batas konkurensi yang relevan, health, log terstruktur, dan correlation ID termasuk baseline.
- Cache, LISTEN/NOTIFY, schema_observations, dan propagasi deadline lewat header adalah tambahan; jangan menjadikannya prasyarat fungsi inti.
- Service dapat dimulai sebagai proses sendiri; kesiapan dependensi dilaporkan oleh readiness. Jangan mengembalikan sukses palsu untuk fitur yang belum dibuat.

## Menjalankan

Konfigurasi dan endpoint: [kontrak HTTP](../../docs/api/token-http.md).
Dari folder service: `go run ./cmd/auth-service`. Variabel secret/key wajib diisi lewat bootstrap lokal.
Pemeriksaan module: `go test ./...` dan `go vet ./...`. Suite integrasi HTTP/Redis dijalankan terpisah dari root: `go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m`.
Rotasi, reuse token, refresh bersamaan, dan pemulihan sesi setelah restart Redis telah diuji pada stack lokal. Expiry alami dan refresh CLI juga telah diperiksa oleh TestNaturalExpiryAndFieldCLI; demo sinkron kelompok tetap terpisah.
