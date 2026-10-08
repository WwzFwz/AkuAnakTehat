# query

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Use case baca Canonical Store melalui API internal, beserta filter, halaman, dan status sumber.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Operasi tersedia berupa List(ctx, filter) dan Get(ctx, hazardID).
- Respons memuat HazardEvent tanpa metadata storage internal, next_cursor, dan metadata sumber sesuai kontrak HTTP.

## Dependensi

- domain/hazard; implementasi HazardQuery di adapter/outbound/postgres.

## Aturan penting

- Tidak menghubungi BMKG/PVMBG pada jalur request.
- limit default 100, maksimum 500; cursor mengacu pada occurred_at dan hazard_id.
- Hasil filter kosong tidak otomatis berarti sumber unavailable.
- Timeout query lokal wajib; data lama dapat disajikan dengan penanda stale.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [filter.go](filter.go) | Normalisasi filter dan encoding serta validasi cursor keyset. |
| [filter_test.go](filter_test.go) | Pengujian `TestNormalizeFilter`, `TestNormalizeFilterRejectsInvalidValues`, `TestCursorRoundTrip`, `TestServiceNormalizesEmptyResults`. |
| [ports.go](ports.go) | HazardQuery sebagai kontrak pembacaan halaman dan detail. |
| [result.go](result.go) | Tipe halaman, detail hazard, dan metadata freshness sumber. |
| [service.go](service.go) | Use case List dan Get, validasi filter dan pemetaan kegagalan penyimpanan. |

## Perilaku dan batas saat ini

Get mengembalikan HazardDetail dengan sources pada level yang sama dengan field hazard. List menyertakan sources dan cursor; repository membatasi halaman berdasarkan jumlah maupun byte.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
