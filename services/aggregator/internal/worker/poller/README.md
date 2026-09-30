# poller

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Penjadwal polling independen per sumber, dengan satu jalur penulisan per sumber.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A sudah diimplementasikan. Berkas tersedia: `breaker.go`, `poller.go`, `poller_test.go`. Cakupan pengujian ada di [bukti ingest](../../../../../docs/evidence/ingest/README.md). Tabel rencana di bawah adalah panduan pemecahan tanggung jawab; sebagian operasi digabung dalam file yang tersedia.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `poller.go` | Lifecycle, interval/jitter, cancellation, dan pencegahan siklus overlap. |
| `bmkg.go` | Fetch dua endpoint paralel lalu proses hasil secara berurutan. |
| `pvmbg.go` | Fetch laporan vulkanik dan kirim batch ke ingest. |
| `breaker.go` | Circuit breaker dan pemulihan tanpa restart. |

## Kontrak dan alur

- Run(ctx) mengelola loop; fetcher dan ingest diberikan sebagai dependensi.
- Checkpoint dibaca dari storage milik Aggregator melalui port application.

## Dependensi

- Port ingest dan HTTP client sumber yang dirangkai oleh main.

## Aturan penting

- Tidak ada SQL, mapping severity, atau publish Kafka di scheduler.
- Default interval BMKG2 s/PVMBG5 s, overlap10 s; semuanya configurable.
- Watermark warning memakai waktu mulai poll sukses; mock memakai waktu perubahan. Asumsi jam bersama dicatat.
- Hasil endpoint BMKG yang sukses tetap diproses jika endpoint lain gagal.
- Poller tidak menumpuk goroutine saat sumber hang.

## Langkah implementasi dan verifikasi

- Implementasikan timeout, breaker, status stale, dan cancellation sejak baseline.
- Verifikasi outage PVMBG tidak menghentikan ingest BMKG.
