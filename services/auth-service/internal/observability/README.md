# observability

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Logging terstruktur, correlation ID, serta liveness/readiness milik service.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `logger.go` | Konfigurasi slog JSON dan redaksi field sensitif. |
| `correlation.go` | Validasi/bangkitkan ID dan simpan pada context. |
| `health.go` | Handler /health dan /ready serta ringkasan dependensi. |

## Kontrak dan alur

- /health menunjukkan proses hidup; /ready menunjukkan kesiapan melayani fungsi service.
- Field log minimum: service, correlation_id, operation, latency_ms, dan hasil.

## Dependensi

- Standard library; dipakai adapter, application, dan worker melalui dependensi yang sesuai.

## Aturan penting

- Jangan log Authorization, X-*-Key, token, password, atau payload raw sebelum proyeksi.
- ID diteruskan lewat HTTP/Kafka dalam alur yang sama; ukur latensi outbound.
- Aggregator tetap ready ketika sumber/Kafka mati jika fungsi yang bergantung DB masih dapat dilayani.
- Untuk service tanpa package ini, utilitas lokal ditempatkan pada http/consumer; tidak membuat shared module bisnis.

## Langkah implementasi dan verifikasi

- Sepakati format log lintas service.
- Pastikan trace script dapat mengikuti satu alur ingest sampai consumer.
