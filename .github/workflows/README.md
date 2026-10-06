# workflows

[Peta repository](../../README.md)

Pemeriksaan CI per module dan pemeriksaan kebocoran secret.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** ci.yml tersedia; eksekusi GitHub Actions belum diverifikasi.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `ci.yml` | Build/test/vet tiap module secara terisolasi dan secret scan. |

## Kontrak dan alur

- scripts/check/check.sh menjalankan tests generator lalu test/vet setiap service dan tool yang memiliki go.mod, dengan GOWORK=off. Folder rancangan tanpa module dilewati.
- CI memakai Go 1.24.2, memvalidasi Compose tanpa mencetak env, membangun/menjalankan service fondasi, ingest dan consumer, lalu menjalankan suite runtime `scripts/check/foundation_test.go`. Stack dihentikan pada akhir job. UID/GID container auth/client mengikuti pemilik secret pada runner Linux. Job Gitleaks 8.24.2 memindai histori Git dengan output disensor.

## Dependensi

- GitHub Actions, Go toolchain yang dipilih tim, dan secret scanner.

## Aturan penting

- Setelah fondasi, workflow menguji transaksi PostgreSQL, ingest, Kafka/SQLite/DLQ, query/otorisasi, expiry token alami, rebuild independen, dan load test k6. Cakupan serta hasil lokal ada di [verifikasi integrasi](../../docs/evidence/integration/README.md).
- Pin toolchain/action yang disepakati; jangan mengklaim scan bersih sebelum dijalankan.
- Jangan memasukkan secret ke log CI atau artefak publik.
- Build sukses bukan bukti semua skenario tugas telah selesai.

## Langkah implementasi dan verifikasi

- Verifikasi hasil workflow saat dijalankan GitHub Actions.
- Pisahkan pemeriksaan unit dari demo/integrasi yang memerlukan Compose.
