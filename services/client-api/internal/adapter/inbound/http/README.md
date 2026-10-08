# http

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Endpoint publik untuk Media, Tim Lapangan, dan BNPB Ops.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- GET /v1/hazards, /v1/hazards/seismic, /v1/hazards/volcanic, /v1/hazards/{id}.
- Permintaan mentah menggunakan include=raw, fields, atau endpoint /v1/hazards/{id}/raw sesuai docs/api/client-http.md.

## Dependensi

- authn, middleware, application, config, observability.

## Aturan penting

- 401 token invalid;403 insufficient_scope;429 overload;503 dependency/source unavailable.
- Jangan mengembalikan data sebelum autentikasi, otorisasi, dan proyeksi.
- Cursor/pagination diteruskan sesuai kontrak Aggregator; jangan membuat urutan halaman berbeda.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [limits_router_test.go](limits_router_test.go) | Pengujian `TestListValidationErrorCodes`, `TestRouterDistinguishesConcurrencyRejection`. |
| [router.go](router.go) | Mendaftarkan endpoint, memvalidasi request, dan membentuk respons HTTP. |
| [router_test.go](router_test.go) | Pengujian `TestHTTPProjectionAndForbiddenRequests`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
