# cache

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Rencana tambahan: singleflight dan micro-cache sebelum proyeksi.

**Pemilik rencana:** B. **Tahap:** Tambahan opsional; dikerjakan setelah baseline terbukti.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `cache.go` | Entri dengan TTL dan batas kapasitas. |
| `key.go` | Key dari request/filter yang dinormalisasi. |
| `coalescer.go` | Penggabungan fetch identik dengan context bersama yang terpisah. |

## Kontrak dan alur

- Get/Set dan FetchShared menjadi usulan API; belum diperlukan oleh jalur inti.

## Dependensi

- DTO application serta HTTP fetch yang dibungkus; tidak menambah Redis.

## Aturan penting

- Default disabled; TTL<= 1 s dan jumlah entri dibatasi.
- Jangan cache error; key harus mencakup semua filter/cursor yang memengaruhi hasil.
- Setiap caller punya deadline sendiri; pembatalan caller pertama tidak membatalkan semua caller.
- Authz dan proyeksi tetap dilakukan per request, bukan berdasarkan identitas pengisi cache.

## Langkah implementasi dan verifikasi

- Implementasikan hanya setelah P1–P5 dasar bekerja.
- Ukur on/off dan verifikasi tidak ada kebocoran lintas scope.
