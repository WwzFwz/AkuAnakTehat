# poller

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Penjadwal polling independen per sumber, dengan satu jalur penulisan per sumber.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

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

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [breaker.go](breaker.go) | State circuit breaker untuk kegagalan beruntun, cooldown, dan probe pemulihan. |
| [poller.go](poller.go) | Polling per sumber, fetch endpoint, penyerahan batch ke ingest, dan pencatatan hasil polling. |
| [poller_test.go](poller_test.go) | Pengujian `TestPartialBMKGAndBreakerRecovery`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
