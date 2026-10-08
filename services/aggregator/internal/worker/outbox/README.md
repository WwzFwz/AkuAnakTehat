# outbox

[Panduan service](../../../README.md) · [Kontrak port](../../../../../docs/api/aggregator-ports.md)

Satu relay membaca maksimum 100 pending row urut id tiap 1s, publish sequential, lalu MarkPublished setelah ACK. Kegagalan sementara pada publish/mark menghentikan batch. Kegagalan ukuran permanen dicatat pada rejected_at/rejection_reason tanpa menandai published; snapshot berikutnya boleh diproses. Payload ditolak tetap disimpan untuk inspeksi dan pemulihan manual. Payload dan event_id dipertahankan saat replay. Crash antara ACK dan mark dapat menghasilkan duplikat.

Tidak ada transaksi DB selama publish jaringan. Pool relay terpisah (2 koneksi) membatasi perebutan koneksi ingest. Housekeeping tiap jam menghapus hanya row published yang ACK-nya lebih tua dari OUTBOX_RETENTION (default 24h). LISTEN/NOTIFY belum diimplementasikan dan tetap tambahan.

Deployment baseline satu Aggregator/relay. Tidak ada claim lock atau koordinasi multi-replica. Tes unit mensimulasikan kegagalan ACK dan mark; tes PostgreSQL menguji pending/mark/cleanup; `make events-check` menguji outage broker nyata.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [ports.go](ports.go) | Message, Store, Publisher, dan klasifikasi kegagalan permanen. |
| [relay.go](relay.go) | Polling outbox berurutan, publish, penandaan ACK atau penolakan permanen, dan retention cleanup. |
| [relay_test.go](relay_test.go) | Pengujian `TestRelayAcknowledgementBoundary`, `TestRejectionMustBeDurableBeforeRelayContinues`, `TestPermanentFailurePreservesEvidenceAndAllowsNextEvent`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
