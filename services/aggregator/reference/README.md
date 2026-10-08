# reference

[Panduan service](../README.md) · [Peta repository](../../../README.md)

Referensi statis nama serta koordinat gunung api milik BNPB.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Load() membaca JSON embedded menjadi map berdasarkan volcano_id; MapVolcanic memeriksa keberadaan ID pada map tersebut.
- Daftar ID untuk fixture disepakati dengan pembuat mock melalui dokumentasi.

## Dependensi

- Dipakai canonicalize/ingest melalui data yang dirangkai main; bukan import kode mock.

## Aturan penting

- Unknown volcano dikarantina, bukan dipetakan ke koordinat null.
- Referensi mock harus diberi label sintetis jika bukan data aktual tervalidasi.
- Tidak membaca filesystem atau database mock lintas service.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [embed.go](embed.go) | Embed dan decode referensi JSON menjadi map volcano_id. Tidak melakukan validasi geografis tambahan. |
| [volcanoes.json](volcanoes.json) | Referensi sintetis nama dan koordinat gunung api yang dikenali Aggregator. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
