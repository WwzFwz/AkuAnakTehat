# Verifikasi jalur 3C — 1 Oktober 2026

Jalur mock → ingest → outbox PostgreSQL → Kafka → consumer SQLite sudah berjalan pada Windows/Docker Desktop lokal. Notifier mengirim simulasi melalui log. Semua data pengujian sintetis.

| Pemeriksaan | Hasil |
| --- | --- |
| Generator secret, test/vet delapan module Go | Lulus; [output](modules.txt). Tes DB memerlukan container terpisah di bawah. |
| Build image Aggregator dan tiga consumer, startup Compose | Lulus; [keadaan akhir](services.txt). |
| PostgreSQL `TestIngestPostgres`, termasuk `outbox_ACK_and_cleanup` | Lulus (1,58s); schema sementara, pending urut id, mark idempoten, cleanup tidak menghapus unpublished. |
| `TestEventPipeline` | Lulus (68,18s); [output](runtime.txt). |
| Regresi fondasi dan ingest | Lulus (66,61s total); [output](regression.txt). |

Uji unit memeriksa decoder additive, ketepatan angka besar, replay/versi lama, SQLite setelah dibuka ulang, filter alert dan dedup notifier, retry, serta larangan commit saat DLQ gagal. Relay diuji dengan kegagalan publish dan mark, termasuk replay event_id yang sama sebelum row berikutnya.

Uji live memeriksa:

1. Snapshot sumber diterima dashboard melalui outbox dan Kafka.
2. Pemda bergabung memakai group sendiri, membaca histori, tanpa mengganti container Aggregator.
3. Replay dan event versi lama tidak menimpa view terbaru; field tambahan dan angka `9007199254740993` tetap utuh.
4. Marker notifier tetap sama setelah replay/restart; NORMAL tidak mengirim alert.
5. Payload invalid masuk DLQ ketiga group dengan konteks asal; record valid berikutnya tetap diproses.
6. Dashboard dihentikan sementara; notifier/pemda tetap maju; dashboard mengejar event setelah hidup kembali.
7. Kafka dihentikan; checkpoint ingest tetap maju dan outbox pending bertambah. Setelah broker pulih, seluruh row sampai batas yang direkam mendapatkan ACK.

Perbaikan yang ditemukan saat validasi: URI SQLite Windows perlu awalan slash pada drive; port loopback Docker memerlukan network consumer tambahan selain bus_net internal. Fixture JWT lama memakai expiry hanya satu detik di masa lalu dan sempat gagal pada lingkungan host/VM ini; expiry dibuat jelas di masa lalu dan subtest memberi nama claim yang gagal. Kode validasi JWT tidak diubah.

## Mengulang

Dari root, setelah bootstrap secret lokal:

```powershell
$env:GOTELEMETRY='off'
$env:GOTOOLCHAIN='local'
$env:GOWORK='off'
docker compose up -d --build --wait --wait-timeout 180
docker compose --profile demo build pemda-portal
docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test
go test ./scripts/check/foundation_test.go ./scripts/check/events_test.go -run TestEventPipeline -v -count=1 -timeout=10m
go test ./scripts/check/foundation_test.go ./scripts/check/ingest_test.go -v -count=1 -timeout=10m
```

POSIX: `make check`, `make events-check`, `make ingest-check`, `make smoke`. Jalankan suite live bergantian: mereka mengubah simulasi mock dan menghentikan service sementara. Uji event meninggalkan record sintetis, pemda aktif, dan volume persisten; tidak menghapus data lama.

Belum dibuktikan: hosted GitHub Actions, load/race test, kehilangan disk broker, banyak replica per consumer/relay, crash nyata tepat di celah kirim/marker, serta demo P1–P5 lengkap. Kegagalan ACK/mark dan DLQ tanpa commit diuji pada batas interface; tidak diklaim sebagai exactly-once. Query/client-api lengkap (3B) masih belum tersedia.
