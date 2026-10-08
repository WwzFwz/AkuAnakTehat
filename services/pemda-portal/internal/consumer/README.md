# consumer

[Panduan service](../../README.md) Â· [Kontrak consumer](../../../../docs/api/consumers.md)

franz-go membaca satu record setiap poll dengan auto-commit dimatikan dan rebalance ditahan selama proses. Secara default efek bisnis dicoba maksimum tiga kali dengan timeout lokal 2 s per percobaan dan jeda 200 ms. MAX_ATTEMPTS serta PROCESS_TIMEOUT dapat dikonfigurasi. Input invalid langsung ke DLQ. Commit offset hanya sesudah efek persisten selesai atau DLQ di-ACK. Kegagalan DLQ/commit menghentikan proses sehingga restart mengulang offset; tidak melompati record gagal. Group baru membaca dari earliest yang masih tersedia. Group lama melanjutkan offset tersimpan.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [consumer.go](consumer.go) | Fetch Kafka, validasi kontrak, retry pemrosesan, DLQ, commit offset, dan probe broker. |
| [consumer_test.go](consumer_test.go) | Menguji batas ACK, kegagalan dependensi, DLQ, retry, dan batas ukuran envelope. |
| [health_test.go](health_test.go) | Menguji deadline probe Kafka dan pencegahan penumpukan probe yang belum selesai. |
| [recovery_test.go](recovery_test.go) | Menguji kegagalan pool SQLite sungguhan dan pemrosesan ulang tanpa kehilangan offset. |
| [trace.go](trace.go) | Mencatat latency protokol Kafka dan correlation ID untuk DLQ serta commit offset. |

## Perilaku dan batas saat ini

Jika event valid gagal diproses setelah retry, worker berhenti tanpa DLQ dan tanpa commit offset. Kebijakan restart Compose memulai ulang worker untuk mencoba offset yang sama. Event yang melanggar kontrak masuk DLQ; offset asal baru di-commit setelah ACK DLQ. trace.go mengukur operasi protokol Kafka serta publish DLQ dan commit offset. Pemulihan storage diuji dengan pool SQLite sungguhan pada recovery_test.go.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
