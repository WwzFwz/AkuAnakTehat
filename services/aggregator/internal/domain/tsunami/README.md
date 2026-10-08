# tsunami

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Model warning dan aturan korelasi ke gempa, tanpa akses jaringan atau storage.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Korelasi menggunakan related_event_id, bukan warning_id.
- Model mendukung beberapa warning untuk satu gempa.

## Dependensi

- Standard library; dipakai canonicalize, ingest, dan adapter BMKG.

## Aturan penting

- Warning bukan HazardEvent baru.
- Metadata internal mock tidak menjadi ketergantungan model ini.
- Aturan warning hanya diterapkan pada gempa yang sesuai kontrak potential_tsunami.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [warning.go](warning.go) | Model warning, field tambahan, dan fungsi peringkat tingkat ancaman. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
