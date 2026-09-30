# Envelope event kanonik — kontrak baseline

**Status:** kontrak integrasi untuk jalur inti berikutnya. Fondasi menyiapkan topic Kafka; belum ada producer outbox atau consumer bisnis yang berjalan.

Topic: `bnpb.hazard-events.v1`. Key: `hazard_id`. Satu partisi, satu broker, replication factor 1, dan retensi tujuh hari pada baseline.

```json
{
  "schema_version": 1,
  "event_id": "0199a100-0000-7000-8000-000000000001",
  "event_type": "hazard.upserted",
  "hazard_id": "0199a100-0000-7000-8000-000000000002",
  "version": 1,
  "correlation_id": "demo-example",
  "published_at": "2026-09-30T01:00:01Z",
  "hazard": {
    "hazard_id": "0199a100-0000-7000-8000-000000000002",
    "source": "BMKG",
    "source_ref_id": "DEMO-E001",
    "hazard_type": "SEISMIC",
    "severity": "SIAGA",
    "area_name": "Wilayah Demo",
    "latitude": -7.5,
    "longitude": 110.4,
    "occurred_at": "2026-09-30T01:00:00Z",
    "ingested_at": "2026-09-30T01:00:01Z",
    "attributes": {"magnitude": 6.7, "depth_km": 10, "potential_tsunami": false}
  }
}
```

Contoh di atas sintetis, bukan bukti kejadian atau hasil pengujian.

## Semantik

- `event_id` mengidentifikasi perubahan yang dicatat dalam outbox; tidak berubah saat retry/replay.
- `version` mulai dari 1 dan meningkat hanya ketika isi bisnis berubah. `hazard_id` pada envelope dan payload harus sama.
- `published_at` pada envelope adalah waktu producer menyiapkan payload publikasi, bukan waktu ACK broker. Pertahankan nilainya bersama payload outbox saat retry. Kolom `outbox.published_at` mempunyai arti terpisah: waktu ACK dicatat oleh relay.
- `ingested_at` adalah penerimaan pertama record logis; pembaruan tidak menggantinya. `occurred_at` berasal dari sumber.
- `attributes` mempertahankan nilai JSON asli; unknown fields dibaca secara toleran. Perubahan destruktif memerlukan kontrak/topic versi lain.
- Ringkasan publik: hazard_id, source, hazard_type, severity, area_name, occurred_at, ingested_at. source_ref_id, latitude, longitude, dan attributes adalah Mentah.
- Kafka ini untuk consumer internal. Field mentah tidak disajikan ke Media melalui topic ini.

## Delivery dan consumer

Delivery yang dituju at-least-once. Dashboard, notifier, dan pemda memakai group berbeda. Efek persisten diselesaikan sebelum commit offset. Dashboard/pemda menerapkan hanya versi yang lebih baru. Notifier masih dapat mengirim duplikat jika crash setelah kirim sebelum pencatatan dedup.

Topic DLQ: `bnpb.hazard-events.v1.dlq`. Pesan memuat payload asli dan konteks group asal, alasan gagal, jumlah percobaan, topic/partition/offset asal, serta correlation ID. Offset asal baru diselesaikan sesudah publish DLQ di-ACK. DLQ bukan bukti efek bisnis telah berhasil.

Menambah partisi dapat mengubah pemetaan key; strategi transisi ordering berada di luar baseline M1.
