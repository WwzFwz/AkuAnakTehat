# aggregator

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Klien HTTP untuk API internal Aggregator.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Memenuhi port Aggregator milik application: list/get.

## Dependensi

- application untuk DTO/port dan net/http; tidak meng-import service lain.

## Aturan penting

- Timeout lokal maksimal1,5 s sejak baseline.
- X-Internal-Key dan X-Correlation-ID dikirim; secret tidak dicatat. Log `upstream_request` memuat correlation ID dan latency panggilan.
- Decode memakai `UseNumber` agar angka raw besar tidak dibulatkan; respons dengan trailing JSON ditolak.
- Retry koneksi maksimal sekali jika budget waktu masih cukup.
- Header propagasi deadline adalah tambahan; jangan menghilangkan timeout lokal jika fitur itu mati.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [client.go](client.go) | HTTP list/get/readiness, retry terbatas, batas respons 8 MiB, klasifikasi error, dan log total operasi fetch. |
| [client_test.go](client_test.go) | Pengujian `TestClassifyUpstreamErrors`, `TestClientListValidatesResponse`, `TestRawNumbersSurviveAndTrailingJSONFails`. |

## Perilaku dan batas saat ini

Log upstream_request saat ini mengukur total fetch, termasuk satu retry transport dan decode bila terjadi. Latency tiap percobaan HTTP belum dipisahkan; ini dicatat sebagai celah U7 pada audit persyaratan, bukan dianggap sudah terpenuhi.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
