# bmkg-mock

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Composition root bmkg-mock; tempat merangkai seluruh dependensi runtime.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

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

- Mock gempa dan warning tsunami dengan kontrak independen.
- Rangkai jalur inti sebelum fitur tambahan. Pastikan health tersedia tanpa port publik bila service internal.
