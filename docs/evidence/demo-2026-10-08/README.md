# Verifikasi kesiapan demo, 8 Oktober 2026

Gladi dijalankan pada Docker Desktop lokal dengan volume yang sudah ada. Ini bukan demonstrasi sinkron bersama penguji dan bukan pengujian bootstrap pada mesin kosong. Start ulang mengikuti README, mempertahankan secret serta volume. Kode aplikasi acuan tetap `f6e99cc`; basis repository saat gladi `0ffb301`. Helper demo dan opsi direktori load test diperiksa dari working tree pada sesi ini; hash sumber dicatat di [manifest](manifest.json).

## Hasil yang benar-benar dijalankan

| Pemeriksaan | Hasil | Bukti |
| --- | --- | --- |
| Bootstrap dan Compose | Secret lama dipertahankan, validasi config berhasil, seluruh service demo start dengan image hasil build | Kondisi akhir pada `services.txt` dan `status.json` |
| Bootstrap serta sembilan module Go | Uji unit dan `go vet` selesai exit code 0 | [modules.txt](modules.txt) |
| Regresi lintas service | Seluruh suite lulus dalam 281,222 detik | [regression.txt](regression.txt) |
| PostgreSQL nyata | Transaksi ingest, watermark, outbox, dan query lulus pada schema uji terisolasi | [postgres.txt](postgres.txt) |
| k6 native Windows | Ketiga skenario lulus dan runner selesai exit code 0 | [load-runner.txt](load-runner.txt), [hasil load](load/) |
| Perintah helper demo | Media, raw 403, dua identitas privileged, inspeksi consumer, skema lama dan baru, outage dan recovery, serta restore berhasil | [helper.txt](helper.txt) |
| Wrapper suite demo | Pemanggilan `verify p4` menjalankan pemeriksaan skema dan rebuild secara nyata | [demo-p4.txt](demo-p4.txt) |
| Keamanan helper | Redirect login tidak diikuti dan nilai kredensial lokal disensor pada output | [helper-unit.txt](helper-unit.txt) |
| Secret scan | Gitleaks 8.24.2 melaporkan 33 commit diperiksa tanpa temuan pada histori yang tersedia di basis `0ffb301` | [secret-scan.txt](secret-scan.txt) |

Gitleaks memeriksa histori Git, bukan secret lokal yang memang diabaikan Git. Pemeriksaan tambahan terhadap 416 berkas tracked dan kandidat saat gladi tidak menemukan nilai literal kredensial lokal. Nilai rahasia tidak dicetak. Hasil scan memiliki batas revisi tersebut dan perlu diperbarui jika kode atau dokumen sensitif berubah sebelum pengumpulan.

## Pengukuran beban terbaru

Lingkungan terekam pada [environment.txt](load/environment.txt) dan [host.json](host.json). Tool memakai k6 0.57.0 native Windows; Docker Desktop 28.0.4 menyediakan 16 CPU dan 7.936.241.664 byte memori. PVMBG sementara diatur delay tepat 3 detik, kemudian dikembalikan ke konfigurasi normal.

| Skenario | Respons bisnis | HTTP 200 | Bisnis 429 | HTTP 200 per detik | p50 / p95 / p99 sukses (ms) | Error selain 429 |
| --- | ---: | ---: | ---: | ---: | --- | ---: |
| seismic-only | 11383 | 6207 | 5176 | 103.27 | 6.08 / 11.17 / 17.99 | 0.00% |
| sustained | 42801 | 9118 | 33683 | 101.20 | 6.48 / 12.80 / 18.60 | 0.00% |
| outage | 1221 | 1221 | 0 | 23.43 | 4.59 / 6.82 / 9.09 | 0.00% |

P95 **seismic saja 11,33 ms**, masih di bawah ambang 300 ms ketika pembacaan volcanic berlangsung bersamaan. Tabel menampilkan percentile gabungan skenario. Pada sustained terdapat 9 penolakan auth 429; metrik auth error setelah mekanisme retry tetap 0%. Sampling [connections.json](load/connections.json) membuktikan sedikitnya 50 koneksi TCP ESTABLISHED selama **87,90 detik berturut-turut**.

Konfigurasi tetap 10 VU seismic dan 10 VU volcanic selama 60 detik; sustained 50 VU selama 90 detik dengan jeda 0,1 detik; outage 20 detik dan recovery 30 detik dengan 5 VU per fase serta satu VU kontrol. Seluruh VU berbagi identitas Tim Lapangan, tetapi sesi refresh terpisah. Tingginya 429 menunjukkan batas beban per identitas; angka throughput seluruh respons tidak disamakan dengan throughput sukses.

Pada ekspor JSON k6 ini, nilai `thresholds: false` berarti threshold tidak dilanggar. Error rate mengeluarkan 429 dari pembilang dan penyebut. Latency hanya menghitung HTTP 200. Hasil ini tidak menggantikan atau menimpa [pengukuran 6 Oktober](../integration/README.md) yang masih menjadi acuan angka laporan.

## Pemulihan dan batas cakupan

- Helper restore dan cleanup suite telah dijalankan. Outage dimatikan, generator PVMBG kembali ke skema 1, serta service demo dijalankan dengan Compose normal. Record sintetis dan volume dipertahankan.
- Skema 1 hanya berlaku untuk record berikutnya; data skema 2 yang sudah tersimpan tetap ada. Group uji unik dan fixture event juga dapat tertinggal sebagai bukti replay.
- Pengukuran berasal dari satu mesin dengan konfigurasi lokal, bukan bukti kapasitas produksi atau ketersediaan banyak host.
- Tidak dilakukan manipulasi TTL atau jam untuk demo expiry. CLI benar-benar melewati access TTL 60 detik pada regresi.
- Uji outage singkat bukan bukti outage 10 atau 20 menit. Gunakan perintah manual pada panduan jika durasi lain diminta penguji.
- Pada saat artefak demo ini dibuat, status workflow GitHub Actions hosted belum terkonfirmasi; hasil lokal tidak dianggap sebagai hasil hosted CI. Verifikasi berikutnya tercatat pada laporan utama.
- Pada saat artefak demo ini dibuat, laporan masih memiliki placeholder. Status terbaru terkait tag final, formulir, dan demo sinkron mengikuti `docs/laporan/REVIEW.md`.

## Mengulang

Ikuti [panduan demo](../../../scripts/demo/README.md). Jalankan pengujian yang mengubah state secara berurutan. Gunakan `--output` pada runner load untuk membuat direktori bukti baru, dan catat versi kode, konfigurasi, serta hasil aktual. Jangan mengubah bukti gagal menjadi lulus; perbaiki masalah dan jalankan pemeriksaan terkait kembali.

Transkrip disimpan sebagai UTF-8 tanpa kode warna ANSI. Spasi di akhir baris dan baris kosong berlebih pada akhir berkas dinormalisasi tanpa mengubah hasil pengujian.
