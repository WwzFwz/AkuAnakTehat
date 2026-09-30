# config

Membaca environment dan memvalidasi hash, alamat, interval serta delay sebelum startup.

**Status:** implementasi fondasi tersedia; pemeriksaan kompilasi dilakukan pada tahap ini. Pengujian perilaku/integrasi belum dilakukan.

## File saat ini

- `config.go`

Nama environment dan batas nilainya mengikuti kontrak HTTP mock; tidak memakai secret default.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.
