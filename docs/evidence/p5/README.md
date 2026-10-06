# P5: Pub/sub dan consumer independen

`TestEventPipeline` memeriksa outbox ACK, replay/dedup persisten, DLQ, stop/catch-up dashboard sementara consumer lain berjalan, dan recovery broker. Pengujian subscriber baru memakai group unik serta store kosong, kemudian memulihkan group/volume pemda normal. Aggregator tidak diganti saat subscriber masuk.

Lihat [hasil dan batas pengujian](../integration/README.md),
[regresi runtime](../integration/regression.txt), dan
[kode pengujian](../../../scripts/check/README.md).
Untuk P5, penguatan uji subscriber baru tercatat di
[event replay](../integration/events-fresh-group.txt).

Bukti ini berasal dari stack lokal. Laporan akhir dan presentasi demo harus
merujuk hasil nyata beserta batasannya; hasil lokal tidak membuktikan deployment
multi-host, high availability, atau keberhasilan workflow GitHub Actions.
