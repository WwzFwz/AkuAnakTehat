# sender

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

Log.Send menyimulasikan notifikasi melalui JSON log notification_simulated dengan event_id, hazard_id, version, severity, dan correlation_id. Tidak menghubungi penyedia pesan eksternal. Sender dapat diganti melalui port lokal application.Sender.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [log.go](log.go) | Mengirim simulasi notifikasi melalui log terstruktur. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
