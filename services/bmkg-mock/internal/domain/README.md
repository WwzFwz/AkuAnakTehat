# domain

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Kontrak data BMKG milik mock; tidak bergantung pada model BNPB.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `seismic_event.go` | Field SeismicEvent sesuai kontrak. |
| `tsunami_warning.go` | Field warning dan metadata internal_modified_at yang tidak diserialisasi. |

## Kontrak dan alur

- SeismicEvent dan TsunamiWarning tetap terpisah; related_event_id menghubungkannya.

## Dependensi

- Standard library; digunakan generator, store, dan HTTP lokal.

## Aturan penting

- Tidak meng-import tipe Aggregator.
- Nama field JSON dan enum sesuai spesifikasi; metadata simulator tidak ditambahkan ke respons resmi.
- Warning hanya untuk event dengan potential_tsunami=true.

## Langkah implementasi dan verifikasi

- Definisikan tipe dan tag JSON.
- Verifikasi bentuk respons terhadap spesifikasi sebelum menghubungkan Aggregator.
