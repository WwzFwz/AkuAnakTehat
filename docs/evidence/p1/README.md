# p1

[Peta repository](../../../README.md)

Bukti nyata P1: pemetaan dan evolusi skema tanpa restart.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `run-notes.md` | Tanggal, commit, lingkungan, perintah, dan interpretasi hasil. |
| `output.log` | Log atau respons asli yang sudah disanitasi. |
| `screenshots/` | Folder opsional tangkapan layar ketika bukti visual diperlukan. |

## Kontrak dan alur

- Respons seismik/vulkanik sebelum dan sesudah schema change.
- Log drift ber-correlation ID dan uptime service tetap.
- Bukti unknown fields masuk attributes; field baru tidak memerlukan migrasi.

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
