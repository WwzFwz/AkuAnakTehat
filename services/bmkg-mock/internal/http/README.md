# http

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Transport endpoint BMKG dengan autentikasi dan perilaku simulasi.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `router.go` | Route data, health, dan admin bila tersedia. |
| `handler.go` | Parse since, ambil snapshot, dan tulis JSON. |
| `health.go` | Health/readiness dan status simulasi. |
| `logging.go` | Log JSON, correlation ID, dan durasi request. |

## Kontrak dan alur

- GET /seismic-events?since=..., GET /tsunami-warnings?since=..., GET /health.

## Dependensi

- store, auth, config.

## Aturan penting

- Delay tetap configurable dalam rentang50–150 ms.
- Terapkan delay di luar lock store; request cancellation menghentikan penantian.
- HTTP handler tidak melakukan pemetaan HazardEvent.

## Langkah implementasi dan verifikasi

- Implementasikan endpoint resmi lebih dahulu.
- Pastikan log punya correlation ID dan tidak membawa kredensial.
