# aggregator

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Klien HTTP untuk API internal Aggregator.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `client.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

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
- X-Internal-Key dan X-Correlation-ID dikirim; secret tidak dicatat.
- Retry koneksi maksimal sekali jika budget waktu masih cukup.
- Header propagasi deadline adalah tambahan; jangan menghilangkan timeout lokal jika fitur itu mati.

## Langkah implementasi dan verifikasi

- Implementasikan validasi respons, batas ukuran, penutupan body, dan pengukuran latensi.
- Verifikasi Aggregator unavailable tidak membuat request menggantung.
