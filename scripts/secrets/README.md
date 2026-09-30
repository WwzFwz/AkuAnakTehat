# secrets

[Peta repository](../../README.md)

Bootstrap kredensial pengembangan per service, tanpa hardcoded secret.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `generate.sh` | Entrypoint shell untuk membuat env/kunci yang belum tersedia. |
| `generate.ps1` | Alternatif Windows bila tim memilih implementasi native PowerShell. |

## Kontrak dan alur

- Menghasilkan env/<service>.env, data client, dan pasangan Ed25519 di env/keys.
- Format key harus sesuai loader auth-service dan verifier client-api.

## Dependensi

- Sumber acak kriptografis dan tooling lokal yang dinyatakan di README saat implementasi.

## Aturan penting

- Sebelum menghasilkan secret, buat .gitignore untuk env/kunci; saat ini folder hanya berisi dokumentasi.
- Tidak merotasi/menimpa secret yang sudah ada tanpa tindakan eksplisit.
- Private key hanya auth-service; public key client-api; mock verifikator menerima hash kredensial.
- Jangan mencetak secret ke stdout, log, atau docs/evidence.

## Langkah implementasi dan verifikasi

- Sepakati nama variabel dan format berkas dengan pemilik service.
- Tulis generator idempoten serta .env.example dengan placeholder saja.
