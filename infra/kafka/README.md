# kafka

[Peta repository](../../README.md)

Konfigurasi broker KRaft dan inisialisasi topic untuk seluruh jalur event.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** startup broker, init topic, dan pesan pada topic pengujian yang bertahan setelah restart telah diverifikasi. Lihat [hasil fondasi](../../docs/evidence/foundation/README.md). Producer/consumer bisnis belum dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `init-topics.sh` | Membuat topic event dan DLQ secara idempoten setelah healthcheck broker lulus. |

## Kontrak dan alur

- Topic utama bnpb.hazard-events.v1: satu partisi, replication factor1, retention7 hari.
- Topic DLQ bnpb.hazard-events.v1.dlq memakai satu partisi, replication factor 1, retention 7 hari.

## Dependensi

- Container Kafka dan kafka-init pada bus_net; konfigurasi runtime ada di docker-compose.yml.

## Aturan penting

- Producer dan consumer memakai broker melalui jaringan; tidak berbagi library bisnis.
- Log Kafka memakai named volume.
- Jangan menganggap --if-not-exists memperbarui konfigurasi topic yang sudah ada.
- Menambah partisi membutuhkan strategi transisi untuk ordering per hazard.

## Langkah implementasi dan verifikasi

- Image apache/kafka:3.9.1; listener broker internal kafka:9092.
- Uji topic init saat start baru dan start ulang.
