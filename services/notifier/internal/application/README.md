# application

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

Service.Handle memeriksa dedup (hazard_id, version), mengirim simulasi hanya untuk SIAGA/AWAS, lalu menulis marker persisten. NORMAL/WASPADA tetap dicatat sebagai selesai tanpa kirim. Crash setelah kirim sebelum marker dapat menghasilkan duplikat; tidak ada klaim exactly-once. Pemrosesan sequential oleh satu instance pemilik SQLite.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [apply.go](apply.go) | Dedup versi hazard, filter SIAGA/AWAS, pengiriman simulasi, lalu marker persisten. |
| [apply_test.go](apply_test.go) | Pengujian `TestPersistentDedupAndDeliveryOrder`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
