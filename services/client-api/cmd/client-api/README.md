# client-api

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Composition root client-api; tempat merangkai seluruh dependensi runtime.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** composition root client-api sudah merangkai JWT verifier, upstream Aggregator, projection, rate limit, dan batas konkurensi. Bukti runtime end-to-end dan P1-P5 lengkap masih terpisah.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `main.go` | Load config, bangun dependensi, jalankan HTTP/worker, dan graceful shutdown. |

## Kontrak dan alur

- main memulai service dan menghentikannya saat SIGTERM; bukan tempat logika bisnis.

## Dependensi

- Package milik service ini; tidak meng-import module service lain.

## Aturan penting

- Secret wajib tidak ada→gagal startup; dependensi belum siap→retry terbatas/backoff dan readiness503.
- Jangan menjadikan semua dependency wajib hidup untuk liveness.
- Selesaikan pekerjaan sesuai deadline shutdown; commit consumer hanya pekerjaan yang sudah selesai.

## Langkah implementasi dan verifikasi

- API downstream dengan verifikasi JWT, otorisasi, dan proyeksi.
- Rangkai jalur inti sebelum fitur tambahan. Pastikan health tersedia tanpa port publik bila service internal.
