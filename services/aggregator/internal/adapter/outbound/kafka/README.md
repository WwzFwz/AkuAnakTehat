# kafka

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Producer ke topic event kanonik; tidak mengetahui daftar atau alamat consumer.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `producer.go` | Konfigurasi franz-go dan lifecycle producer. |
| `publish.go` | Publish key/payload/header dan menunggu ACK. |

## Kontrak dan alur

- Implementasi Publisher.Publish(ctx, message) milik worker/outbox.
- Topic bnpb.hazard-events.v1; key hazard_id; correlation ID dibawa pada header.

## Dependensi

- worker/outbox; franz-go direncanakan saat implementasi.

## Aturan penting

- Producer idempoten, acks=all, timeout awal 5 detik.
- Sukses hanya setelah ACK; event_id tetap saat replay outbox.
- Tidak menandai row published; itu tanggung jawab relay melalui store.
- Satu broker/RF1 tidak menjamin selamat dari kehilangan disk broker.

## Langkah implementasi dan verifikasi

- Implementasikan publisher dengan timeout dan log latensi.
- Verifikasi broker unavailable tidak menghilangkan row pending.
