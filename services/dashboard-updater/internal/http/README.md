# http

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

GET /health melaporkan proses hidup. GET /ready memeriksa SQLite dan Kafka dengan deadline total 2s; bukan jaminan lag nol. GET /view mengembalikan envelope mentah internal dengan limit 1..200 (default 100) dan cursor after eksklusif. Respons berbentuk {data: [...], next_cursor: "..."}; string kosong berarti halaman terakhir. Kegagalan DB menghasilkan 503, limit invalid 400. Port host hanya loopback, bukan API Media.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [handler.go](handler.go) | Melayani endpoint health, readiness, dan inspeksi state lokal consumer. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
