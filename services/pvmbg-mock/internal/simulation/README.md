# simulation

State runtime PVMBG untuk schema v1/v2 dan outage hang/error, dengan notifikasi perubahan untuk request aktif.

**Status:** implementasi fondasi tersedia; pemeriksaan kompilasi dilakukan pada tahap ini. Pengujian perilaku/integrasi belum dilakukan.

## File saat ini

- `state.go`

Admin tetap responsif; disable outage membangunkan request hang. Perubahan skema hanya memengaruhi record yang dibuat sesudahnya.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.
