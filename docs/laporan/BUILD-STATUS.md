# Validasi keluaran laporan

Dijalankan lokal pada 8 Oktober 2026. Ini validasi artefak laporan, bukan pengulangan runtime aplikasi.

- `scripts/build.ps1 -RefreshAssets` selesai dengan exit code 0.
- Setelah audit README dan persyaratan ditambahkan, `scripts/build.ps1 -RefreshAssets` kembali selesai dengan exit code 0. Halaman hasil pengukuran, verifikasi, dan indeks bukti serta diagram yang berubah dirender untuk inspeksi; ukuran seluruh halaman dan log tata letak diperiksa.
- Revisi editorial berikutnya dibangun dengan `scripts/build.ps1` dan selesai dengan exit code 0. Sampul, alur arsitektur, asumsi, batasan, dan listing kode dirender untuk inspeksi. Seluruh teks sampul berwarna hitam dan nomor baris berada di dalam bingkai kode.
- Setelah rincian penggunaan LLM dikosongkan untuk diisi anggota, `scripts/build.ps1` kembali berhasil. PDF diperiksa untuk memastikan rincian penggunaan tersebut sudah tidak tercantum, placeholder tersedia, dan tata letak tetap valid.
- PDF keluaran: `IF4031_M1_AkuAnakTehat.pdf`, 38 halaman.
- Verifikasi akhir disusun menurut skenario dan hasil P1 hingga P5 serta U7. Build dengan `-RefreshAssets` berhasil; petikan regresi kini berasal dari run 306,951 s, sedangkan bukti tes instrumentasi Client API ditautkan terpisah. Bagian verifikasi dirender untuk memeriksa pemisahan halaman.
- Enam diagram PlantUML, satu grafik hasil uji, tiga kutipan kode, serta tabel/log bukti berhasil dibuat ulang.
- 50 tautan repository pada PDF diperiksa keberadaannya di revisi yang dirujuk.
- 9 hash input kode/bukti cocok dengan `assets/provenance.json`. Setiap input menyebut revisinya; kode memakai `5cb1367`, sedangkan input load terbaru memakai `6cbcf52`.
- Tidak ada referensi LaTeX tak terdefinisi, karakter hilang, maupun peringatan overfull/underfull box pada log kompilasi akhir.
- Seluruh halaman termasuk arsitektur berukuran A4 potret (595,28 kali 841,89 pt) tanpa rotasi. Tidak ada perubahan format menjadi halaman landscape.
- Diagram arsitektur menggunakan ELK dengan garis ortogonal; sudut konektor berbentuk siku-siku.
- Tool: Tectonic 0.17.0, PlantUML 1.2026.8, Python 3.12, PyMuPDF, matplotlib.

Tectonic mengeluarkan pesan konfigurasi Fontconfig lokal serta peringatan versi PDF gambar 1.7 terhadap setting internal 1.5. Build tetap berhasil dan diagram/font terbaca pada PDF hasil render. Ini bukan peringatan bahwa pengujian aplikasi gagal.

Sumber fakta adalah spesifikasi pengguna serta implementasi dan bukti yang dipin. Angka utama P2 memakai pengujian ulang 8 Oktober 2026, dengan p95 seismic 10,92 ms dan 50 TCP selama 88,53 s. Bukti historis 6 Oktober serta gladi awal 8 Oktober tetap tersedia. Scan histori mencakup 46 commit hingga `2353479`. Temuan U7 pada latency retry HTTP telah ditutup dengan tes module Client API dan go vet. Regresi stack serta k6 belum dijalankan ulang untuk perubahan instrumentasi tersebut. Rincian LLM, CI, scan secret revisi pengumpulan, dan informasi pengumpulan yang belum terkonfirmasi tetap ditandai pada `REVIEW.md`. Laporan perlu review anggota sebelum pengumpulan.
