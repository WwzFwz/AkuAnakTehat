# Review sebelum pengumpulan

## Data anggota yang sudah dikonfirmasi

- Kelompok: AkuAnakTehat.
- Dzaky Aurelia Fawwaz — 13523065: perancangan; mock BMKG/PVMBG, Auth Service, ingest Aggregator, relay dan producer Kafka, ketiga consumer, infrastruktur, integrasi, dan pengujian lintas komponen.
- Muhammad Alfansya — 13523005: membantu verifikasi/validasi kesesuaian spesifikasi; query/API internal Aggregator, Client API, CLI Tim Lapangan, dan script k6.

Kontribusi penulisan laporan kedua anggota adalah **Keseluruhan**, sesuai konfirmasi anggota.

## Hal yang masih perlu dilengkapi

1. Tautan laporan PDF final, commit/tag `milestone-1`, dan formulir kelompok. Jangan membuat tag final sebelum kelompok menyetujui versi laporan/kode yang dinilai.

2. Gladi atau demo sinkron sesuai ketentuan pengumpulan.

## Status instrumentasi wajib

Temuan U7 tentang latency retry HTTP sudah ditutup. Log per percobaan, correlation ID yang sama, deadline bersama, sanitasi, batas retry, pembatalan, dan timeout diverifikasi oleh tes module Client API serta go vet. [Bukti U7](../evidence/http-attempts-2026-10-08/README.md) melengkapi regresi stack dan load test yang sudah dijalankan ulang setelah perubahan instrumentasi pada source `a465139`.

Cari `\pending` serta `[Isi` untuk memeriksa apakah masih ada placeholder yang tertinggal.

## Pemeriksaan fakta dan cakupan

- Rujukan kode dan bukti final mengikuti `4976a1f`. Regresi serta load memakai source `a465139`, termasuk instrumentasi per percobaan HTTP.
- Regresi stack final lulus dalam 306,263 s; integrasi PostgreSQL juga lulus. Bukti pada `docs/evidence/final-2026-10-08/`.
- Metrik final berasal dari native Windows k6. P95 seismic 12,51 ms adalah metrik endpoint seismic; p95 gabungan skenario berbeda.
- Sustained berisi 42.511 respons bisnis, 8.793 sukses, 33.718 bisnis 429, dan 9 auth 429; 50 TCP teramati selama 87,78 s. Error 0% memakai denominator tanpa 429. Throughput sukses 97,59/s dan total 471,80/s.
- Outage singkat menghasilkan 480/480 seismic sehat, 435/480 volcanic dengan status basi, dan recovery 625/730 HEALTHY. Deteksi dan pemulihan tidak instan.
- Outage tambahan 600 s dengan recovery 90 s lulus tanpa restart container. Seluruh 30.753 respons bisnis sukses; 14.281 pembacaan seismic selama outage sehat.
- Bootstrap dengan source bersih, kredensial baru, dan enam volume kosong lulus. Migrasi versi 3 bersih; 42 hazard diterima API serta seluruh consumer. Volume, kredensial, dan hazard acuan stack utama dipertahankan.
- Bootstrap memakai Docker serta cache build host yang tersedia; bukan instalasi mesin baru dari nol. Percobaan runner awal salah meminta limit consumer 500, diperbaiki menjadi 200, lalu diulang dengan project kosong baru. Artefak awal tetap disimpan.
- Singleflight, micro-cache, LISTEN/NOTIFY, schema_observations dan header deadline belum dibangun.
- Tidak ada klaim exactly-once notifikasi, revocation access JWT seketika, atau HA satu broker RF1.
- Spesifikasi menerima kutipan log/hasil ukur; tidak semua bukti harus berupa screenshot. Listing kode bukan bukti runtime dan diberi label berbeda.
- Network ownership diperiksa pada Compose dan konfigurasi aplikasi. Operator Docker serta container test merupakan akses administratif, bukan dibatasi sebagai adversary oleh network aplikasi.
- Periksa kembali pengumuman asisten yang datang setelah dokumen sumber. Dokumen M1 menyebut tenggat 9 Oktober 2026 pukul 23.59 WIB dan nama `IF4031_M1_<NamaKelompok>.pdf`.

## Berkas sumber acuan

Dokumen M1 (29 halaman), dokumen terpusat (10 halaman), dan rancangan DOCX diberikan pengguna. Nama/hash dicatat pada `assets/input-documents.json`. Berkas asli tidak diduplikasi ke Git. Pembahasan teori teknologi menggunakan dokumentasi primer yang dirujuk dalam daftar pustaka.
