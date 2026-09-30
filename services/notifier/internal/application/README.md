# application

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Keputusan pengiriman peringatan dan dedup per versi.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `notifier.go` | Cek duplikat, periksa severity, kirim, catat. |
| `ports.go` | DedupStore dan Sender. |

## Kontrak dan alur

- Handle(ctx,event); Seen/Record(hazard_id,version); Send(ctx,event).

## Dependensi

- contract; dedup dan sender mengimplementasikan port application.

## Aturan penting

- SIAGA/AWAS memicu peringatan; enum dibandingkan dengan urutan eksplisit, bukan urutan string.
- Kirim dahulu lalu catat; crash di celah ini boleh menduplikasi notifikasi.
- Duplikat berarti pasangan hazard_id/version sama; versi baru adalah perubahan sah.
- Application tidak meng-commit offset; consumer melakukannya setelah hasil sukses.

## Langkah implementasi dan verifikasi

- Sepakati apa yang dicatat untuk event non-alert dan event duplikat.
- Verifikasi SIAGA→AWAS menghasilkan dua peringatan sah.
