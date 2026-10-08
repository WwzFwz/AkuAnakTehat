# Laporan M1 AkuAnakTehat

Sumber LaTeX modular untuk review kelompok, beserta PDF, diagram vector, potongan kode asli, dan tabel/grafik dari bukti pengujian. Nama keluaran: `IF4031_M1_AkuAnakTehat.pdf`.

## Mengedit di VS Code

1. Buka `laporan.code-workspace` melalui **File > Open Workspace from File** agar konfigurasi VS Code pada folder laporan aktif.
2. Pasang ekstensi LaTeX Workshop bila belum ada. Ekstensi PlantUML opsional untuk preview sumber diagram.
3. Edit bagian yang sesuai pada tabel berikut. Semua file bagian memiliki komentar `TeX root` ke `main.tex`.
4. Jalankan task **LaTeX Workshop: Build LaTeX project**, pilih **Build laporan (Tectonic)**, atau gunakan perintah PowerShell di bawah.
5. Buka `build/main.pdf` untuk preview/SyncTeX; PDF dengan nama pengumpulan diperbarui oleh build script.

Konfigurasi `.vscode` berada di folder laporan, tidak mengubah pengaturan editor seluruh repository. Auto-build dimatikan agar tidak mengunduh paket/compile setiap kali mengetik.

## Peta file untuk review

| File | Bagian yang diedit |
| --- | --- |
| `main.tex` | Urutan penggabungan seluruh bagian. |
| `preamble.tex` | Font, warna, margin, tabel, listing, macro gambar dan placeholder. |
| `metadata.tex` | Identitas/tanggal, revisi kode acuan, dan kontribusi penulisan yang belum diketahui. |
| `sections/00-sampul.tex` | Sampul. |
| `sections/01-kontribusi.tex` | Kontribusi desain, implementasi, dan penulisan sesuai spesifikasi. |
| `sections/02-deskripsi-sistem.tex` | Tujuan dan lingkup M1. |
| `sections/03-arsitektur.tex` | Diagram, service, protokol, container, storage ownership. |
| `sections/04-teknologi-dan-asumsi.tex` | Teknologi, alasan, asumsi dan perubahan dari rancangan awal. |
| `sections/05-menjalankan-sistem.tex` | Ringkasan start, check, CLI dan trace. |
| `sections/06-p1-interoperabilitas.tex` | P1: pemetaan, tolerant reader, schema evolution, transaksi. |
| `sections/07-p2-konkurensi.tex` | P2: isolasi, proteksi, stale, konfigurasi dan hasil load. |
| `sections/08-p3-autentikasi.tex` | P3: trust, JWT, scope, refresh, risiko dan bukti. |
| `sections/09-p4-service-dan-storage.tex` | P4: service boundary, akses DB, perbandingan storage. |
| `sections/10-p5-pubsub.tex` | P5: outbox, Kafka, delivery, dedup, DLQ dan catch-up. |
| `sections/11-verifikasi-dan-batasan.tex` | Hasil lokal, kegagalan yang diperbaiki, status CI, batas klaim. |
| `sections/12-deklarasi-ai.tex` | Deklarasi LLM dan bagian yang perlu dikonfirmasi anggota. |
| `appendices/a-matriks-kriteria.tex` | Pemetaan tiap kriteria spesifikasi ke mekanisme/bukti. |
| `appendices/b-indeks-bukti.tex` | Indeks file bukti, contoh trace, placeholder demo. |
| `references.tex` | Spesifikasi, rancangan awal, kode dan dokumentasi primer. |
| `diagrams/01-*.puml` sampai `05-*.puml`, `07-*.puml` | Enam sumber diagram arsitektur dan sequence; nomor 06 dipakai grafik pengujian. |
| `assets/figures/` | PDF/SVG diagram dan PDF/PNG grafik yang sudah bisa dipakai. |
| `assets/code/`, `assets/evidence/` | Kutipan kode, petikan log, tabel hasil/trace yang dihasilkan script. |
| `assets/provenance.json` | Revisi acuan, path sumber, SHA256, dan awal baris excerpt. |
| `REVIEW.md` | Daftar hal yang masih harus diisi/diverifikasi sebelum dikumpulkan. |
| `BUILD-STATUS.md` | Hasil pemeriksaan kompilasi, tautan sumber, dan asal bukti. |

