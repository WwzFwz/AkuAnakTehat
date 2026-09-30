# consumer

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Loop konsumsi Kafka independen untuk notifier.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `consumer.go` | Koneksi franz-go, subscription, rebalance, dan shutdown. |
| `processor.go` | Decode lalu panggil application secara berurutan per partisi. |
| `retry.go` | Retry terbatas dan jalur DLQ setelah kegagalan berulang. |
| `logging.go` | Log event/offset/correlation ID dan latensi outbound. |

## Kontrak dan alur

- Run(ctx) berlangganan topic bnpb.hazard-events.v1.
- Port processor: Handle(ctx, event); commit manual mengikuti hasil pemrosesan.
- Group yang berbeda wajib dipakai dashboard, notifier, dan pemda.

## Dependensi

- contract, application, dan publisher DLQ lokal. Tidak menghubungi Canonical Store.

## Aturan penting

- Offset selesai hanya sesudah efek bisnis persisten atau setelah ACK publish DLQ.
- Jangan commit melewati pesan sebelumnya yang belum selesai.
- Earliest berlaku untuk group tanpa offset valid, bukan replay otomatis setiap restart.
- Rebalance/shutdown tidak boleh meng-commit pekerjaan yang belum selesai.

## Langkah implementasi dan verifikasi

- Implementasikan konsumsi satu partisi dahulu.
- Verifikasi stop/start consumer, backlog catch-up, dan duplikat setelah crash.
- Untuk dashboard/pemda, publisher DLQ dapat berupa file lokal di package ini; tidak meng-import kode notifier.
