# application

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Orkestrasi baca downstream: otorisasi, panggil Aggregator, lalu proyeksi.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `service.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `hazard_service.go` | Use case list/get hazard. |
| `ports.go` | Port pembacaan Aggregator. |
| `dto.go` | Salinan DTO respons internal dan respons publik milik client-api. |

## Kontrak dan alur

- List(ctx, claims, filter) dan Get(ctx, claims, hazardID).
- Port Aggregator membaca melalui HTTP; tidak mengekspos driver/database.

## Dependensi

- authz, projection, dan DTO lokal. Adapter outbound mengimplementasikan port.

## Aturan penting

- Tidak meng-import module Aggregator, membaca canonical-db, atau menghubungi mock.
- Tidak melewatkan proyeksi pada cabang detail/error/cache.
- Kegagalan Aggregator menghasilkan ketidaktersediaan yang eksplisit.
- Metadata sumber tetap tersedia tanpa membocorkan pesan error internal.

## Langkah implementasi dan verifikasi

- Sepakati HTTP contract terlebih dahulu.
- Hubungkan jalur tanpa cache; cache opsional ditambahkan setelah pengujian baseline.
