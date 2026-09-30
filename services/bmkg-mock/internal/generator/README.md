# generator

Loop periodik yang dapat dibatalkan. ID runtime unik per proses; record baru maksimal setiap 10 detik.

**Status:** implementasi fondasi tersedia; pemeriksaan kompilasi dilakukan pada tahap ini. Pengujian perilaku/integrasi belum dilakukan.

## File saat ini

- `generator.go`

Generator menyiapkan warning sebelum/sesudah event serta eskalasi ID warning yang sama. Hanya `Run` yang memanggil `Tick` secara serial.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.
