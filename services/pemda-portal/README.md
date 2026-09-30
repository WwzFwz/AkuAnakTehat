# pemda-portal

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Consumer ketiga untuk demo subscription baru tanpa perubahan producer.

**Pemilik utama:** C. **Port rencana:** 8093. Semua komponen masih berupa dokumentasi; tidak ada binary, module Go, Dockerfile, atau endpoint yang sudah berjalan.

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`cmd/pemda-portal`](cmd/pemda-portal/README.md) | C | Inti/pendukung | Composition root pemda-portal; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | C | Inti/pendukung | Konfigurasi lokal pemda-portal; nama variabel berikut adalah usulan yang perlu disepakati. |
| [`internal/contract`](internal/contract/README.md) | C | Inti/pendukung | Kontrak event milik pemda-portal; salinan lokal yang kompatibel dengan envelope producer. |
| [`internal/consumer`](internal/consumer/README.md) | C | Inti/pendukung | Loop konsumsi Kafka independen untuk pemda-portal. |
| [`internal/application`](internal/application/README.md) | C | Inti/pendukung | Penerapan event ke view hazard terbaru milik consumer. |
| [`internal/store`](internal/store/README.md) | C | Inti/pendukung | Penyimpanan view SQLite persisten milik pemda-portal. |
| [`internal/http`](internal/http/README.md) | C | Inti/pendukung | Endpoint pembacaan view lokal dan health consumer. |

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
