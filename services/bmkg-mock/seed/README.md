# seed

[Panduan service](../README.md) · [Peta repository](../../../README.md)

Fixture historis milik BMKG; data demo, bukan data bencana live.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `seismic-events.json` | Minimal20 gempa historis dengan ID tetap. |
| `tsunami-warnings.json` | Warning terkait untuk sebagian gempa berpotensi tsunami. |

## Kontrak dan alur

- Dimuat oleh store saat startup; format sama dengan entitas sumber kecuali metadata internal simulator yang dikelola loader.

## Dependensi

- domain/store lokal; file ini tidak diakses langsung oleh Aggregator.

## Aturan penting

- Seed konsisten lintas restart dan tidak menduplikasi identitas.
- Warning menunjuk gempa yang valid atau skenario kedatangan tertunda yang disengaja.
- Labeli data sintetis dan gunakan timestamp UTC historis.

## Langkah implementasi dan verifikasi

- Buat fixture setelah kontrak domain disepakati.
- Pastikan initial lookback Aggregator mencakup waktu seed.
