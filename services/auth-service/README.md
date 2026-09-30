# auth-service

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Penerbit token dan pemilik auth-store.

**Pemilik utama:** B. **Port rencana:** 8090. Module Go, Dockerfile, composition root, dan endpoint fondasi sudah tersedia. Verifikasi kompilasi berhasil; integrasi runtime dan bukti demo belum selesai.

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`internal/domain`](internal/domain/README.md) | B | Inti/pendukung | Identitas client, scope server, dan lifecycle keluarga refresh token. |
| [`internal/application`](internal/application/README.md) | B | Inti/pendukung | Use case pertukaran kredensial dan refresh tanpa login ulang pada alur normal. |
| [`internal/token`](internal/token/README.md) | B | Inti/pendukung | Penandatangan JWT Ed25519 dan pembangkit token acak. |
| [`internal/adapter/outbound/redis`](internal/adapter/outbound/redis/README.md) | B | Inti/pendukung | Persistensi refresh token dan status keluarga dengan TTL. |
| [`internal/adapter/inbound/http`](internal/adapter/inbound/http/README.md) | B | Inti/pendukung | Transport penerbitan token dan proteksi endpoint sensitif. |
| [`cmd/auth-service`](cmd/auth-service/README.md) | B | Inti/pendukung | Composition root auth-service; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | B | Inti/pendukung | Konfigurasi lokal auth-service; nama variabel berikut adalah usulan yang perlu disepakati. |
| [`internal/observability`](internal/observability/README.md) | B | Inti/pendukung | Logging terstruktur, correlation ID, serta liveness/readiness milik service. |

## Berkas tingkat service yang direncanakan

| File | Tanggung jawab |
| --- | --- |
| `go.mod` | Satu module mandiri untuk service ini; versi Go dipilih dan dipin saat implementasi. |
| `Dockerfile` | Build dengan konteks folder service sendiri dan runtime minimal non-root. |
| `.dockerignore` | Mengecualikan secret, artefak lokal, dan berkas yang tidak diperlukan saat build. |

## Batas dan urutan kerja

- Tidak meng-import business logic atau DTO dari module service lain.
- Interface kecil didefinisikan oleh package pemakai; concrete adapter dirangkai melalui composition root.
- Sepakati kontrak → implementasikan jalur inti → hubungkan dependensi → jalankan skenario tugas → kumpulkan bukti.
- Timeout lokal, batas konkurensi yang relevan, health, log terstruktur, dan correlation ID termasuk baseline.
- Cache, LISTEN/NOTIFY, schema_observations, dan propagasi deadline lewat header adalah tambahan; jangan menjadikannya prasyarat fungsi inti.
- Service dapat dimulai sebagai proses sendiri; kesiapan dependensi dilaporkan oleh readiness. Jangan mengembalikan sukses palsu untuk fitur yang belum dibuat.

## Menjalankan fondasi

Konfigurasi dan endpoint: [kontrak HTTP](../../docs/api/token-http.md).
Dari folder service: `go run ./cmd/auth-service`. Variabel secret/key wajib diisi lewat bootstrap lokal.
Verifikasi awal: `go test ./...` (kompilasi; belum ada test suite perilaku).
Rotasi Redis belum diverifikasi terhadap Redis nyata; jangan menganggap fondasi ini bukti P3.
