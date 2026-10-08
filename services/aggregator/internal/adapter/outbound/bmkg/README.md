# bmkg

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

HTTP client dan tolerant decoder untuk BMKG; tidak berisi aturan pemetaan severity.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- FetchSeismic(ctx, since, corr) dan FetchWarnings(ctx, since, corr).
- Hasil berupa tipe input application/canonicalize atau domain/tsunami, disertai record yang perlu dikarantina.

## Dependensi

- net/http; application/canonicalize dan domain/tsunami.
- Kredensial dan client HTTP diberikan oleh composition root.

## Aturan penting

- X-BMKG-Key; timeout lokal awal 1 detik.
- Tidak meng-import kode mock atau membuang field tak dikenal.
- Teruskan correlation ID, ukur latensi, tutup body respons, dan batasi ukuran respons.
- Record invalid tidak membatalkan record valid lain; format respons rusak menjadi error endpoint.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [client.go](client.go) | Fetch dua endpoint BMKG melalui sourcehttp, lalu decode melalui canonicalize. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
