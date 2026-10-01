# kafka

[Panduan service](../../../../README.md) · [Envelope](../../../../../../docs/api/hazard-event.md)

**Pemilik:** C. **Status:** producer.go mengimplementasikan worker/outbox.Publisher dengan franz-go 1.18.1.

Producer memakai idempotency bawaan, acks=all, key hazard_id, dan header event_id/correlation_id. KAFKA_PUBLISH_TIMEOUT default 5s membatasi penantian lokal. Bila ACK belum pasti saat deadline, satu publikasi in-flight tetap dilacak; relay menunggu hasil record yang sama pada siklus berikutnya sebelum mengirim versi berikutnya. Tidak menumpuk publikasi baru setiap timeout.

Sukses berarti callback ACK diterima. Replay setelah restart/mark gagal tetap mungkin duplikat sehingga consumer harus idempoten. Producer tidak mengetahui alamat/daftar consumer. Kafka satu broker RF1 bukan jaminan ketahanan terhadap kehilangan disk broker.
