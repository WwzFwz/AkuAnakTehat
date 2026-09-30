# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal aggregator; nama variabel berikut adalah usulan yang perlu disepakati.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `config.go` | Struct konfigurasi dan pembacaan env/berkas lokal. |
| `validate.go` | Validasi invarian milik service sebelum worker dimulai. |

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port rencana 9000.
- BMKG_URL, PVMBG_URL: Alamat sumber; secret sumber terpisah.
- DATABASE_URL: Akses Canonical Store khusus Aggregator.
- BMKG_POLL_INTERVAL, PVMBG_POLL_INTERVAL: Default2 s/5 s; overlap10 s.
- BMKG_TIMEOUT, PVMBG_TIMEOUT: Default1 s/4 s; timeout DB lokal.
- KAFKA_BROKERS, HAZARD_TOPIC: Producer outbox; polling1 s, publish timeout5 s.
- INTERNAL_KEY_HASH: Verifikator X-Internal-Key.
- OUTBOX_NOTIFY_ENABLED, SCHEMA_OBSERVATIONS_ENABLED: Tambahan, default false.

## Dependensi

- Environment variable dan file lokal yang tidak ter-commit; diteruskan main ke komponen.

## Aturan penting

- Tidak membaca env/file konfigurasi service lain.
- Tidak memakai secret default dan tidak mencetak konfigurasi sensitif.
- Validasi durasi positif, limit, TTL, dan path yang relevan; konfigurasi antarservice dicatat sebagai kontrak deployment.

## Langkah implementasi dan verifikasi

- Finalisasi nama env di .env.example saat file itu dibuat.
- Pisahkan error konfigurasi dari dependensi yang sementara unavailable.
