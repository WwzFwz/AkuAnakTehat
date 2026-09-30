# ingest

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Orkestrasi pemetaan, korelasi, deteksi perubahan, penulisan atomik, dan kemajuan polling.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A sudah diimplementasikan. Berkas tersedia: `ports.go`, `service.go`. Cakupan pengujian ada di [bukti ingest](../../../../../docs/evidence/ingest/README.md). Tabel rencana di bawah adalah panduan pemecahan tanggung jawab; sebagian operasi digabung dalam file yang tersedia.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `service.go` | Pipeline batch ingest per endpoint. |
| `ports.go` | UnitOfWork, Tx, checkpoint, dan port status sumber. |
| `batch.go` | Record valid, record ditolak, watermark kandidat, dan correlation ID. |
| `source_status.go` | Status sumber dan waktu polling sukses/gagal. |
| `event_envelope.go` | Payload hazard.upserted yang disimpan ke outbox. |

## Kontrak dan alur

- WithTx(ctx, callback) menyediakan repository hazard, warning, outbox, watermark, dan quarantine yang memakai transaksi sama.
- Operasi usulan: ApplyBatch(ctx, batch).
- Port checkpoint membaca watermark persisten; StatusStore mencatat hasil polling per sumber/endpoint.

## Dependensi

- canonicalize, domain/hazard, dan domain/tsunami.
- Implementasi port disediakan adapter/outbound/postgres melalui main.

## Aturan penting

- Upsert, outbox, dan watermark terkait harus satu transaksi; tidak ada repository Tx yang membuka transaksi sendiri.
- Hash sama tidak membuat event baru; hash berubah menaikkan version.
- Warning yang datang lebih dahulu disimpan; warning terlambat harus memperbarui hazard yang sudah ada.
- Jangan memajukan watermark sebelum seluruh batch ditangani, termasuk pencatatan record karantina.

## Langkah implementasi dan verifikasi

- Sepakati port UnitOfWork dengan A/B/C.
- Implementasikan pipeline inti dan log drift; schema_observations merupakan tambahan.
- Verifikasi replay, crash sebelum commit, dan korelasi ulang warning.
