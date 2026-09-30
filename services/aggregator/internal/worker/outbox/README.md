# outbox

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Relay polling event tersimpan dan housekeeping baris terkirim; tetap di dalam Aggregator.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `ports.go` | OutboxStore dan Publisher. |
| `relay.go` | Ambil pending, publish berurutan, lalu tandai published. |
| `housekeeping.go` | Hapus baris published yang melewati retensi lokal. |

## Kontrak dan alur

- OutboxStore: Pending(ctx, limit), MarkPublished(ctx, id, time), DeletePublishedBefore(ctx, cutoff).
- Publisher: Publish(ctx, message), sukses berarti broker telah ACK.
- Run(ctx) memakai polling default 1 s; wake-up NOTIFY adalah tambahan.

## Dependensi

- Port lokal worker; adapter postgres dan Kafka diberikan dari main.

## Aturan penting

- Jangan melewati kegagalan publish lalu menerbitkan versi berikutnya dari hazard yang sama.
- MarkPublished hanya setelah ACK; crash pada celah ACK/mark dapat menyebabkan duplikat.
- Housekeeping hanya baris yang sudah published, default lebih tua dari24 jam.
- Payload pending menyimpan event_id stabil; tidak membuat ID baru setiap retry.

## Langkah implementasi dan verifikasi

- Implementasikan jalur polling dahulu.
- Verifikasi Kafka mati lalu pulih dan crash setelah ACK sebelum mark.
