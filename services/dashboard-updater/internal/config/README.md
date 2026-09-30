# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi lokal dashboard-updater; nama variabel berikut adalah usulan yang perlu disepakati.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `config.go` | Struct konfigurasi dan pembacaan env/berkas lokal. |
| `validate.go` | Validasi invarian milik service sebelum worker dimulai. |

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port rencana 8091.
- KAFKA_BROKERS, HAZARD_TOPIC: Stream event kanonik.
- CONSUMER_GROUP: Group independen dashboard.
- SQLITE_PATH: File pada volume milik dashboard.
- START_OFFSET, MAX_ATTEMPTS, DLQ_TOPIC: Earliest untuk group baru; retry/commit policy.

## Dependensi

- Environment variable dan file lokal yang tidak ter-commit; diteruskan main ke komponen.

## Aturan penting

- Tidak membaca env/file konfigurasi service lain.
- Tidak memakai secret default dan tidak mencetak konfigurasi sensitif.
- Validasi durasi positif, limit, TTL, dan path yang relevan; konfigurasi antarservice dicatat sebagai kontrak deployment.

## Langkah implementasi dan verifikasi

- Finalisasi nama env di .env.example saat file itu dibuat.
- Pisahkan error konfigurasi dari dependensi yang sementara unavailable.
