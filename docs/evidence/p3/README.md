# P3: Identitas, scope, dan refresh token

`TestQueryIntegration` memakai kredensial Media, Tim Lapangan, dan BNPB yang berbeda; Media mendapat ringkasan dan ditolak saat meminta raw. `TestNaturalExpiryAndFieldCLI` melewati TTL alami 60s, menjalankan dua request CLI berjarak 65s dalam satu sesi, dan membuktikan token kedaluwarsa tetap ditolak. Suite fondasi menguji signature/claim serta reuse refresh token.

Lihat [hasil dan batas pengujian](../integration/README.md),
[regresi runtime](../integration/regression.txt), dan
[kode pengujian](../../../scripts/check/README.md).

Bukti ini berasal dari stack lokal. Laporan akhir dan presentasi demo harus
merujuk hasil nyata beserta batasannya; hasil lokal tidak membuktikan deployment
multi-host, high availability, atau keberhasilan workflow GitHub Actions.
