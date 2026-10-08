# notifier

[Peta repository](../../README.md) · [Kontrak](../../docs/api/consumers.md)

Notifikasi SIAGA/AWAS dengan dedup persisten dan pengiriman simulasi.

**Status:** diimplementasikan. Go 1.24.2; franz-go 1.18.1; modernc SQLite 1.36.1. Satu module dan image mandiri, runtime non-root 10001, tanpa shared business package.

`docker compose up -d --build notifier` dari root menjalankan service bersama dependensi Kafka.

Group default `notifier`, topic `bnpb.hazard-events.v1`; terhubung ke bus_net dan consumer_net untuk port demo loopback. Named volume `notifier-data` mempertahankan SQLite setelah restart. Jalankan satu instance untuk satu volume; menghapus SQLite tanpa mengatur ulang offset Kafka tidak membangun ulang view secara otomatis.

Endpoint internal: `http://127.0.0.1:8092/health`, `/ready`, `/processed`. Endpoint baca memiliki pagination; detail konfigurasi dan respons ada di [kontrak consumer](../../docs/api/consumers.md).

## Komponen

- [cmd/notifier](cmd/notifier/README.md): lifecycle dan wiring.
- [config](internal/config/README.md): config.go.
- [contract](internal/contract/README.md): event.go, event_test.go.
- [consumer](internal/consumer/README.md): konsumsi Kafka, retry, DLQ, offset, tracing, dan pengujian pemulihan.
- [application](internal/application/README.md): apply.go, apply_test.go.
- [http](internal/http/README.md): handler.go.
- [dedup](internal/dedup/README.md): sqlite.go.
- [sender](internal/sender/README.md): log.go.
- [dlq](internal/dlq/README.md): record.go.

## Verifikasi

`go test ./...` dan `go vet ./...` dari folder service. `make events-check` dari root menguji aliran sumber, replay, restart, DLQ, subscriber tambahan, consumer offline, serta outage broker. [Hasil](../../docs/evidence/events/README.md).

Lihat [audit persyaratan](../../docs/requirements-audit.md) untuk hubungan implementasi dengan spesifikasi dan bukti pengujian terbaru.
