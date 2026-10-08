# Consumer event — implementasi 3C

Semua consumer menerima [envelope v1](hazard-event.md), tanpa akses PostgreSQL Aggregator. Kontrak Go disalin lokal tiap module agar service dapat dibangun sendiri; perubahan harus diuji terhadap ketiga decoder.

| Service / group default | Port loopback | Volume | Efek bisnis |
| --- | --- | --- | --- |
| dashboard-updater | 8091 | dashboard-data | SQLite `/data/view.db`, versi terbaru per hazard |
| notifier | 8092 | notifier-data | SQLite `/data/processed.db`, dedup `(hazard_id,version)` |
| pemda-portal | 8093 | pemda-data | SQLite `/data/view.db`, subscriber tambahan melalui profile `demo` |

Satu instance untuk setiap volume. Ketiganya memakai `bus_net` dan `consumer_net` (bridge untuk publikasi port loopback pada Docker Desktop). Tidak ada koneksi ke `store_net` atau akses langsung Canonical Store. Endpoint berikut internal dan memuat data mentah; API Media tetap melalui client-api.

## HTTP

- `GET /health`: 200 bila proses hidup.
- `GET /ready`: 200 bila SQLite dan broker dapat diakses, 503 bila gagal; deadline total 2s. Ini bukan ukuran consumer lag atau jaminan semua event sudah diterapkan.
- Dashboard/pemda: `GET /view?limit=100&after=<hazard_id>` → `{ "data": [<envelope lengkap>], "next_cursor": "..." }`.
- Notifier: `GET /processed?limit=100&after=<event_id>` → `{ "data": [{"event_id":"...","hazard_id":"...","version":1,"alert":true,"processed_at":"..."}], "next_cursor":"..." }`.
- `limit` 1..200, default 100; cursor eksklusif dan hasil urut ID. Cursor kosong berarti halaman terakhir. Limit invalid → 400; storage unavailable → 503. Pagination merupakan live view, bukan snapshot konsisten lintas halaman.

`alert=true` menyatakan sender simulasi telah dipanggil. Sender menulis JSON log `notification_simulated`; tidak menghubungi kanal notifikasi eksternal.

## Penyelesaian offset

Auto-commit mati. Consumer membaca satu record, menahan rebalance, lalu menyelesaikan efek bisnis sebelum commit offset. Input invalid langsung ke DLQ. Kegagalan pemrosesan dicoba 3 kali, deadline 2s tiap percobaan, jeda 200ms. Setelah percobaan pemrosesan habis, worker berhenti tanpa DLQ maupun commit offset; Compose memulai ulang untuk mencoba record yang sama setelah dependensi pulih. Hanya input invalid masuk DLQ dengan batas penantian 5s; offset asal di-commit hanya setelah ACK DLQ. Commit memakai timeout 5s.

Kegagalan DLQ atau commit menghentikan worker/proses, dan Compose restart mengulang offset terakhir yang belum tersimpan. Tidak ada commit record berikutnya yang menutupi kegagalan record sebelumnya. Rebalance timeout 60s melebihi batas percobaan yang dikonfigurasi (maksimum 5 × 5s ditambah jeda, DLQ dan commit).

DLQ `bnpb.hazard-events.v1.dlq` menggunakan key dan bytes payload asli. Header: `consumer_group`, `failure_reason` (`invalid_event`), `attempts`, `source_topic`, `source_partition`, `source_offset`, `correlation_id`. Payload invalid tanpa correlation ID tetap memiliki header kosong. DLQ juga at-least-once; crash setelah ACK sebelum commit dapat menggandakan pesan DLQ. DLQ bukan bukti efek bisnis berhasil dan belum memiliki alat redrive otomatis.

Dashboard/pemda melakukan UPSERT hanya bila version masuk lebih tinggi. Unknown fields dan angka JSON disimpan utuh. Notifier memeriksa dedup, mengirim SIAGA/AWAS, lalu mencatat marker; crash pada celah kirim/marker dapat mengirim duplikat. NORMAL/WASPADA dicatat tanpa kirim.

## Konfigurasi dan pemulihan

`HTTP_ADDR`, `SQLITE_PATH`, `KAFKA_BROKERS`, `KAFKA_TOPIC`, `KAFKA_DLQ_TOPIC`, `KAFKA_GROUP_ID`, `PROCESS_TIMEOUT`, dan `MAX_ATTEMPTS` dibaca masing-masing service. Default topic sesuai kontrak di atas; broker `kafka:9092`. Timeout proses 1ms..5s dan percobaan 1..5. Group berbeda diperlukan agar ketiganya menerima seluruh event.

Group baru mulai dari earliest yang masih ada; group lama memakai committed offset. Retensi Kafka default 7 hari membatasi catch-up. Jangan menghapus volume SQLite sambil mempertahankan committed offset lalu menganggap view pulih otomatis: pemulihan memerlukan replay terencana/reset group dan data yang masih tersedia. Outbox published disimpan default 24h; outbox pending tidak dihapus housekeeping.

Jalankan `docker compose --profile demo up -d --build pemda-portal` untuk subscriber ketiga. Producer tidak berubah. `make events-check` menjalankan uji live termasuk restart/outage dan menyimpan data sintetis pada topic/view demo; jalankan saat tidak ada demo lain bersamaan.
