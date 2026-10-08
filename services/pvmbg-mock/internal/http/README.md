# http

Route data/health/readiness, autentikasi, filter since, delay dan log JSON berkorelasi.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [common.go](common.go) | Helper respons HTTP, correlation ID, logging, dan penanganan request milik mock. |
| [router.go](router.go) | Mendaftarkan endpoint, memvalidasi request, dan membentuk respons HTTP. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
