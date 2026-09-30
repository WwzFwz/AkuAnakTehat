# authz

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Menentukan apakah scope terverifikasi boleh memenuhi permintaan.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `policy.go` | Kebijakan izin ringkasan/raw dan field eksplisit. |
| `scope.go` | Nama scope dan pemeriksaan keanggotaan. |

## Kontrak dan alur

- Authorize(claims, requestedFields, rawRequested) menghasilkan izin atau insufficient_scope.

## Dependensi

- authn untuk VerifiedClaims; dipakai application/handler.

## Aturan penting

- Media yang meminta latitude/longitude/source_ref_id/attributes secara eksplisit ditolak403.
- Scope berasal dari token terverifikasi, bukan role atau flag dari client.
- Default-deny untuk operasi/field baru.

## Langkah implementasi dan verifikasi

- Sepakati arti fields/include/raw endpoint dengan HTTP adapter.
- Verifikasi Media ditolak untuk setiap bentuk permintaan raw.
