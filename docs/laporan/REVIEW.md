# Review sebelum pengumpulan

## Data anggota yang sudah dikonfirmasi

- Kelompok: AkuAnakTehat.
- Dzaky Aurelia Fawwaz — 13523065: perancangan; mock BMKG/PVMBG, Auth Service, ingest Aggregator, relay dan producer Kafka, ketiga consumer, infrastruktur, integrasi, dan pengujian lintas komponen.
- Muhammad Alfansya — 13523005: membantu verifikasi/validasi kesesuaian spesifikasi; query/API internal Aggregator, Client API, CLI Tim Lapangan, dan script k6.

Kontribusi penulisan laporan kedua anggota adalah **Keseluruhan**, sesuai konfirmasi anggota.

## Placeholder yang memang belum diketahui

1. Deklarasi penggunaan LLM (`sections/12-deklarasi-ai.tex`) sengaja dikosongkan untuk diisi sendiri oleh anggota kelompok.
2. Status/tautan run GitHub Actions dari akun yang punya akses (`sections/11-verifikasi-dan-batasan.tex`). API tanpa autentikasi mengembalikan 404 saat diperiksa; bukan bukti workflow lulus maupun gagal.
3. Tautan laporan PDF final dan commit/tag pengumpulan (`appendices/a-matriks-kriteria.tex`). Jangan membuat tag final sebelum kelompok menyetujui versi laporan/kode yang dinilai.

4. Scan secret untuk histori revisi pengumpulan. Pemeriksaan 8 Oktober menggunakan Gitleaks 8.24.2 melaporkan 46 commit diperiksa hingga `2353479` tanpa temuan. Bukti tersimpan pada `docs/evidence/reliability-2026-10-08/secret-scan.txt`; commit sesudah revisi itu tidak otomatis tercakup.

## Status instrumentasi wajib

Temuan U7 tentang latency retry HTTP sudah ditutup. Log per percobaan, correlation ID yang sama, deadline bersama, sanitasi, batas retry, pembatalan, dan timeout diverifikasi oleh tes module Client API serta go vet. [Bukti U7](../evidence/http-attempts-2026-10-08/README.md) tersedia terpisah dari regresi stack dan load test yang belum dijalankan ulang setelah perubahan instrumentasi.

Cari `\pending` serta `[Isi` untuk meninjau bagian yang belum final. Placeholder ini disengaja agar data tidak dikarang.

## Pemeriksaan fakta dan cakupan

- Rujukan kode mengikuti `5cb1367`, termasuk instrumentasi per percobaan HTTP. Bukti regresi stack dan load test tetap berasal dari implementasi `2353479`.
- Bukti gladi awal tetap dipin pada `811e314`. Pengujian setelah perbaikan dipin pada `6cbcf52`, termasuk regresi 306,951 s dan k6 terbaru; angka P2 dalam laporan memakai run terbaru. Pengujian memakai volume yang sudah ada, bukan mesin kosong.
- Metrik final berasal dari native Windows k6, bukan direktori `load-docker-invalid` yang memiliki durasi negatif.
- P95 seismic 10,92 ms adalah metrik endpoint seismic; p95 gabungan skenario berbeda. Throughput semua respons bukan throughput sukses.
- Sustained terbaru berisi 42.801 respons bisnis, 9.028 sukses, 33.773 bisnis 429, dan 10 auth 429; 50 TCP teramati selama 88,53 s. Error 0% memakai denominator tanpa 429. Throughput sukses 100,20/s dan total respons 475,03/s.
- Outage terbaru menghasilkan 480/480 seismic sehat dan 480/480 volcanic dengan data basi. Recovery 420/730 HEALTHY sebelum pengembalian konfigurasi mock; pemulihan tidak instan.
- Outage load yang direkam 20 s dan recovery 30 s. Jangan menyebutnya bukti outage 10–20 menit yang sudah dijalankan.
- Singleflight, micro-cache, LISTEN/NOTIFY, schema_observations dan header deadline belum dibangun.
- Tidak ada klaim exactly-once notifikasi, revocation access JWT seketika, atau HA satu broker RF1.
- Spesifikasi menerima kutipan log/hasil ukur; tidak semua bukti harus berupa screenshot. Listing kode bukan bukti runtime dan diberi label berbeda.
- Network ownership diperiksa pada Compose dan konfigurasi aplikasi. Operator Docker serta container test merupakan akses administratif, bukan dibatasi sebagai adversary oleh network aplikasi.
- Periksa kembali pengumuman asisten yang datang setelah dokumen sumber. Dokumen M1 menyebut tenggat 9 Oktober 2026 pukul 23.59 WIB dan nama `IF4031_M1_<NamaKelompok>.pdf`.

## Berkas sumber acuan

Dokumen M1 (29 halaman), dokumen terpusat (10 halaman), dan rancangan DOCX diberikan pengguna. Nama/hash dicatat pada `assets/input-documents.json`. Berkas asli tidak diduplikasi ke Git. Pembahasan teori teknologi menggunakan dokumentasi primer yang dirujuk dalam daftar pustaka.
