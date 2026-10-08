# observability

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Logging terstruktur, correlation ID, serta liveness/readiness milik service.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- /health menunjukkan proses hidup; /ready/ingest memeriksa database ingest; /ready memeriksa pool query.
- Log request dan outbound memuat identitas service, correlation_id, serta latency_ms; nama operasi dan hasil mengikuti jenis log.

## Dependensi

- Dipakai komponen dalam module service ini; tidak dibagikan sebagai library bisnis lintas service.

## Aturan penting

- Jangan log Authorization, X-*-Key, token, password, atau payload raw sebelum proyeksi.
- Correlation ID diteruskan ke dependensi melalui adapter yang relevan.
- Readiness ingest tetap 200 ketika sumber mati selama database tersedia; status sumber tersimpan terpisah. Relay Kafka berjalan terpisah dari readiness ingest.
- Untuk service tanpa package ini, utilitas lokal ditempatkan pada http/consumer; tidak membuat shared module bisnis.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [health.go](health.go) | Handler health/readiness dan log request HTTP dengan correlation ID. |
| [outbound.go](outbound.go) | Context correlation ID serta tracing PostgreSQL, Kafka, dan panggilan outbound. |
| [outbound_test.go](outbound_test.go) | Memeriksa correlation ID dan latency tanpa kebocoran SQL, parameter, atau error driver. |

## Perilaku dan batas saat ini

Logger JSON dibuat di cmd/aggregator/main.go. outbound.go memakai pgx dan franz-go untuk hook driver. ID operasi protokol Kafka dapat berbeda dari ID event; publish per record tetap membawa correlation ID ingest. SQL, parameter, dan pesan error driver tidak dicetak.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
