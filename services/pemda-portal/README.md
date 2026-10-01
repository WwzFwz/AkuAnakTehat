# pemda-portal

[Peta repository](../../README.md) · [Kontrak](../../docs/api/consumers.md)

View hazard terbaru dari Kafka, tersimpan pada SQLite milik service.

**Pemilik:** C. **Status:** implementasi 3C tersedia. Go 1.24.2; franz-go 1.18.1; modernc SQLite 1.36.1. Satu module dan image mandiri, runtime non-root 10001, tanpa shared business package.

`docker compose --profile demo up -d --build pemda-portal` mengaktifkan subscriber ketiga tanpa mengubah producer.

Group default `pemda-portal`, topic `bnpb.hazard-events.v1`; terhubung ke bus_net dan consumer_net untuk port demo loopback. Named volume `pemda-data` mempertahankan SQLite setelah restart. Jalankan satu instance untuk satu volume; menghapus SQLite tanpa mengatur ulang offset Kafka tidak membangun ulang view secara otomatis.

Endpoint internal: `http://127.0.0.1:8093/health`, `/ready`, `/view`. Endpoint baca memiliki pagination; detail konfigurasi dan respons ada di [kontrak consumer](../../docs/api/consumers.md).

## Komponen

- [cmd/pemda-portal](cmd/pemda-portal/README.md): lifecycle dan wiring.
- [config](internal/config/README.md): config.go.
- [contract](internal/contract/README.md): event.go, event_test.go.
- [consumer](internal/consumer/README.md): consumer.go, consumer_test.go.
- [application](internal/application/README.md): apply.go.
- [store](internal/store/README.md): sqlite.go, sqlite_test.go.
- [http](internal/http/README.md): handler.go.

## Verifikasi

`go test ./...` dan `go vet ./...` dari folder service. `make events-check` dari root menguji aliran sumber, replay, restart, DLQ, subscriber tambahan, consumer offline, serta outage broker. [Hasil](../../docs/evidence/events/README.md).
