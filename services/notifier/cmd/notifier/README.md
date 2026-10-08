# notifier entrypoint

[Panduan service](../../README.md)

`main.go` membaca konfigurasi, membuka SQLite, merangkai application/consumer/HTTP, dan memasang JSON logger. SIGTERM membatalkan worker dan menghentikan HTTP. Kegagalan worker menghentikan proses (exit 1) agar Compose me-restart tanpa melompati offset belum selesai. Service tidak memerlukan akses PostgreSQL Aggregator.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [main.go](main.go) | Memuat konfigurasi, merangkai dependensi, menjalankan worker dan HTTP, serta menangani shutdown. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
