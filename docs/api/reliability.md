# Batas payload dan pemulihan kegagalan

Dokumen ini melengkapi [envelope event](hazard-event.md), [penyimpanan](storage.md), dan [consumer](consumers.md). Ukuran dihitung dalam byte, dengan 1 MiB sebesar 1.048.576 byte.

## Batas yang diterapkan

| Jalur | Batas | Perilaku |
| --- | --- | --- |
| Envelope event kanonik | 4 MiB, inklusif | Dihitung setelah serialisasi JSON UTF-8 lengkap, termasuk metadata dan attributes. |
| Producer dan topic Kafka | 5 MiB per batch | Menyediakan ruang untuk key, header, dan overhead protokol di luar envelope. |
| Fetch consumer dan replikasi broker | 6 MiB | Lebih besar dari batas batch. |
| Respons HTTP sumber dan API kanonik | 8 MiB | Sumber yang melampaui batas ditolak sebagai respons gagal; checkpoint tidak maju. Halaman kanonik dibatasi berdasarkan jumlah record dan byte. |
| Payload halaman view dashboard dan pemda | 7 MiB | Cursor melanjutkan dari record terakhir yang benar-benar dikembalikan; metadata respons mempunyai ruang tersendiri. |
| Pembaca client-api, CLI, dan helper demo | 8 MiB | Menampung halaman kanonik dan satu event terbesar yang didukung. |

Batas transport tidak dibuat identik karena objek yang dibatasi berbeda. Batas aplikasi 4 MiB menjadi kontrak bersama ingest dan consumer. Konfigurasi Kafka diterapkan pada topic baru maupun topic yang sudah ada oleh `kafka-init`.

Pengukuran sebelum perubahan pada stack lokal berisi 1.522 outbox dengan ukuran representasi JSONB 709 sampai 1.018 byte dan p95 1.018 byte. Angka ini merupakan pengamatan dataset demo, bukan dasar untuk menyatakan semua input akan kecil. Pengujian tambahan memakai envelope tepat 4 MiB, jauh di atas data demo tersebut. Tidak ada klaim dukungan ukuran tak terbatas atau latency tetap untuk payload besar.

## Event yang terlalu besar

Pada ingest, envelope yang melebihi 4 MiB membatalkan transaksi record tersebut, termasuk perubahan korelasi warning. Payload sumber utuh disimpan di `quarantine` dengan alasan `event_envelope_too_large`. Record berikutnya tetap diproses. Jumlah penolakan masuk statistik polling dan status sumber; tidak ada pemotongan attributes secara diam-diam.

Relay juga memeriksa outbox lama atau hasil penulisan administratif. Kegagalan permanen akibat ukuran disimpan pada `outbox.rejected_at` dan `rejection_reason`, disertai log `outbox_rejected`. Row tetap ada dengan `published_at` kosong. Relay hanya melanjutkan setelah penolakan tersimpan. Jika penyimpanan penolakan gagal, relay berhenti pada row itu dan mencoba lagi. Pembersihan outbox published tidak menghapus row yang ditolak.

Penolakan permanen berarti event tersebut **belum terkirim**. Snapshot hazard versi berikutnya dapat tetap terkirim, sehingga consumer mungkin melihat lompatan versi. Pemulihan row yang ditolak memerlukan pemeriksaan operator dan penyelesaian penyebab kegagalan; belum ada redrive otomatis. Kegagalan jaringan atau broker sementara tetap mempertahankan row pending dan urutan retry.

Jika seluruh respons sumber melampaui 8 MiB, tidak ada pemrosesan parsial dari body terpotong. Sumber ditandai bermasalah dan dicoba kembali dengan checkpoint lama. Kontrak mock belum mempunyai pagination sumber, sehingga respons yang terus melampaui batas memerlukan perubahan kontrak sumber atau konfigurasi dan pengujian kapasitas lanjutan. Ini batas yang berbeda dari pemulihan banyak record dalam respons yang masih muat.

## Transaksi ingest dan retry consumer

Hazard, korelasi, dan outbox untuk satu record disimpan dalam satu transaksi. Setiap transaksi tetap memiliki timeout lokal 2 detik. Checkpoint respons ditulis terpisah setelah semua record berhasil disimpan atau dikarantina. Bila record di tengah gagal, prefix yang sudah commit tetap tersedia, sedangkan checkpoint tidak maju. Polling berikutnya mengulang respons; hash mencegah outbox ganda untuk isi yang sama. Ini memungkinkan backlog diproses melampaui total 2 detik tanpa menghilangkan timeout per transaksi.

Consumer memisahkan event yang melanggar kontrak dari kegagalan operasional. Event tidak valid dikirim ke DLQ; offset asal baru di-commit setelah ACK DLQ. Event valid yang gagal disimpan setelah retry tetap mempertahankan offset. Worker mengembalikan kegagalan dan dijalankan ulang oleh kebijakan restart Compose. Bila kegagalan storage menetap, kemajuan group tersebut tertahan sampai storage pulih. Tidak ada jaminan bahwa retry mengatasi kegagalan permanen penyimpanan.

## Freshness dan trace

Respons list, detail, dan raw detail menyertakan `sources`. Metadata status ini tetap tersedia ketika pengguna memilih sebagian field hazard. Proyeksi Media tetap menolak attributes, koordinat, dan ID sumber. Metadata sumber disaring melalui tipe yang hanya memuat `source`, `status`, dan `stale_since`.

Panggilan HTTP sumber, PostgreSQL, Redis, dan Kafka mencatat `latency_ms`, identitas operasi, serta correlation ID. Trace PostgreSQL tidak menulis SQL, parameter, atau pesan error driver; Redis hanya menulis nama perintah. Publish, DLQ, dan commit offset menggunakan correlation ID record. Permintaan protokol Kafka seperti fetch, heartbeat, atau permintaan yang mencakup beberapa record memiliki ID operasi tersendiri, bukan klaim satu correlation ID event untuk seluruh request jaringan.

## Pengujian yang dapat diulang

- `scripts/check/check.ps1` atau `scripts/check/check.sh` menjalankan unit test dan vet, termasuk batas decode dan pemulihan saat pool SQLite sungguhan habis.
- `docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test` menguji batas 4 MiB dikurangi satu byte, tepat 4 MiB, dan 4 MiB ditambah satu byte, pagination byte, serta 600 record dengan kegagalan di tengah dan replay menggunakan PostgreSQL.
- `py scripts/demo/demo.py verify all` menjalankan regresi stack. `verify p5` juga menjalankan fixture outbox administratif 4 MiB melalui relay, Kafka, tiga consumer, dan API; fixture melebihi batas diperiksa tersimpan sebagai penolakan sebelum record berikutnya terkirim. Ini bukan klaim mock menghasilkan payload besar secara alami.
- `py scripts/loadtest/run.py --output docs/evidence/<nama-run>/load` mengukur ulang beban. Skenario outage memeriksa seismic sukses dan sumber BMKG sehat secara terpisah dari respons volcanic yang basi atau unavailable.

Jalankan suite yang mengubah keadaan stack secara berurutan. Fixture integrasi menghapus row hazard dan outbox uji setelah selesai, tetapi record sintetis historis dapat tetap berada di Kafka dan view consumer sampai volume demo direset.
