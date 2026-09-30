# domain

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Kontrak data PVMBG milik mock; tidak bergantung pada model BNPB.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `volcanic_report.go` | VolcanicReport dengan confidence_level opsional. |
| `volcano.go` | Daftar ID demo yang disepakati untuk generator. |

## Kontrak dan alur

- VolcanicReport memakai report_id, volcano_id, alert_level, statistik, reported_at, dan confidence_level jika aktif.

## Dependensi

- Standard library; digunakan generator, store, dan HTTP lokal.

## Aturan penting

- Tidak meng-import tipe Aggregator.
- Nama field JSON dan enum sesuai spesifikasi; metadata simulator tidak ditambahkan ke respons resmi.
- Confidence_level v1 tidak dikirim; ketika ada, nilainya0–1.

## Langkah implementasi dan verifikasi

- Definisikan tipe dan tag JSON.
- Verifikasi bentuk respons terhadap spesifikasi sebelum menghubungkan Aggregator.
