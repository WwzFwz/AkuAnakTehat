# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal auth-service; nama variabel mengikuti config.go dan .env.example.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port default 8090.
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

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [config.go](config.go) | Membaca environment dan memvalidasi konfigurasi sebelum service dijalankan. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
