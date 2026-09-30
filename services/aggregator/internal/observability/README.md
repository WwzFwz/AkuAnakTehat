# observability

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Logging terstruktur, correlation ID, serta liveness/readiness milik service.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A sudah diimplementasikan. Berkas tersedia: `health.go`. Cakupan pengujian ada di [bukti ingest](../../../../docs/evidence/ingest/README.md). Tabel rencana di bawah adalah panduan pemecahan tanggung jawab; sebagian operasi digabung dalam file yang tersedia.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `logger.go` | Konfigurasi slog JSON dan redaksi field sensitif. |
| `correlation.go` | Validasi/bangkitkan ID dan simpan pada context. |
| `health.go` | Handler /health dan /ready serta ringkasan dependensi. |

## Kontrak dan alur

- /health menunjukkan proses hidup; /ready/ingest memeriksa database. /ready tetap 503 hingga query B tersedia.
- Field log minimum: service, correlation_id, operation, latency_ms, dan hasil.

## Dependensi

- Standard library; dipakai adapter, application, dan worker melalui dependensi yang sesuai.

## Aturan penting

- Jangan log Authorization, X-*-Key, token, password, atau payload raw sebelum proyeksi.
- ID diteruskan lewat HTTP/Kafka dalam alur yang sama; ukur latensi outbound.
- Readiness ingest tetap 200 ketika sumber mati selama database tersedia; status sumber tersimpan terpisah. Kafka belum dipakai jalur A.
- Untuk service tanpa package ini, utilitas lokal ditempatkan pada http/consumer; tidak membuat shared module bisnis.

## Langkah implementasi dan verifikasi

- Sepakati format log lintas service.
- Pastikan trace script dapat mengikuti satu alur ingest sampai consumer.
