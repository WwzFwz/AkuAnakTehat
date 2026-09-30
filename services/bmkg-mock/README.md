# bmkg-mock

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Mock gempa dan warning tsunami dengan kontrak independen.

**Pemilik utama:** A. **Port rencana:** 8081. Semua komponen masih berupa dokumentasi; tidak ada binary, module Go, Dockerfile, atau endpoint yang sudah berjalan.

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`cmd/bmkg-mock`](cmd/bmkg-mock/README.md) | A | Inti/pendukung | Composition root bmkg-mock; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | A | Inti/pendukung | Konfigurasi lokal bmkg-mock; nama variabel berikut adalah usulan yang perlu disepakati. |
| [`internal/domain`](internal/domain/README.md) | A | Inti/pendukung | Kontrak data BMKG milik mock; tidak bergantung pada model BNPB. |
| [`internal/store`](internal/store/README.md) | A | Inti/pendukung | Penyimpanan in-memory BMKG, dimiliki mock dan aman untuk akses bersamaan. |
| [`internal/generator`](internal/generator/README.md) | A | Inti/pendukung | Pembangkitan data baru BMKG untuk membuktikan polling berkala. |
| [`internal/auth`](internal/auth/README.md) | A | Inti/pendukung | Validasi kredensial domain BMKG. |
| [`internal/http`](internal/http/README.md) | A | Inti/pendukung | Transport endpoint BMKG dengan autentikasi dan perilaku simulasi. |
| [`seed`](seed/README.md) | A | Inti/pendukung | Fixture historis milik BMKG; data demo, bukan data bencana live. |

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
