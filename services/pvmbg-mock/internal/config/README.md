# config

Membaca environment dan memvalidasi hash, alamat, interval serta delay sebelum startup.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Nama environment dan batas nilainya mengikuti kontrak HTTP mock; tidak memakai secret default.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [config.go](config.go) | Membaca environment dan memvalidasi konfigurasi sebelum service dijalankan. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
