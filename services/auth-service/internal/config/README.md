# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal auth-service; nama variabel berikut adalah usulan yang perlu disepakati.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `config.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `config.go` | Struct konfigurasi dan pembacaan env/berkas lokal. |
| `validate.go` | Validasi invarian milik service sebelum worker dimulai. |

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port rencana 8090.
- JWT_PRIVATE_KEY_FILE, JWT_ISSUER, JWT_AUDIENCE: Kunci penandatangan dan claims.
- CLIENTS_FILE: Tiga client dengan hash secret dan scope berbeda.
- ACCESS_TOKEN_TTL, REFRESH_TOKEN_TTL: Default60 s/8 jam.
- REDIS_ADDR, REDIS_PASSWORD, REDIS_TIMEOUT: Akses auth-store, timeout200 ms.
- TOKEN_RATE_LIMIT: Pembatasan endpoint token.

## Dependensi

- Environment variable dan file lokal yang tidak ter-commit; diteruskan main ke komponen.

## Aturan penting

- Tidak membaca env/file konfigurasi service lain.
- Tidak memakai secret default dan tidak mencetak konfigurasi sensitif.
- Validasi durasi positif, limit, TTL, dan path yang relevan; konfigurasi antarservice dicatat sebagai kontrak deployment.

## Langkah implementasi dan verifikasi

- Finalisasi nama env di .env.example saat file itu dibuat.
- Pisahkan error konfigurasi dari dependensi yang sementara unavailable.
