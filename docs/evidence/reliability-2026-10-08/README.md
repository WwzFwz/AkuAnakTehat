# Pengujian ulang k6 pada 8 Oktober 2026

Ketiga skenario lulus pada revisi `2353479dfa7e6693912a28f302b99e5c559b7eef`. Hasil memenuhi kriteria load test P2 pada spesifikasi M1 untuk konfigurasi lokal ini. Kesimpulan ini tidak menyatakan seluruh requirement proyek atau kapasitas produksi sudah teruji.

## Kriteria spesifikasi dan hasil

| Pemeriksaan | Kriteria | Hasil | Penilaian |
| --- | --- | --- | --- |
| Isolasi latensi sumber | p95 BMKG-only di bawah 300 ms saat request volcanic bersamaan dan delay PVMBG 3 s | p95 seismic 10,9246 ms; 3.001 respons seismic dan 3.163 respons volcanic berhasil | Lulus |
| Koneksi sustained | Sedikitnya 50 koneksi paralel selama 60 s | 50 koneksi TCP ESTABLISHED teramati selama 88,53 s berturut-turut, dengan 71 sampel | Lulus |
| Error sustained | Di bawah 1% setelah HTTP 429 dikeluarkan dari pembilang dan penyebut | 0 dari 9.028 respons non-429 | Lulus |
| Stabilitas | Tidak crash selama beban | Container selain mock PVMBG mempertahankan ID, waktu mulai, dan jumlah restart; tidak ada OOM | Lulus |
| Seismic saat outage PVMBG | Tetap dilayani normal | 480 dari 480 respons sukses, memiliki data, dan BMKG berstatus HEALTHY | Lulus |
| Volcanic saat outage | Data tersimpan dengan penanda basi atau unavailable eksplisit | 480 dari 480 respons memiliki data dan penanda basi; seluruhnya HTTP 200 | Lulus |
| Pemulihan | Kembali normal setelah outage dimatikan tanpa restart aplikasi | 420 dari 730 respons fase recovery sudah HEALTHY; log menunjukkan polling sukses sebelum mock dibuat ulang untuk mengembalikan konfigurasi delay | Lulus, dengan jeda pemulihan |

## Latensi dan throughput

Latency hanya dihitung dari respons bisnis HTTP 200, sehingga respons 429 yang cepat tidak memperbagus percentile. Throughput total mencakup respons berhasil dan penolakan terkontrol; throughput sukses dilaporkan terpisah.

| Skenario | p50 | p95 | p99 | Respons bisnis per detik | HTTP 200 per detik | HTTP 429 |
| --- | --- | --- | --- | --- | --- | --- |
| Seismic dan volcanic bersamaan | 7,71 ms | 10,78 ms | 12,54 ms | 189,17 | 102,65 | 5.196 dari 11.360 |
| Sustained | 8,99 ms | 12,95 ms | 15,33 ms | 475,03 | 100,20 | 33.773 dari 42.801 |
| Outage dan recovery | 4,54 ms | 6,53 ms | 7,83 ms | 32,45 | 32,45 | 0 dari 1.690 |

Angka tabel latency menggabungkan endpoint dalam masing-masing skenario. Ambang khusus BMKG-only memakai metrik seismic tersendiri sebesar 10,9246 ms, bukan p95 gabungan 10,78 ms. Error bisnis dan error autentikasi setelah retry masing-masing 0% pada semua skenario. Sustained juga mencatat 10 respons autentikasi 429 yang ditangani retry; angka ini terpisah dari 33.773 penolakan request bisnis.

## Analisis

**Latency mempunyai margin besar terhadap spesifikasi.** p95 seismic 10,92 ms masih jauh di bawah 300 ms meskipun fetch PVMBG membutuhkan 3 s. Pembacaan dari Canonical Store memisahkan request client dari waktu tunggu sumber. Dibanding gladi sebelumnya yang mencatat 11,33 ms, hasil berada pada kisaran yang sama. Satu pengulangan belum cukup untuk menyimpulkan peningkatan performa yang signifikan.

