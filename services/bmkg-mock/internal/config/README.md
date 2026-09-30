# config

Membaca environment dan memvalidasi hash, alamat, interval serta delay sebelum startup.

**Status:** implementasi fondasi tersedia; Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap.

## File saat ini

- `config.go`

Nama environment dan batas nilainya mengikuti kontrak HTTP mock; tidak memakai secret default.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.
