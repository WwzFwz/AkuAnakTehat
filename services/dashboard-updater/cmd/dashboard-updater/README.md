# dashboard-updater entrypoint

[Panduan service](../../README.md)

`main.go` membaca konfigurasi, membuka SQLite, merangkai application/consumer/HTTP, dan memasang JSON logger. SIGTERM membatalkan worker dan menghentikan HTTP. Kegagalan worker menghentikan proses (exit 1) agar Compose me-restart tanpa melompati offset belum selesai. Service tidak memerlukan akses PostgreSQL Aggregator.