**Penolakan 429 tinggi, tetapi tidak disembunyikan.** Pada sustained, 78,91% request bisnis ditolak. Semua VU menggunakan satu identitas dengan batas 100 request/detik dan burst 200, sehingga throughput sukses sekitar 100,20 request/detik konsisten dengan konfigurasi tersebut. Spesifikasi mengecualikan penolakan terkontrol dari error selama jumlahnya dilaporkan. Karena itu hasil lulus terhadap spesifikasi, tetapi 475,03 respons/detik tidak boleh disebut sebagai 475 request sukses/detik. Run ini menguji kestabilan dengan pembatasan beban; kapasitas maksimum server belum diukur.

**Pemulihan berhasil, tetapi tidak instan.** Pada jendela recovery 30 s, 57,53% respons sudah menyatakan PVMBG HEALTHY. Sebanyak 310 respons lainnya masih belum sehat, tetapi semuanya tetap HTTP 200. Perilaku ini konsisten dengan polling 5 s, cooldown breaker 10 s, dan fetch sumber 3 s. Log mencatat polling PVMBG kembali sukses pada 15:46:56 WIB, sebelum pengembalian konfigurasi mock pada 15:47:15 WIB. Threshold internal `recovery_available > 0.5` lulus, namun marginnya kecil dan proporsi ini dipengaruhi posisi siklus polling. Persentase tersebut bukan SLA waktu pemulihan. Spesifikasi tidak menetapkan persentase recovery 50%; itu threshold script pengujian.

**Batas pengukuran tetap berlaku.** Model beban menggunakan jumlah VU tetap dengan jeda per iterasi. Koneksi TCP aktif tidak berarti 50 request selalu sedang diproses bersamaan. Pengujian menggunakan volume lokal yang sudah berisi data, satu host Docker, serta halaman 20 record; tidak mengukur beban serentak event 4 MiB. Outage memakai mode error selama 20 s dengan recovery 30 s, sehingga belum menjadi bukti outage panjang atau mode sumber menggantung.

## Konfigurasi dan artefak

- k6 native Windows 0.57.0; Docker Desktop 28.0.4, 16 CPU dan RAM container 7.936.241.664 byte menurut keluaran `docker info`.
- Skenario pertama memakai 10 VU seismic dan 10 VU volcanic selama 60 s. Sustained memakai 50 VU selama 90 s. Outage memakai 5 VU pada setiap fase dan satu VU kontrol.
- PVMBG diatur dengan delay minimum dan maksimum 3 s selama seluruh run. Runner membuat ulang mock sebelum pengujian dan setelah semua skenario untuk memulihkan konfigurasi. Pemulihan di tengah skenario outage hanya melalui endpoint admin.
- [Hasil seismic](load/seismic-only.json), [hasil sustained](load/sustained.json), dan [hasil outage](load/outage.json) menyimpan metrik mentah; berkas `.txt` bersebelahan memuat ringkasan k6.
- [Sampel koneksi](load/connections.json), [lingkungan](load/environment.txt), [batas client](load/client-limits.txt), dan [delay PVMBG](load/pvmbg-delay.txt) mendukung konfigurasi yang dilaporkan.
- [Keadaan awal container](load/containers-before.txt) dan [keadaan akhir container](load/containers-after.txt) memisahkan restart konfigurasi mock dari crash aplikasi. [Log polling tersaring](load/pvmbg-poll-log.json) dan [status akhir sumber](load/final-source-status.json) membuktikan pemulihan.
- [Keluaran runner](load/runner.txt) berakhir dengan tiga PASS dan exit code proses 0. PowerShell memberi label `NativeCommandError` pada progress Docker yang ditulis ke stderr; label ini bukan kegagalan threshold k6.

Semua service kembali sehat dan outage dimatikan setelah run. Kode aplikasi, threshold, dan laporan PDF tidak diubah pada pengujian ini.
