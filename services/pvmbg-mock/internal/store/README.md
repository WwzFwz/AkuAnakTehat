# store

Storage in-memory aman untuk akses bersamaan; snapshot tidak membocorkan slice/pointer mutable.

**Status:** implementasi fondasi tersedia; pemeriksaan kompilasi dilakukan pada tahap ini. Pengujian perilaku/integrasi belum dilakukan.

## File saat ini

- `memory.go`

Filter `since` inklusif. BMKG warning memakai waktu perubahan tersembunyi; PVMBG memakai `reported_at`. Lock dilepas sebelum delay/network. Data runtime hilang saat restart dan belum dibatasi retensi.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.
