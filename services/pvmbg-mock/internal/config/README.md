# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal pvmbg-mock; nama variabel berikut adalah usulan yang perlu disepakati.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `config.go` | Struct konfigurasi dan pembacaan env/berkas lokal. |
| `validate.go` | Validasi invarian milik service sebelum worker dimulai. |

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port rencana 8082.
- PVMBG_TOKEN_HASH, ADMIN_KEY_HASH: Token baca dan admin terpisah.
- DELAY_MIN, DELAY_MAX: Default500 ms–3 s; configurable.
- GENERATION_INTERVAL: Maks10 s untuk laju minimum tugas.
- SEED_DIR: Seed v1 tanpa confidence_level.

## Dependensi

- Environment variable dan file lokal yang tidak ter-commit; diteruskan main ke komponen.

## Aturan penting

- Tidak membaca env/file konfigurasi service lain.
- Tidak memakai secret default dan tidak mencetak konfigurasi sensitif.
- Validasi durasi positif, limit, TTL, dan path yang relevan; konfigurasi antarservice dicatat sebagai kontrak deployment.

## Langkah implementasi dan verifikasi

- Finalisasi nama env di .env.example saat file itu dibuat.
- Pisahkan error konfigurasi dari dependensi yang sementara unavailable.
