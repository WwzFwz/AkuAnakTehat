# http

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: handler.go.

GET /health melaporkan proses hidup. GET /ready memeriksa SQLite dan Kafka dengan deadline total 2s; bukan jaminan lag nol. GET /view mengembalikan envelope mentah internal dengan limit 1..200 (default 100) dan cursor after eksklusif. Respons berbentuk {data: [...], next_cursor: "..."}; string kosong berarti halaman terakhir. Kegagalan DB menghasilkan 503, limit invalid 400. Port host hanya loopback, bukan API Media.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
