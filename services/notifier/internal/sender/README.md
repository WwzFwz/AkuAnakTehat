# sender

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Adapter simulasi pengiriman peringatan.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `sender.go` | Implementasi port Sender. |
| `message.go` | Format pesan dari event yang boleh memicu peringatan. |

## Kontrak dan alur

- Send(ctx,event) berhasil hanya setelah efek simulasi yang dipilih selesai.

## Dependensi

- application/contract dan logger lokal.

## Aturan penting

- M1 cukup simulasi yang dapat dibuktikan dari log; tidak menghubungi penerima eksternal sungguhan.
- Hormati timeout/cancellation.
- Bawa event_id/hazard_id/version/correlation_id untuk audit duplikat.

## Langkah implementasi dan verifikasi

- Sepakati bentuk bukti notifikasi simulasi.
- Sediakan cara menguji error sementara dan sukses tanpa mengirim pesan eksternal.
