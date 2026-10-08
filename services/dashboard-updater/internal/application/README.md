# application

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

Service.Handle meneruskan event valid ke port View.Apply. Adapter melakukan satu UPSERT atomik yang hanya mengganti baris ketika version masuk lebih tinggi. Tidak ada request ke Aggregator.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [apply.go](apply.go) | Menerapkan snapshot event ke view SQLite melalui port lokal. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
