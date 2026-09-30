# client-api

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

API downstream dengan verifikasi JWT, otorisasi, dan proyeksi.

**Pemilik utama:** B. **Port rencana:** 8080. Semua komponen masih berupa dokumentasi; tidak ada binary, module Go, Dockerfile, atau endpoint yang sudah berjalan.

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`internal/authn`](internal/authn/README.md) | B | Inti/pendukung | Membuktikan identitas client melalui verifikasi JWT dengan public key lokal. |
| [`internal/authz`](internal/authz/README.md) | B | Inti/pendukung | Menentukan apakah scope terverifikasi boleh memenuhi permintaan. |
| [`internal/projection`](internal/projection/README.md) | B | Inti/pendukung | Menyusun respons menggunakan allowlist field sesuai hak akses. |
| [`internal/application`](internal/application/README.md) | B | Inti/pendukung | Orkestrasi baca downstream: otorisasi, panggil Aggregator, lalu proyeksi. |
| [`internal/adapter/outbound/aggregator`](internal/adapter/outbound/aggregator/README.md) | B | Inti/pendukung | Klien HTTP untuk API internal Aggregator. |
| [`internal/adapter/inbound/http`](internal/adapter/inbound/http/README.md) | B | Inti/pendukung | Endpoint publik untuk Media, Tim Lapangan, dan BNPB Ops. |
| [`internal/middleware`](internal/middleware/README.md) | B | Inti/pendukung | Proteksi trafik inti dan batas resource jalur baca. |
| [`internal/cache`](internal/cache/README.md) | B | Tambahan | Rencana tambahan: singleflight dan micro-cache sebelum proyeksi. |
| [`cmd/client-api`](cmd/client-api/README.md) | B | Inti/pendukung | Composition root client-api; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | B | Inti/pendukung | Konfigurasi lokal client-api; nama variabel berikut adalah usulan yang perlu disepakati. |
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
