# keys

[Peta repository](../../README.md)

Lokasi runtime pasangan kunci Ed25519 hasil generator; berkas PEM diabaikan Git.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** generator tersedia dan memvalidasi pasangan key pada pemanggilan berikutnya. Kunci bukan artefak repository.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `jwt-private.pem` | Ed25519 PKCS8 PRIVATE KEY PEM lokal, hanya auth-service. |
| `jwt-public.pem` | Ed25519 PKIX PUBLIC KEY PEM lokal untuk verifier client-api. |

## Kontrak dan alur

- Auth-service membaca private key; client-api hanya membaca public key.

## Dependensi

- scripts/secrets dan mount per service di Compose.

## Aturan penting

- Tidak menyimpan secret asli dalam Git. Aturan ignore tersedia sebelum generator dijalankan.
- Jangan memakai kunci dokumentasi/example sebagai kunci runtime.
- Rotasi public key statis memerlukan distribusi dan restart verifier sesuai batasan M1.

## Langkah implementasi dan verifikasi

- Generator memakai izin 0600 pada OS pendukung; Windows menggunakan ACL workspace.
- Ikuti [bootstrap secret](../../scripts/secrets/README.md); generator tidak merotasi key secara otomatis.
