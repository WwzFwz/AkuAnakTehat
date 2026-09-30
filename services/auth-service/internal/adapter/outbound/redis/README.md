# redis

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Persistensi refresh token dan status keluarga dengan TTL.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `rotate.lua`, `store.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `store.go` | Koneksi go-redis, timeout, dan operasi store. |
| `keys.go` | Namespace key token/family dan TTL. |
| `rotate.lua` | Rotasi token lama/pengganti dan keputusan reuse. |
| `scripts.go` | Embed serta pemanggilan script Lua. |

## Kontrak dan alur

- Memenuhi port RefreshStore application untuk create, rotate, dan family revocation.
- Operasi rotasi mengembalikan hasil terklasifikasi, bukan error string mentah ke client.

## Dependensi

- domain/application, go-redis; hanya auth-service menjangkau auth-store.

## Aturan penting

- Timeout lokal awal200 ms.
- Token used mempertahankan family_id; periksa status revoked sebelum menerbitkan pengganti.
- Validasi argumen/key type sebelum mutasi; atomisitas Lua tidak berarti rollback otomatis atas semua runtime error.
- AOF/volume dan timeout perlu dikonfigurasi; tidak mengklaim zero data loss untuk semua crash.

## Langkah implementasi dan verifikasi

- Sepakati bentuk key dan expiry sebelum menulis script.
- Verifikasi dua request refresh bersamaan, reuse, dan response-loss limitation.
