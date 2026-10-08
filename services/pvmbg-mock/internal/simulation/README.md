# simulation

State runtime PVMBG untuk schema v1/v2 dan outage hang/error, dengan notifikasi perubahan untuk request aktif.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Admin tetap responsif; disable outage membangunkan request hang. Perubahan skema hanya memengaruhi record yang dibuat sesudahnya.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [state.go](state.go) | State schema version dan outage yang dapat diubah saat mock berjalan. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
