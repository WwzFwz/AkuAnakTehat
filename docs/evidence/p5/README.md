# p5

[Peta repository](../../../README.md)

Bukti nyata P5: fan-out, backlog, dan consumer baru.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `run-notes.md` | Tanggal, commit, lingkungan, perintah, dan interpretasi hasil. |
| `output.log` | Log atau respons asli yang sudah disanitasi. |
| `screenshots/` | Folder opsional tangkapan layar ketika bukti visual diperlukan. |

## Kontrak dan alur

- Log producer serta dua consumer dengan group berbeda.
- Consumer mati tidak menahan producer/consumer lain; backlog dan catch-up setelah start.
- Consumer ketiga menerima event tanpa perubahan producer.
- Catatan duplikasi/commit/retry/DLQ sesuai konfigurasi yang sebenarnya.

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
