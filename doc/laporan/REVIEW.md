# Review sebelum pengumpulan

## Data anggota yang sudah dikonfirmasi

- Kelompok: AkuAnakTehat.
- Dzaky Aurelia Fawwaz — 13523065: perancangan; implementasi semua bagian selain jalur 3B.
- Muhammad Alfansya — 13523005: membantu verifikasi/validasi kesesuaian spesifikasi; implementasi jalur 3B.
- Jalur 3B: query/API internal Aggregator, client-api, CLI Tim Lapangan, script k6.

## Placeholder yang memang belum diketahui

1. Kontribusi penulisan tiap anggota (`metadata.tex`). Spesifikasi hlm. 26 meminta desain, implementasi, **dan** penulisan; jangan menghapus kolom ini.
2. Rincian bantuan LLM tambahan per anggota, khususnya 3B (`sections/12-deklarasi-ai.tex`). Jangan mengklaim semua anggota sudah memahami kode sebelum mereka meninjau sendiri.
3. Status/tautan run GitHub Actions dari akun yang punya akses (`sections/11-verifikasi-dan-batasan.tex`). API tanpa autentikasi mengembalikan 404 saat diperiksa; bukan bukti workflow lulus maupun gagal.
4. Tautan laporan PDF final dan commit/tag pengumpulan (`appendices/a-matriks-kriteria.tex`). Jangan membuat tag final sebelum kelompok menyetujui versi laporan/kode yang dinilai.
5. Tanggal/tautan dokumentasi demonstrasi sinkron jika diminta (`appendices/b-indeks-bukti.tex`). Bukti otomatis bukan demo sinkron.

6. Scan secret untuk histori revisi pengumpulan. Bukti yang ditemukan hanya mencakup lima commit sampai `f73c035` pada fondasi; jangan menganggapnya mencakup revisi akhir.

Cari `\pending` serta `[Isi` untuk meninjau bagian yang belum final. Placeholder ini disengaja agar data tidak dikarang.

## Pemeriksaan fakta dan cakupan

- Narasi mengikuti implementasi `f6e99cc`; kode aplikasi tidak diubah saat menyusun laporan.
- Metrik final berasal dari native Windows k6, bukan direktori `load-docker-invalid` yang memiliki durasi negatif.
- P95 seismic 8,62 ms adalah metrik endpoint seismic; p95 gabungan skenario berbeda. Throughput semua respons bukan throughput sukses.
- Sustained: 42.619 bisnis, 9.689 sukses, 32.930 bisnis 429, 10 auth 429; 50 TCP teramati selama 90,46 s. Error 0% memakai denominator tanpa 429.
- Outage load yang direkam 20 s dan recovery 30 s. Jangan menyebutnya bukti outage 10–20 menit yang sudah dijalankan.
- Singleflight, micro-cache, LISTEN/NOTIFY, schema_observations dan header deadline belum dibangun.
- Tidak ada klaim exactly-once notifikasi, revocation access JWT seketika, atau HA satu broker RF1.
- Spesifikasi menerima kutipan log/hasil ukur; tidak semua bukti harus berupa screenshot. Listing kode bukan bukti runtime dan diberi label berbeda.
- Network ownership diperiksa pada Compose dan konfigurasi aplikasi. Operator Docker serta container test merupakan akses administratif, bukan dibatasi sebagai adversary oleh network aplikasi.
- Periksa kembali pengumuman asisten yang datang setelah dokumen sumber. Dokumen M1 menyebut tenggat 9 Oktober 2026 pukul 23.59 WIB dan nama `IF4031_M1_<NamaKelompok>.pdf`.

## Berkas sumber acuan

Dokumen M1 (29 halaman), dokumen terpusat (10 halaman), dan rancangan DOCX diberikan pengguna. Nama/hash dicatat pada `assets/input-documents.json`. Berkas asli tidak diduplikasi ke Git. Pembahasan teori teknologi menggunakan dokumentasi primer yang dirujuk dalam daftar pustaka.
