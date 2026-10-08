# sourcehttp

Transport HTTP bersama di dalam module Aggregator, dipakai adapter BMKG dan PVMBG. Tidak dibagikan lintas service dan tidak memuat DTO atau aturan bisnis.

`client.go` membatasi timeout, dua koneksi per host, respons 8 MiB, menutup body, meneruskan correlation ID, dan menolak redirect agar credential sumber tidak berpindah host. `client_test.go` memeriksa header, since, deadline request hang, dan respons terlalu besar. Error transport disanitasi; payload dan credential tidak dicatat.

Adapter sumber memilih route/header kemudian memanggil decoder `canonicalize.Decode`. Decoder memisahkan record invalid dari record valid; JSON envelope rusak menggagalkan endpoint tanpa memajukan watermark.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [client.go](client.go) | Transport HTTP sumber, timeout, header, pembatasan body, penolakan redirect, dan latency. |
| [client_test.go](client_test.go) | Pengujian `TestDeadlineCredentialsAndResponseLimit`. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
