# seed

Fixture sintetis historis dengan ID tetap, di-embed ke binary dan dimuat saat startup.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Seed bertanggal 1 September 2026 UTC. Query awal tanpa `since` mengambil seluruh seed. Fixture ini bukan kejadian bencana nyata.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [embed.go](embed.go) | Embed dan memuat seed JSON lokal yang digunakan saat startup. |
| [seismic-events.json](seismic-events.json) | Seed historis gempa sintetis sesuai kontrak BMKG. |
| [tsunami-warnings.json](tsunami-warnings.json) | Seed warning tsunami yang merujuk event gempa sintetis. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
