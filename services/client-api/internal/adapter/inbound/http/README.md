# http

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Endpoint publik untuk Media, Tim Lapangan, dan BNPB Ops.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `router.go` | Route list seismik/vulkanik/gabungan, detail, serta permintaan raw. |
| `handler.go` | Parse request, panggil use case, serialisasikan respons. |
| `cursor.go` | Validasi representasi cursor tanpa mengakses DB. |
| `errors.go` | Kontrak error dan status HTTP. |

## Kontrak dan alur

- GET /v1/hazards, /v1/hazards/seismic, /v1/hazards/volcanic, /v1/hazards/{id}.
- Bentuk permintaan field mentah disepakati di docs/api sebelum coding.

## Dependensi

- authn, middleware, application, config, observability.

## Aturan penting

- 401 token invalid;403 insufficient_scope;429 overload;503 dependency/source unavailable.
- Jangan mengembalikan data sebelum autentikasi, otorisasi, dan proyeksi.
- Cursor/pagination diteruskan sesuai kontrak Aggregator; jangan membuat urutan halaman berbeda.

## Langkah implementasi dan verifikasi

- Finalisasi envelope respons dan status.
- Verifikasi tiga identitas berbeda lewat request HTTP nyata.
