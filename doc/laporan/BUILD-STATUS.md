# Validasi keluaran laporan

Dijalankan lokal pada 8 Oktober 2026. Ini validasi artefak laporan, bukan pengulangan runtime aplikasi.

- `scripts/build.ps1 -RefreshAssets` selesai dengan exit code 0.
- PDF keluaran: `IF4031_M1_AkuAnakTehat.pdf`, 34 halaman.
- Enam diagram PlantUML, satu grafik hasil uji, tiga kutipan kode, serta tabel/log bukti berhasil dibuat ulang.
- 39 tautan sumber diperiksa keberadaannya pada revisi `f6e99cc`.
- 9 hash input kode/bukti cocok dengan `assets/provenance.json`.
- Tidak ada referensi LaTeX tak terdefinisi, karakter hilang, maupun peringatan overfull/underfull box pada log kompilasi akhir.
- Seluruh halaman dirender untuk inspeksi tata letak; halaman arsitektur memakai landscape.
- Tool: Tectonic 0.17.0, PlantUML 1.2026.8, Python 3.12, PyMuPDF, matplotlib.

Tectonic mengeluarkan pesan konfigurasi Fontconfig lokal serta peringatan versi PDF gambar 1.7 terhadap setting internal 1.5. Build tetap berhasil dan diagram/font terbaca pada PDF hasil render. Ini bukan peringatan bahwa pengujian aplikasi gagal.

Sumber fakta adalah spesifikasi pengguna serta implementasi/bukti yang dipin; hasil runtime tetap bertanggal 6 Oktober 2026. Data penulisan, rincian LLM, CI, scan secret revisi final, dan informasi pengumpulan yang belum terkonfirmasi ditandai pada `REVIEW.md` dan laporan. Laporan perlu review anggota sebelum pengumpulan.
