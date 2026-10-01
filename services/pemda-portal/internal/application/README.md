# application

[Panduan service](../../README.md) · [Kontrak consumer](../../../../docs/api/consumers.md)

**Pemilik:** C. **Status:** diimplementasikan.

Berkas: apply.go.

Service.Handle meneruskan event valid ke port View.Apply. Adapter melakukan satu UPSERT atomik yang hanya mengganti baris ketika version masuk lebih tinggi. Tidak ada request ke Aggregator.

Jalankan `go test ./...` dan `go vet ./...` dari module service. Integrasi Kafka/SQLite diuji melalui `make events-check` dari root; lihat [bukti](../../../../docs/evidence/events/README.md).
