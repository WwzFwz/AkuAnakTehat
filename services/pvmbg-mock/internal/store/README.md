# store

Storage in-memory aman untuk akses bersamaan; snapshot tidak membocorkan slice/pointer mutable.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

Filter `since` inklusif. BMKG warning memakai waktu perubahan tersembunyi; PVMBG memakai `reported_at`. Lock dilepas sebelum delay/network. Data runtime hilang saat restart dan belum dibatasi retensi.

Lihat README service dan `docs/api/mock-http.md` pada root repository untuk kontrak lengkap.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [memory.go](memory.go) | Menyimpan data mock dalam memori dengan sinkronisasi dan filter waktu. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
