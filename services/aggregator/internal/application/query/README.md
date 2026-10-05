# query

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Use case baca Canonical Store melalui API internal, beserta filter, halaman, dan status sumber.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** application query, port, repository, dan HTTP inbound sudah diimplementasikan. Wiring runtime tersedia pada composition root Aggregator.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `service.go` | List/get hazard melalui port baca. |
| `ports.go` | HazardQuery sebagai kebutuhan pembacaan application. |
| `filter.go` | Filter tipe/severity/since, limit, dan cursor opaque. |
| `result.go` | Page result dan metadata freshness per sumber. |

## Kontrak dan alur

- Operasi usulan: List(ctx, filter) dan Get(ctx, hazardID).
- Respons memuat HazardEvent tanpa metadata storage internal, next_cursor, dan metadata sumber sesuai kontrak HTTP.

## Dependensi

- domain/hazard; implementasi HazardQuery di adapter/outbound/postgres.

## Aturan penting

- Tidak menghubungi BMKG/PVMBG pada jalur request.
- limit default 100, maksimum 500; cursor mengacu pada occurred_at dan hazard_id.
- Hasil filter kosong tidak otomatis berarti sumber unavailable.
- Timeout query lokal wajib; data lama dapat disajikan dengan penanda stale.

## Langkah implementasi dan verifikasi

- DTO respons sudah mengikuti kontrak Aggregator HTTP dan client-api.
- Tentukan perilaku endpoint gabungan saat satu sumber belum punya data.
