# http

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Transport API internal Aggregator: routing, autentikasi internal, validasi input, dan serialisasi hasil query.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `router.go` | Route /internal/hazards dan detail berdasarkan ID. |
| `handler.go` | Menerjemahkan HTTP ke application/query. |
| `auth.go` | Validasi X-Internal-Key. |
| `cursor.go` | Encode/decode cursor keyset. |
| `errors.go` | Pemetaan error application ke status HTTP. |

## Kontrak dan alur

- GET /internal/hazards menerima filter yang disepakati di docs/api.
- GET /internal/hazards/{id} membaca satu hazard.
- Header wajib bisnis: X-Internal-Key; X-Correlation-ID diteruskan.

## Dependensi

- application/query, config, dan observability lokal.
- Tidak membuka koneksi PostgreSQL sendiri.

## Aturan penting

- Validasi enum, timestamp, limit, dan cursor sebelum query.
- Timeout lokal aktif sejak baseline. X-Request-Deadline hanya rencana tambahan dan harus dibatasi server.
- Port internal tidak dipublikasikan ke host.
- Tidak mengembalikan detail error SQL atau secret.

## Langkah implementasi dan verifikasi

- Finalisasi kontrak HTTP dengan B.
- Hubungkan use case, auth internal, timeout, dan handler health/readiness.
