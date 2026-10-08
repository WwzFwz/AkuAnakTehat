# observability

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Logging terstruktur, correlation ID, serta liveness/readiness milik service.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- /health menunjukkan proses hidup; /ready menunjukkan kesiapan melayani fungsi service.
- Log request dan outbound memuat identitas service, correlation_id, serta latency_ms; nama operasi dan hasil mengikuti jenis log.

## Dependensi

- Dipakai komponen dalam module service ini; tidak dibagikan sebagai library bisnis lintas service.

## Aturan penting

- Jangan log Authorization, X-*-Key, token, password, atau payload raw sebelum proyeksi.
- Correlation ID diteruskan ke dependensi melalui adapter yang relevan.
- Handler health/readiness berada di adapter inbound HTTP; package ini menyediakan instrumentasi request.
- Untuk service tanpa package ini, utilitas lokal ditempatkan pada http/consumer; tidak membuat shared module bisnis.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [http.go](http.go) | Logger JSON, context correlation ID, dan middleware pencatatan request HTTP. |

## Perilaku dan batas saat ini

http.go menyediakan Logger, Middleware, dan ID. Adapter outbound Aggregator menggunakan ID request; detail batas instrumentasi retry tercatat pada README adapter tersebut.

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
