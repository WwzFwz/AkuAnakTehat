# simulation

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

State dan kontrol runtime delay, outage, serta evolusi skema PVMBG.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `state.go` | State terlindungi mutex atau snapshot immutable. |
| `delay.go` | Delay yang dapat dibatalkan melalui context. |
| `outage.go` | Mode error/hang untuk endpoint data. |
| `schema.go` | Aktivasi confidence_level untuk laporan baru. |

## Kontrak dan alur

- SetOutage(enabled, mode), SetSchemaVersion(version), dan Snapshot().
- Body eksplisit pada script demo lebih mudah diulang daripada toggle tanpa body.

## Dependensi

- Dipakai HTTP handler dan generator lokal; tidak bergantung BNPB.

## Aturan penting

- Endpoint admin tidak ikut diputus oleh outage.
- Perubahan state berlaku tanpa restart.
- Jangan memegang lock selama sleep/hang.
- Schema awal tidak memuat confidence_level; laporan lama tetap berdampingan dengan laporan baru.

## Langkah implementasi dan verifikasi

- Implementasikan admin idempoten dengan body enabled/version.
- Verifikasi outage dapat dimatikan saat data request sedang hang.
