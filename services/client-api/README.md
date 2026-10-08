# client-api

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

API downstream dengan verifikasi JWT, otorisasi, dan proyeksi.

**Port default:** 8080. Implementasi terhubung melalui Compose. Cakupan dan batas verifikasi tersedia pada [audit persyaratan](../../docs/requirements-audit.md).

## Komponen

| Folder | Tanggung jawab |
| --- | --- |
| [`internal/authn`](internal/authn/README.md) | Membuktikan identitas client melalui verifikasi JWT dengan public key lokal. |
| [`internal/authz`](internal/authz/README.md) | Menentukan apakah scope terverifikasi boleh memenuhi permintaan. |
| [`internal/projection`](internal/projection/README.md) | Menyusun respons menggunakan allowlist field sesuai hak akses. |
| [`internal/application`](internal/application/README.md) | Orkestrasi baca downstream: otorisasi, panggil Aggregator, lalu proyeksi. |
| [`internal/adapter/outbound/aggregator`](internal/adapter/outbound/aggregator/README.md) | Klien HTTP untuk API internal Aggregator. |
| [`internal/adapter/inbound/http`](internal/adapter/inbound/http/README.md) | Endpoint publik untuk Media, Tim Lapangan, dan BNPB Ops. |
| [`internal/middleware`](internal/middleware/README.md) | Proteksi trafik inti dan batas resource jalur baca. |
| [`internal/cache`](internal/cache/README.md) | Rencana tambahan: singleflight dan micro-cache sebelum proyeksi. |
| [`cmd/client-api`](cmd/client-api/README.md) | Composition root client-api; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | Konfigurasi lokal client-api; nama variabel mengikuti config.go dan .env.example. |
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

Konfigurasi dan endpoint: [kontrak HTTP](../../docs/api/client-http.md).
Dari folder service: `go run ./cmd/client-api`. Variabel secret/key wajib diisi lewat bootstrap lokal.
Pemeriksaan module: `go test ./...` dan `go vet ./...`. Suite integrasi HTTP/Redis dijalankan terpisah dari root: `go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m`.
Aggregator query sudah tersedia; request data bergantung pada readiness Aggregator. Cache tetap belum diimplementasikan.
