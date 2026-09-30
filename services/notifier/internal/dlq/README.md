# dlq

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Penerbitan pesan gagal dengan konteks asal consumer.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `message.go` | Payload asli serta group/reason/attempt/topic/partition/offset/correlation ID. |
| `publisher.go` | Publish DLQ dan tunggu ACK Kafka. |

## Kontrak dan alur

- Publish(ctx,failedMessage) mengembalikan sukses hanya setelah ACK.
- Topic rancangan: bnpb.hazard-events.v1.dlq.

## Dependensi

- Kafka client lokal; dipanggil consumer setelah retry terbatas.

## Aturan penting

- Commit offset asal dilakukan sesudah publish DLQ berhasil.
- DLQ dapat berisi duplikat jika crash setelah publish sebelum commit.
- Masuk DLQ berarti efek bisnis belum berhasil, bukan notifikasi terkirim.
- Alasan gagal tidak memuat kredensial.

## Langkah implementasi dan verifikasi

- Finalisasi metadata dan klasifikasi retryable/permanent.
- Verifikasi broker gagal saat publish DLQ tidak membuat consumer melompati event.
