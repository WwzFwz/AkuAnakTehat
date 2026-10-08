# canonicalize

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Pemetaan data sumber ke HazardEvent; pemilik tipe input yang digunakan adapter dan ingest.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Operasi tersedia berupa MapSeismic(input, warnings) dan MapVolcanic(input, volcano).
- Mapper menghasilkan field bisnis; identitas persisten, version, dan transaksi ditangani ingest.

## Dependensi

- domain/hazard dan domain/tsunami.
- ingest dan adapter sumber boleh meng-import canonicalize; canonicalize tidak meng-import keduanya.

## Aturan penting

- Tanpa HTTP, SQL, atau publish Kafka.
- Jika ada warning, severity mengikuti aturan warning; jangan otomatis memilih maksimum antara severity magnitude dan warning.
- volcano_id tidak dikenal menghasilkan error terklasifikasi untuk karantina.
- Unknown fields masuk attributes; susunan warning harus stabil agar tidak menghasilkan perubahan hash palsu.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [input.go](input.go) | Tipe input dan batch serta tolerant decoder yang mempertahankan field tambahan. |
| [mapper.go](mapper.go) | Pemetaan seismik dan vulkanik serta korelasi warning menjadi HazardEvent. |
| [mapper_test.go](mapper_test.go) | Pengujian `TestSeverityAndWarningRules`, `TestTolerantReaderAndSemanticHash`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
