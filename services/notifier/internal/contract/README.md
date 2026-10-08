# contract

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

Decode memvalidasi envelope v1, identitas/key, version, timestamp, koordinat, severity, dan field hazard wajib. Unknown fields diabaikan saat validasi tetapi byte payload lengkap tetap disimpan; angka tambahan tidak dibulatkan ke float64. Kontrak ini lokal service, tidak meng-import Aggregator.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [event.go](event.go) | Validasi envelope v1 dan hazard wajib dengan batas 4 MiB; kontrak lokal service. |
| [event_test.go](event_test.go) | Pengujian `TestContract`. |

## Perilaku dan batas saat ini

Envelope lengkap dibatasi 4 MiB. Producer dan topic Kafka menyediakan batas batch 5 MiB, sedangkan fetch consumer 6 MiB. Angka tersebut membatasi objek yang berbeda dan menyediakan ruang untuk overhead transport.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
