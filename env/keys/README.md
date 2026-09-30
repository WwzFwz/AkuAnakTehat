# keys

[Peta repository](../../README.md)

Lokasi runtime pasangan kunci Ed25519; saat ini hanya dokumentasi.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `jwt-private.pem` | Rencana artefak lokal hasil generator; jangan di-commit. |
| `jwt-public.pem` | Rencana artefak lokal untuk verifier client-api. |

## Kontrak dan alur

- Auth-service membaca private key; client-api hanya membaca public key.

## Dependensi

- scripts/secrets dan mount per service di Compose.

## Aturan penting

- Tidak menyimpan secret asli dalam Git. Tambahkan aturan ignore sebelum menjalankan generator.
- Jangan memakai kunci dokumentasi/example sebagai kunci runtime.
- Rotasi public key statis memerlukan distribusi dan restart verifier sesuai batasan M1.

## Langkah implementasi dan verifikasi

- Sepakati format PEM dan izin akses berkas.
- Buat kunci ketika tahap bootstrap implementasi dimulai.
