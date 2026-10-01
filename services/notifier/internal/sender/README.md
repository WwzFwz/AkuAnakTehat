# sender

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: log.go.

Log.Send menyimulasikan notifikasi melalui JSON log notification_simulated dengan event_id, hazard_id, version, severity, dan correlation_id. Tidak menghubungi penyedia pesan eksternal. Sender dapat diganti melalui port lokal application.Sender.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
