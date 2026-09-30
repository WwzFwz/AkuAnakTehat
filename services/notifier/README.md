# notifier

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Consumer notifikasi SIAGA/AWAS dengan dedup persisten.

**Pemilik utama:** C. **Port rencana:** 8092. Semua komponen masih berupa dokumentasi; tidak ada binary, module Go, Dockerfile, atau endpoint yang sudah berjalan.

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`cmd/notifier`](cmd/notifier/README.md) | C | Inti/pendukung | Composition root notifier; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | C | Inti/pendukung | Konfigurasi lokal notifier; nama variabel berikut adalah usulan yang perlu disepakati. |
| [`internal/contract`](internal/contract/README.md) | C | Inti/pendukung | Kontrak event milik notifier; salinan lokal yang kompatibel dengan envelope producer. |
| [`internal/consumer`](internal/consumer/README.md) | C | Inti/pendukung | Loop konsumsi Kafka independen untuk notifier. |
| [`internal/application`](internal/application/README.md) | C | Inti/pendukung | Keputusan pengiriman peringatan dan dedup per versi. |
| [`internal/dedup`](internal/dedup/README.md) | C | Inti/pendukung | Pencatatan pasangan hazard_id/version yang sudah selesai diproses notifier. |
| [`internal/sender`](internal/sender/README.md) | C | Inti/pendukung | Adapter simulasi pengiriman peringatan. |
| [`internal/dlq`](internal/dlq/README.md) | C | Inti/pendukung | Penerbitan pesan gagal dengan konteks asal consumer. |
| [`internal/http`](internal/http/README.md) | C | Inti/pendukung | Endpoint health internal notifier. |

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
