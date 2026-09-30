# http

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Transport endpoint PVMBG dengan autentikasi dan perilaku simulasi.

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

- GET /volcanic-reports?since=..., POST /admin/schema-version, POST /admin/outage, GET /health.

## Dependensi

- store, auth, config, dan simulation.

## Aturan penting

- Delay data500 ms–3 s configurable; admin tetap responsif selama outage.
- Terapkan delay di luar lock store; request cancellation menghentikan penantian.
- Mode hang harus berhenti saat client memutus context. Health dapat melaporkan outage; bedakan dari proses sungguhan yang mati.

## Langkah implementasi dan verifikasi

- Implementasikan endpoint resmi lebih dahulu.
- Pastikan log punya correlation ID dan tidak membawa kredensial.
