# pvmbg-mock

[Peta repository](../../README.md) · [Kontrak integrasi](../../docs/api/README.md)

Mock vulkanik dengan delay, outage, dan schema evolution.

**Pemilik utama:** A. **Port rencana:** 8082. Semua komponen masih berupa dokumentasi; tidak ada binary, module Go, Dockerfile, atau endpoint yang sudah berjalan.

## Komponen

| Folder | Pemilik | Tahap | Tujuan |
| --- | --- | --- | --- |
| [`cmd/pvmbg-mock`](cmd/pvmbg-mock/README.md) | A | Inti/pendukung | Composition root pvmbg-mock; tempat merangkai seluruh dependensi runtime. |
| [`internal/config`](internal/config/README.md) | A | Inti/pendukung | Konfigurasi lokal pvmbg-mock; nama variabel berikut adalah usulan yang perlu disepakati. |
| [`internal/domain`](internal/domain/README.md) | A | Inti/pendukung | Kontrak data PVMBG milik mock; tidak bergantung pada model BNPB. |
| [`internal/store`](internal/store/README.md) | A | Inti/pendukung | Penyimpanan in-memory PVMBG, dimiliki mock dan aman untuk akses bersamaan. |
| [`internal/generator`](internal/generator/README.md) | A | Inti/pendukung | Pembangkitan data baru PVMBG untuk membuktikan polling berkala. |
| [`internal/auth`](internal/auth/README.md) | A | Inti/pendukung | Validasi kredensial domain PVMBG. |
| [`internal/http`](internal/http/README.md) | A | Inti/pendukung | Transport endpoint PVMBG dengan autentikasi dan perilaku simulasi. |
| [`seed`](seed/README.md) | A | Inti/pendukung | Fixture historis milik PVMBG; data demo, bukan data bencana live. |
| [`internal/simulation`](internal/simulation/README.md) | A | Inti/pendukung | State dan kontrol runtime delay, outage, serta evolusi skema PVMBG. |

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
