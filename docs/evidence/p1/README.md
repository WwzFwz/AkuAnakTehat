# P1: Interoperabilitas dan evolusi skema

`TestIngestPipeline` memverifikasi ingest kedua sumber serta log schema drift. `TestDynamicSchemaAndStaleAPI` mengaktifkan schema v2, membaca confidence_level melalui API raw, dan memastikan record lama tetap terbaca. Mapping, hash, korelasi tsunami, tolerant reader, serta presisi angka diuji pada module Aggregator.

Lihat [hasil dan batas pengujian](../integration/README.md),
[regresi runtime](../integration/regression.txt), dan
[kode pengujian](../../../scripts/check/README.md).

Bukti ini berasal dari stack lokal. Laporan akhir dan presentasi demo harus
merujuk hasil nyata beserta batasannya; hasil lokal tidak membuktikan deployment
multi-host, high availability, atau keberhasilan workflow GitHub Actions.
