# contract

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Kontrak event milik notifier; salinan lokal yang kompatibel dengan envelope producer.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `event.go` | Envelope schema_version/event_id/event_type/hazard_id/version/correlation_id dan HazardEvent. |
| `decode.go` | Decode tolerant reader dan validasi field wajib. |

## Kontrak dan alur

- Menerima hazard.upserted schema_version1 berisi HazardEvent lengkap.
- DTO diperbarui berdasarkan docs/api, bukan import struct Aggregator.

## Dependensi

- Standard library; dipakai consumer dan application lokal.

## Aturan penting

- Field tambahan tidak membuat pesan valid gagal dibaca.
- Unknown attributes dipertahankan sebagai JSON.
- Validasi konsistensi hazard_id envelope/payload serta versi positif.
- Published timestamp di envelope harus diberi definisi yang jelas; jangan menyamakan waktu dibuat dengan ACK tanpa kesepakatan.

## Langkah implementasi dan verifikasi

- Sepakati contoh event dengan producer.
- Uji field tambahan, payload rusak, enum/schema version yang tidak didukung.
