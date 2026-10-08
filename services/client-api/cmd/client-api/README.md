# client-api

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Composition root client-api; tempat merangkai seluruh dependensi runtime.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- main memulai service dan menghentikannya saat SIGTERM; bukan tempat logika bisnis.

## Dependensi

- Package milik service ini; tidak meng-import module service lain.

## Aturan penting

- Konfigurasi atau key tidak valid menggagalkan startup. Readiness memeriksa dependensi runtime; mekanisme retry mengikuti adapter, bukan retry startup generik.
- Jangan menjadikan semua dependency wajib hidup untuk liveness.
- Shutdown HTTP dibatasi oleh context; liveness tidak menunggu semua dependensi sehat.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [main.go](main.go) | Memuat konfigurasi, merangkai dependensi, menjalankan worker dan HTTP, serta menangani shutdown. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
