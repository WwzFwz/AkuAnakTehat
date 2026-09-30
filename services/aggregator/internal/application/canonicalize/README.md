# canonicalize

[Panduan service](../../../README.md) · [Peta repository](../../../../../README.md)

Pemetaan data sumber ke HazardEvent; pemilik tipe input yang digunakan adapter dan ingest.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `input.go` | SeismicInput, VolcanicInput, dan referensi gunung, termasuk unknown fields. |
| `seismic_mapper.go` | Pemetaan gempa dan severity dari magnitude/warning. |
| `volcanic_mapper.go` | Pemetaan laporan gunung api dan referensi nama/koordinat. |
| `tsunami_correlator.go` | Menggabungkan warning terkait secara deterministik. |

## Kontrak dan alur

- Operasi usulan: MapSeismic(input, warnings) dan MapVolcanic(input, volcano).
- Mapper menghasilkan field bisnis; identitas persisten, version, dan transaksi ditangani ingest.

## Dependensi

- domain/hazard dan domain/tsunami.
- ingest dan adapter sumber boleh meng-import canonicalize; canonicalize tidak meng-import keduanya.

## Aturan penting

- Tanpa HTTP, SQL, atau publish Kafka.
- Jika ada warning, severity mengikuti aturan warning; jangan otomatis memilih maksimum antara severity magnitude dan warning.
- volcano_id tidak dikenal menghasilkan error terklasifikasi untuk karantina.
- Unknown fields masuk attributes; susunan warning harus stabil agar tidak menghasilkan perubahan hash palsu.

## Langkah implementasi dan verifikasi

- Sepakati tipe input sebelum membuat decoder.
- Uji ambang 5.0/6.5, warning terlambat, beberapa warning, dan atribut baru.
