# p2

[Peta repository](../../../README.md)

Bukti nyata P2: konkurensi, availability, dan graceful degradation.

**Pemilik rencana:** A/B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `run-notes.md` | Tanggal, commit, lingkungan, perintah, dan interpretasi hasil. |
| `output.log` | Log atau respons asli yang sudah disanitasi. |
| `screenshots/` | Folder opsional tangkapan layar ketika bukti visual diperlukan. |

## Kontrak dan alur

- Konfigurasi k6, mesin uji, durasi, koneksi, dan output mentah.
- p95 BMKG-only saat PVMBG delay3 s; throughput/p50/p95/p99/error/429.
- Respons stale/unavailable saat outage dan pemulihan tanpa restart.

## Dependensi

- Script demo/loadtest dan implementasi yang telah dijalankan.

## Aturan penting

- Saat ini belum ada bukti eksekusi dan tidak ada klaim lulus.
- Jangan simpan token, secret, atau data sensitif; sanitasi tanpa menghilangkan konteks penting.
- Setiap angka/claim pada laporan harus merujuk hasil nyata yang dapat ditelusuri.
- Simpan hasil gagal yang relevan dengan penjelasan; jangan mengubah hasil agar tampak lulus.

## Langkah implementasi dan verifikasi

- Jalankan skenario sesudah komponen terkait selesai.
- Tambahkan catatan hasil dan tautan bukti dari laporan akhir.
