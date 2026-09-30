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

- scripts/check/check.sh menjalankan tests generator lalu test/vet setiap service yang memiliki go.mod, dengan GOWORK=off. Folder rancangan tanpa module dilewati.
- CI memakai Go 1.24.2, memvalidasi Compose tanpa mencetak env, membangun/menjalankan empat image fondasi, lalu menjalankan suite runtime `scripts/check/foundation_test.go`. Stack dihentikan pada akhir job. UID/GID container auth/client mengikuti pemilik secret pada runner Linux. Job Gitleaks 8.24.2 memindai histori Git dengan output disensor.

## Dependensi

- GitHub Actions, Go toolchain yang dipilih tim, dan secret scanner.

## Aturan penting

- Workflow executable sudah tersedia untuk fondasi; tidak menjalankan demo P1–P5.
- Setelah fondasi, workflow menguji transaksi ingest dalam schema PostgreSQL terisolasi dan alur live mock → canonical store/outbox, termasuk outage/recovery. Ini tetap bukan P1–P5 lengkap.
- Pin toolchain/action yang disepakati; jangan mengklaim scan bersih sebelum dijalankan.
- Jangan memasukkan secret ke log CI atau artefak publik.
- Build sukses bukan bukti semua skenario tugas telah selesai.

## Langkah implementasi dan verifikasi

- Verifikasi hasil workflow saat dijalankan GitHub Actions.
- Pisahkan pemeriksaan unit dari demo/integrasi yang memerlukan Compose.
