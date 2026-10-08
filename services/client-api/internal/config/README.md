# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal client-api; nama variabel mengikuti config.go dan .env.example.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port default 8080.
- AGGREGATOR_URL, INTERNAL_KEY: Akses API internal; tanpa DATABASE_URL.
- JWT_PUBLIC_KEY_FILE, JWT_ISSUER, JWT_AUDIENCE: Verifikasi lokal.
- AGGREGATOR_TIMEOUT: Default1,5 s.
- MAX_CONCURRENT, RATE_LIMIT_RPS, RATE_LIMIT_BURST: Proteksi trafik baseline.
- PAGE_DEFAULT, PAGE_MAX: Default100/maks 500.
- Cache dan forwarding deadline belum mempunyai implementasi atau sakelar konfigurasi aktif.

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
