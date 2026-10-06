# aggregator

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Klien HTTP untuk API internal Aggregator.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** HTTP client Aggregator sudah diimplementasikan, termasuk timeout total, retry koneksi terbatas, validasi response, correlation ID, dan klasifikasi error. Bukti P1-P5 lengkap masih terpisah.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `client.go` | HTTP client, base URL, kredensial internal, dan pool koneksi. |
| `request.go` | Penyusunan query serta forwarding correlation ID. |
| `response.go` | Decode respons ke DTO client-api dan klasifikasi error. |

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

## Langkah implementasi dan verifikasi

- Validasi respons, batas ukuran, penutupan body, timeout total, dan klasifikasi error sudah diimplementasikan.
- Verifikasi Aggregator unavailable tidak membuat request menggantung.
