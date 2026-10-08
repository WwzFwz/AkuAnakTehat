# Validasi keluaran laporan

Dijalankan lokal pada 8 Oktober 2026. Ini validasi artefak laporan, bukan pengulangan runtime aplikasi.

- `scripts/build.ps1 -RefreshAssets` selesai dengan exit code 0.
- Setelah audit README dan persyaratan ditambahkan, `scripts/build.ps1 -RefreshAssets` kembali selesai dengan exit code 0. Halaman hasil pengukuran, verifikasi, dan indeks bukti serta diagram yang berubah dirender untuk inspeksi; ukuran seluruh halaman dan log tata letak diperiksa.
- Revisi editorial berikutnya dibangun dengan `scripts/build.ps1` dan selesai dengan exit code 0. Sampul, alur arsitektur, asumsi, batasan, dan listing kode dirender untuk inspeksi. Seluruh teks sampul berwarna hitam dan nomor baris berada di dalam bingkai kode.
- PDF keluaran: `IF4031_M1_AkuAnakTehat.pdf`, 38 halaman.
- Enam diagram PlantUML, satu grafik hasil uji, tiga kutipan kode, serta tabel/log bukti berhasil dibuat ulang.
- 50 tautan repository pada PDF diperiksa keberadaannya di revisi yang dirujuk.
- 9 hash input kode/bukti cocok dengan `assets/provenance.json`. Setiap input menyebut revisinya; kode memakai `2353479`, sedangkan input load terbaru memakai `6cbcf52`.
- Tidak ada referensi LaTeX tak terdefinisi, karakter hilang, maupun peringatan overfull/underfull box pada log kompilasi akhir.
- Seluruh halaman termasuk arsitektur berukuran A4 potret (595,28 kali 841,89 pt) tanpa rotasi. Tidak ada perubahan format menjadi halaman landscape.
- Diagram arsitektur menggunakan ELK dengan garis ortogonal; sudut konektor berbentuk siku-siku.
- Tool: Tectonic 0.17.0, PlantUML 1.2026.8, Python 3.12, PyMuPDF, matplotlib.

Tectonic mengeluarkan pesan konfigurasi Fontconfig lokal serta peringatan versi PDF gambar 1.7 terhadap setting internal 1.5. Build tetap berhasil dan diagram/font terbaca pada PDF hasil render. Ini bukan peringatan bahwa pengujian aplikasi gagal.

Sumber fakta adalah spesifikasi pengguna serta implementasi dan bukti yang dipin. Angka utama P2 memakai pengujian ulang 8 Oktober 2026, dengan p95 seismic 10,92 ms dan 50 TCP selama 88,53 s. Bukti historis 6 Oktober serta gladi awal 8 Oktober tetap tersedia. Scan histori mencakup 46 commit hingga `2353479`. Audit masih mencatat celah U7 pada latency per percobaan retry HTTP client-api; laporan tidak mengklaim semua kewajiban sudah terpenuhi. Rincian LLM, CI, scan secret revisi pengumpulan, dan informasi pengumpulan yang belum terkonfirmasi tetap ditandai pada `REVIEW.md`. Laporan perlu review anggota sebelum pengumpulan.
