# P2: Konkurensi dan availability

`TestDynamicSchemaAndStaleAPI` memeriksa stale volcanic, seismic yang tetap sehat, dan recovery tanpa restart. Runner k6 menjalankan seismic/volcanic bersamaan dengan delay PVMBG 3s, sustained 50 VU selama 90s, serta outage/recovery. Jumlah koneksi TCP diukur terpisah; metrik lengkap dan denominator error ada di indeks integrasi.

Lihat [hasil dan batas pengujian](../integration/README.md),
[regresi runtime](../integration/regression.txt), dan
[kode pengujian](../../../scripts/check/README.md).

Bukti ini berasal dari stack lokal. Laporan akhir dan presentasi demo harus
merujuk hasil nyata beserta batasannya; hasil lokal tidak membuktikan deployment
multi-host, high availability, atau keberhasilan workflow GitHub Actions.
