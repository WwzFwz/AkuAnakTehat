# contract

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: event.go, event_test.go.

Decode memvalidasi envelope v1, identitas/key, version, timestamp, koordinat, severity, dan field hazard wajib. Unknown fields diabaikan saat validasi tetapi byte payload lengkap tetap disimpan; angka tambahan tidak dibulatkan ke float64. Kontrak ini lokal service, tidak meng-import Aggregator.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
