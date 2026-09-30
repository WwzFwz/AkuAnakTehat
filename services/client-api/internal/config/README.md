# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal client-api; nama variabel berikut adalah usulan yang perlu disepakati.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Uji integrasi runtime belum dilakukan. Berkas yang sudah ada: `config.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `config.go` | Struct konfigurasi dan pembacaan env/berkas lokal. |
| `validate.go` | Validasi invarian milik service sebelum worker dimulai. |

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port rencana 8080.
- AGGREGATOR_URL, INTERNAL_KEY: Akses API internal; tanpa DATABASE_URL.
- JWT_PUBLIC_KEY_FILE, JWT_ISSUER, JWT_AUDIENCE: Verifikasi lokal.
- AGGREGATOR_TIMEOUT: Default1,5 s.
- MAX_CONCURRENT, RATE_LIMIT_RPS, RATE_LIMIT_BURST: Proteksi trafik baseline.
- PAGE_DEFAULT, PAGE_MAX: Default100/maks 500.
- READ_CACHE_ENABLED, DEADLINE_FORWARDING_ENABLED: Tambahan, default false.

## Dependensi

- Environment variable dan file lokal yang tidak ter-commit; diteruskan main ke komponen.

## Aturan penting

- Tidak membaca env/file konfigurasi service lain.
- Tidak memakai secret default dan tidak mencetak konfigurasi sensitif.
- Validasi durasi positif, limit, TTL, dan path yang relevan; konfigurasi antarservice dicatat sebagai kontrak deployment.

## Langkah implementasi dan verifikasi

- Finalisasi nama env di .env.example saat file itu dibuat.
- Pisahkan error konfigurasi dari dependensi yang sementara unavailable.
