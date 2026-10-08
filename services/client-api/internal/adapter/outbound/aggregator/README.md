# aggregator

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Klien HTTP untuk API internal Aggregator.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Memenuhi port Aggregator milik application: list/get.

## Dependensi

- application untuk DTO/port dan net/http; tidak meng-import service lain.

## Aturan penting

- Timeout lokal maksimal 1,5 s untuk keseluruhan operasi, termasuk retry.
- X-Internal-Key dan X-Correlation-ID dikirim; secret tidak dicatat. Setiap percobaan HTTP menghasilkan satu log `upstream_request` dengan correlation ID yang sama, nomor `attempt`, `latency_ms`, status HTTP, dan `result`.
- Decode memakai `UseNumber` agar angka raw besar tidak dibulatkan; respons dengan trailing JSON ditolak.
- Retry koneksi maksimal sekali jika budget waktu masih cukup.
- Header propagasi deadline adalah tambahan; jangan menghilangkan timeout lokal jika fitur itu mati.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [client.go](client.go) | HTTP list/get/readiness, retry terbatas, batas respons 8 MiB, klasifikasi error, dan log per percobaan HTTP. |
| [client_test.go](client_test.go) | Pengujian `TestClassifyUpstreamErrors`, `TestClientListValidatesResponse`, `TestRawNumbersSurviveAndTrailingJSONFails`. |
| [attempt_test.go](attempt_test.go) | Retry gagal lalu berhasil, correlation ID dan deadline bersama, sanitasi log, batas retry, timeout, dan penutupan body. |

## Perilaku dan batas saat ini

Latency diukur ulang pada setiap pemanggilan `HTTP.Do`, termasuk pembacaan, decode, dan penutupan body respons pada percobaan tersebut. Status bernilai 0 jika respons HTTP tidak diperoleh. Hasil dicatat sebagai kategori tetap seperti `success`, `transport_error`, `timeout`, `canceled`, `http_error`, `body_read_error`, `response_too_large`, atau `invalid_json`; URL, kredensial, isi respons, dan teks galat upstream tidak dicetak.

Retry hanya dilakukan untuk kegagalan transport, maksimal satu kali, dengan deadline yang sama. Percobaan di sini adalah pemanggilan `HTTP.Do` pada adapter, bukan setiap paket jaringan atau retry koneksi internal library HTTP. Bukti pengujian tersedia pada [verifikasi U7](../../../../../../docs/evidence/http-attempts-2026-10-08/README.md).

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
