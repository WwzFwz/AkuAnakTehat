# kafka

[Peta repository](../../README.md)

Konfigurasi broker KRaft dan inisialisasi topic untuk seluruh jalur event.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `create-topics.sh` | Menunggu broker siap lalu membuat topic event dan DLQ secara idempoten. |

## Kontrak dan alur

- Topic utama bnpb.hazard-events.v1: satu partisi, replication factor1, retention7 hari.
- Topic DLQ dibuat sebelum consumer memerlukannya.

## Dependensi

- Container Kafka dan kafka-init pada bus_net; konfigurasi runtime kelak di docker-compose.yml.

## Aturan penting

- Producer dan consumer memakai broker melalui jaringan; tidak berbagi library bisnis.
- Log Kafka memakai named volume.
- Jangan menganggap --if-not-exists memperbarui konfigurasi topic yang sudah ada.
- Menambah partisi membutuhkan strategi transisi untuk ordering per hazard.

## Langkah implementasi dan verifikasi

- Pin image saat implementasi dan konfigurasi listener internal.
- Uji topic init saat start baru dan start ulang.
