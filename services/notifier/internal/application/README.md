# application

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: apply.go, apply_test.go.

Service.Handle memeriksa dedup (hazard_id, version), mengirim simulasi hanya untuk SIAGA/AWAS, lalu menulis marker persisten. NORMAL/WASPADA tetap dicatat sebagai selesai tanpa kirim. Crash setelah kirim sebelum marker dapat menghasilkan duplikat; tidak ada klaim exactly-once. Pemrosesan sequential oleh satu instance pemilik SQLite.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