## Build lokal Windows

Dari root repository:

```powershell
# Sekali saja bila belum mempunyai Tectonic:
powershell -NoProfile -File doc/laporan/scripts/bootstrap-tools.ps1

# Build memakai gambar yang sudah disertakan:
powershell -NoProfile -File doc/laporan/scripts/build.ps1
```

Bootstrap mengambil Tectonic 0.17.0 dan PlantUML 1.2026.8 dari rilis resmi ke `.local/report-tools` (diabaikan Git); tidak menginstal ke sistem atau mengubah PATH. Tectonic mengunduh paket TeX yang diperlukan pada penggunaan pertama, sehingga build pertama lebih lama dan membutuhkan jaringan. Build selanjutnya memakai cache. Jika sudah memasang Tectonic, script mencari PATH; alternatifnya set `TECTONIC_BINARY` ke executable pilihan.

Alternatif distribusi TeX lengkap: dari folder ini jalankan `latexmk -pdf -outdir=build main.tex`, lalu salin `build/main.pdf` ke nama PDF pengumpulan. Bibliografi menggunakan `thebibliography`, tanpa ketergantungan Biber. Build script yang disertakan telah diarahkan ke Tectonic.

## Mengubah dan merender diagram

Diagram memakai [PlantUML](https://plantuml.com/command-line), dengan layout ELK bawaan jar untuk arsitektur dan konektor siku-siku sehingga tidak memerlukan Graphviz eksternal. File `.puml` dapat diedit sebagai teks. Tidak perlu mengambil screenshot dari web: PDF vector yang disertakan tajam ketika diperbesar.

Untuk membuat ulang seluruh aset diperlukan Java, Python 3, `matplotlib`, dan `PyMuPDF`, serta jar PlantUML dari bootstrap:

```powershell
py -m pip install matplotlib PyMuPDF
powershell -NoProfile -File doc/laporan/scripts/build.ps1 -RefreshAssets
```

Bila jar berada di lokasi lain, set `PLANTUML_JAR`. Generator membaca revisi yang dipin pada `metadata.tex`, bukan diam-diam mengambil source terbaru. Ganti revisi hanya setelah bukti pengujian untuk revisi baru diverifikasi. Angka narasi P2 juga harus ditinjau ketika memperbarui bukti; script tidak menulis ulang analisis secara otomatis.

Jika render lokal tidak tersedia, sumber PlantUML tetap dapat ditempel pada editor PlantUML pilihan. Jangan unggah secret atau log privat ke layanan eksternal. Sumber diagram di sini hanya menjelaskan arsitektur, tidak memuat kredensial.

## Screenshot dan potongan kode

Spesifikasi menerima potongan log, pengukuran, atau keluaran load test sebagai bukti; screenshot bukan satu-satunya bentuk yang sah. Laporan memakai listing kode yang dapat disalin serta transkrip test nyata. Tidak ada screenshot palsu. Jika menambah screenshot demo, simpan pada `assets/screenshots/`, beri caption perintah/skenario/tanggal, sensor token, dan gunakan `\includegraphics`. Jangan mengganti angka bukti lama dengan angka yang belum diukur.

Rujukan kode/bukti dipin ke `f6e99cc`, hasil eksekusi **6 Oktober 2026**, sedangkan tanggal penyusunan laporan **8 Oktober 2026**. Status CI tetap belum terkonfirmasi; respons API GitHub 404 tidak membuktikan workflow gagal.

Seluruh halaman PDF memakai A4 potret, termasuk diagram arsitektur. Susunan diagram dibuat lebih ringkas agar label tetap terbaca pada lebar halaman.
