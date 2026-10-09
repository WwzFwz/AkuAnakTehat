# Validasi keluaran laporan

Dijalankan lokal pada 8 Oktober 2026. Pemeriksaan ini memvalidasi artefak laporan; hasil runtime dirujuk melalui bukti pengujian final.

- Build `scripts/build.ps1 -RefreshAssets` dan build akhir selesai dengan exit code 0.
- PDF `IF4031_M1_AkuAnakTehat.pdf` berisi 39 halaman A4 potret, 595,28 kali 841,89 pt, tanpa rotasi.
- Sampul seluruhnya hitam, tanpa teks revisi acuan. Nomor baris berada di dalam bingkai listing kode.
- Enam diagram PlantUML, satu grafik hasil k6, tiga kutipan kode, serta tabel dan petikan hasil pengujian berhasil dibangkitkan.
- Grafik dan tabel P2 memakai bukti final `4976a1f`. P95 seismic 12,51 ms; 50 TCP selama 87,78 s. Petikan regresi memakai run 306,263 s.
- Bagian verifikasi mencakup hasil akhir P1 hingga P5, U7, outage 600 detik, dan bootstrap enam volume kosong. Halaman terkait dirender untuk inspeksi.
- Seluruh 56 tautan repository pada PDF tersedia pada revisi yang dirujuk.
- Sembilan hash input laporan cocok dengan `assets/provenance.json`.
- Sebanyak 44 checksum artefak dan sumber pada manifest bukti final cocok dengan byte Git setelah normalisasi LF.
- Tidak ada referensi LaTeX tak terdefinisi, karakter hilang, maupun peringatan overfull/underfull box pada log akhir.
- Pernyataan penggunaan AI sudah dicantumkan untuk kedua anggota. Kontribusi penulisan kedua anggota tetap Keseluruhan.
- Diagram arsitektur tetap memakai ELK dengan konektor ortogonal.
- Tool yang digunakan adalah Tectonic 0.17.0, PlantUML 1.2026.8, Python 3.12, PyMuPDF, dan matplotlib.

Tectonic mengeluarkan pesan konfigurasi Fontconfig lokal serta peringatan versi PDF gambar 1.7 terhadap setting internal 1.5. Build berhasil dan hasil render terbaca.

Source runtime yang diuji adalah `a465139`, termasuk instrumentasi per percobaan HTTP. Bukti serta runner final dipin pada `4976a1f`. Regresi stack, k6, outage panjang, dan bootstrap terisolasi lulus; stack utama sudah dipulihkan dengan volume, kredensial, dan data acuan tetap tersedia. Workflow hosted CI `foundation` run `#21` berhasil pada branch `main` untuk commit `013099b`; scan Gitleaks final hingga `48987b7` tidak menemukan kebocoran. Finalisasi pengumpulan tetap mengikuti daftar pada `REVIEW.md`.
