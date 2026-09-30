# tsunami

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Model warning dan aturan korelasi ke gempa, tanpa akses jaringan atau storage.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A sudah diimplementasikan. Berkas tersedia: `warning.go`. Cakupan pengujian ada di [bukti ingest](../../../../../docs/evidence/ingest/README.md). Tabel rencana di bawah adalah panduan pemecahan tanggung jawab; sebagian operasi digabung dalam file yang tersedia.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `warning.go` | TsunamiWarning dan atribut tambahan yang dipertahankan. |
| `severity.go` | Urutan Waspada, Siaga, dan Awas untuk memilih warning tertinggi. |

## Kontrak dan alur

- Korelasi menggunakan related_event_id, bukan warning_id.
- Model mendukung beberapa warning untuk satu gempa.

## Dependensi

- Standard library; dipakai canonicalize, ingest, dan adapter BMKG.

## Aturan penting

- Warning bukan HazardEvent baru.
- Metadata internal mock tidak menjadi ketergantungan model ini.
- Aturan warning hanya diterapkan pada gempa yang sesuai kontrak potential_tsunami.

## Langkah implementasi dan verifikasi

- Tentukan representasi field tambahan.
- Verifikasi pilihan severity tertinggi saat beberapa warning terkait tersedia.
