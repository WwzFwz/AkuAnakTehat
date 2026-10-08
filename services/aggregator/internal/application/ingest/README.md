# ingest

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Orkestrasi pemetaan, korelasi, deteksi perubahan, penulisan atomik, dan kemajuan polling.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- WithTx(ctx, callback) menyediakan repository hazard, warning, outbox, watermark, dan quarantine yang memakai transaksi sama.
- Operasi tersedia berupa ApplyBatch(ctx, batch).
- Port checkpoint membaca watermark persisten; StatusStore mencatat hasil polling per sumber/endpoint.

## Dependensi

- canonicalize, domain/hazard, dan domain/tsunami.
- Implementasi port disediakan adapter/outbound/postgres melalui main.

## Aturan penting

- Upsert dan outbox atomik per record. Checkpoint disimpan dalam transaksi terpisah setelah seluruh respons ditangani; prefix yang sudah commit aman dipoll ulang.
- Hash sama tidak membuat event baru; hash berubah menaikkan version.
- Warning yang datang lebih dahulu disimpan; warning terlambat harus memperbarui hazard yang sudah ada.
- Jangan memajukan watermark sebelum seluruh batch ditangani, termasuk pencatatan record karantina.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [ports.go](ports.go) | UnitOfWork, Tx, record outbox dan karantina, serta port checkpoint dan status polling. |
| [service.go](service.go) | ApplyBatch, transaksi per record, korelasi, hash, envelope outbox, karantina, dan checkpoint akhir respons. |

## Perilaku dan batas saat ini

Batch dan item input berada di `../canonicalize/input.go`. Jika envelope melebihi 4 MiB, transaksi record dibatalkan dan payload sumber dikarantina. Checkpoint tetap tidak maju ketika ada kegagalan yang belum tersimpan; hash mencegah outbox ganda ketika prefix diulang.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
