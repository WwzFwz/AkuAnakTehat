# auth-service

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Composition root auth-service; tempat merangkai seluruh dependensi runtime.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** implementasi fondasi awal tersedia dan lolos kompilasi. Cakupan verifikasi runtime fondasi tercatat pada [hasil pengujian](../../../../docs/evidence/foundation/README.md); ini belum bukti P1?P5 lengkap. Berkas yang sudah ada: `main.go`. Tabel rencana di bawah tetap menjadi panduan pemecahan file lanjutan; tidak semua nama file rencana sudah dibuat.

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

- Penerbit token dan pemilik auth-store.
- Rangkai jalur inti sebelum fitur tambahan. Pastikan health tersedia tanpa port publik bila service internal.
