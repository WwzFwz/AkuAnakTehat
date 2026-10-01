# consumer

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: consumer.go, consumer_test.go.

franz-go membaca satu record setiap poll dengan auto-commit dimatikan dan rebalance ditahan selama proses. Efek bisnis dicoba maksimum tiga kali dengan timeout lokal 2s per percobaan dan jeda 200ms. Input invalid langsung ke DLQ. Commit offset hanya sesudah efek persisten selesai atau DLQ di-ACK. Kegagalan DLQ/commit menghentikan proses sehingga restart mengulang offset; tidak melompati record gagal. Group baru membaca dari earliest yang masih tersedia. Group lama melanjutkan offset tersimpan.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
