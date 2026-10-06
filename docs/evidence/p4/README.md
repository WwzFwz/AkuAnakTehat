# P4: Independensi service dan fleksibilitas data

`TestIndependentRebuild` menghentikan, membangun ulang, lalu menjalankan notifier; API tetap melayani baca dan timestamp start container lain tidak berubah. `TestDynamicSchemaAndStaleAPI` memverifikasi old/new JSONB tanpa migrasi atau penggantian container. Batas akses Canonical Store mengikuti network Compose dan kredensial per service.

Lihat [hasil dan batas pengujian](../integration/README.md),
[regresi runtime](../integration/regression.txt), dan
[kode pengujian](../../../scripts/check/README.md).

Bukti ini berasal dari stack lokal. Laporan akhir dan presentasi demo harus
merujuk hasil nyata beserta batasannya; hasil lokal tidak membuktikan deployment
multi-host, high availability, atau keberhasilan workflow GitHub Actions.
