# workflows

[Peta repository](../../README.md)

Rencana pemeriksaan CI per module dan pemeriksaan kebocoran secret.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `ci.yml` | Build/test/vet tiap module secara terisolasi dan secret scan. |

## Kontrak dan alur

- Matrix delapan service; gunakan GOWORK=off untuk membuktikan module tidak bergantung sibling.
- Pipeline baru dibuat ketika file kode dan go.mod tersedia.

## Dependensi

- GitHub Actions, Go toolchain yang dipilih tim, dan secret scanner.

## Aturan penting

- Tidak ada workflow executable pada tahap dokumentasi ini.
- Pin toolchain/action yang disepakati; jangan mengklaim scan bersih sebelum dijalankan.
- Jangan memasukkan secret ke log CI atau artefak publik.
- Build sukses bukan bukti semua skenario tugas telah selesai.

## Langkah implementasi dan verifikasi

- Tambahkan workflow setelah bootstrap module.
- Pisahkan pemeriksaan unit dari demo/integrasi yang memerlukan Compose.
