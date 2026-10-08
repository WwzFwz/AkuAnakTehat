# generator

Loop periodik yang dapat dibatalkan. ID runtime unik per proses; record baru maksimal setiap 10 detik.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Generator menyiapkan warning sebelum/sesudah event serta eskalasi ID warning yang sama. Hanya `Run` yang memanggil `Tick` secara serial.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [generator.go](generator.go) | Menghasilkan record sintetis baru secara berkala sesuai konfigurasi. |
| [generator_test.go](generator_test.go) | Pengujian `TestWarningOrderingAndEscalationWatermark`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
