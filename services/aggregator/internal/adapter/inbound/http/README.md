# http

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Transport API internal Aggregator: routing, autentikasi internal, validasi input, dan serialisasi hasil query.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- GET /internal/hazards menerima filter yang disepakati di docs/api.
- GET /internal/hazards/{id} membaca satu hazard.
- Header wajib bisnis: X-Internal-Key; X-Correlation-ID diteruskan.

## Dependensi

- application/query, config, dan observability lokal.
- Tidak membuka koneksi PostgreSQL sendiri.

## Aturan penting

- Validasi enum, timestamp, limit, dan cursor sebelum query.
- Timeout lokal aktif sejak baseline. X-Request-Deadline hanya rencana tambahan dan harus dibatasi server.
- Port internal tidak dipublikasikan ke host.
- Tidak mengembalikan detail error SQL atau secret.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [auth.go](auth.go) | Memeriksa kredensial internal tanpa membocorkan nilai secret. |
| [cursor.go](cursor.go) | Menerjemahkan query HTTP ke filter dan memvalidasi cursor. |
| [errors.go](errors.go) | Memetakan error application ke status dan payload galat HTTP. |
| [router.go](router.go) | Mendaftarkan endpoint, memvalidasi request, dan membentuk respons HTTP. |
| [router_test.go](router_test.go) | Pengujian `TestListSuccessAndCorrelation`, `TestListRejectsCredentialsAndQuery`, `TestDetailNotFound`, `TestQueryTimeoutIsUnavailable`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
