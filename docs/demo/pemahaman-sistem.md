# Memahami sistem sebelum demo

Dokumen ini membantu anggota menjelaskan implementasi dengan kata-kata sendiri. Baca bersama [diagram dan README utama](../../README.md), kemudian praktikkan [panduan demo](../../scripts/demo/README.md). Jawaban berikut menjelaskan kode yang tersedia, bukan hafalan untuk menggantikan pemahaman.

## Cerita sistem dalam satu menit

BMKG dan PVMBG menyediakan data dengan skema serta kredensial berbeda. Aggregator mengambil data secara berkala dan mengubahnya menjadi HazardEvent. Data kanonik disimpan bersama checkpoint dan outbox dalam PostgreSQL. Client membaca data tersimpan melalui Client API dengan hak akses yang diatur Auth Service. Perubahan juga dikirim ke Kafka agar dashboard, notifier, dan portal Pemda dapat memprosesnya secara independen.

Pemisahan tersebut membuat sumber yang lambat tidak langsung memperlambat pembacaan pengguna. Sebagai konsekuensi, data tidak selalu merupakan keadaan terbaru sensor. Respons membawa status sumber agar data basi tidak disamarkan sebagai data segar.

## Pertanyaan yang perlu bisa dijawab

| Pertanyaan | Penjelasan yang sesuai implementasi |
| --- | --- |
| Mengapa tidak langsung memanggil mock pada setiap request pengguna? | Kecepatan dan availability sumber berbeda. Membaca hasil materialisasi memisahkan latency pengguna dari HTTP sumber, dengan konsekuensi keterlambatan ingest. |
| Apa beda source_ref_id dan hazard_id? | Source reference ID berasal dari instansi. Hazard ID dibentuk BNPB dan dipertahankan lewat keunikan pasangan sumber dan ID sumber. |
| Apa yang terjadi jika warning datang sebelum gempa? | Warning disimpan, lalu dikorelasikan saat gempa terkait tersedia. Warning yang datang kemudian juga dapat memperbarui hazard. |
| Mengapa content hash dibutuhkan? | Polling berulang dapat mengambil record yang sama. Hanya perubahan makna konten yang meningkatkan version dan menghasilkan event baru. |
| Apa fungsi watermark dan overlap? | Checkpoint membatasi data yang diminta pada polling berikutnya. Overlap memberi toleransi keterlambatan terbatas; ini bukan jaminan semua data yang terlambat tanpa batas akan ditemukan. |
| Mengapa field baru tidak memerlukan restart? | Reader memisahkan field wajib dari atribut tambahan. JSONB menyimpan tambahan tanpa kolom DDL baru. Field wajib yang tidak valid tetap ditolak atau dikarantina. |
| Mengapa query dan ingest masih satu service? | Keduanya menggunakan data yang dimiliki Aggregator. Pemisahan pool dan worker cukup untuk lingkup M1 tanpa menambah kontrak jaringan dan pemilik transaksi baru. |
| Mengapa PostgreSQL, bukan MongoDB? | Transaksi hazard, checkpoint, dan outbox serta query kolom kanonik cocok dengan PostgreSQL. JSONB sudah memenuhi kebutuhan atribut fleksibel. MongoDB merupakan alternatif, tetapi tetap memerlukan desain konsistensi dan operasional. |
| Mengapa ada outbox? | Menulis database lalu publish secara terpisah dapat kehilangan event ketika proses crash. Niat publish disimpan bersama data sehingga relay dapat melanjutkan setelah pulih. |
| Apakah outbox menjamin tanpa duplikasi? | Tidak. ACK Kafka dapat berhasil sebelum tanda published tersimpan. Replay masih mungkin; consumer harus idempoten sesuai efek bisnisnya. |
| Mengapa setiap consumer memakai group berbeda? | Satu group membagi pekerjaan. Group berbeda diperlukan agar masing-masing consumer menerima stream lengkap. |
| Bagaimana consumer mengejar pesan setelah mati? | Group lama melanjutkan dari committed offset selama pesan masih dalam retensi Kafka. SQLite mempertahankan view dan marker dedup. |
| Mengapa subscriber baru perlu group baru dan store kosong saat diuji? | Agar replay histori terbukti berasal dari Kafka, bukan data sisa pengujian sebelumnya. |
| Apa fungsi DLQ? | Menyimpan pesan invalid atau yang gagal setelah batas percobaan beserta metadata asal. Offset asal baru di-commit setelah publish DLQ berhasil. |
| Mengapa notifier belum exactly-once? | Pengiriman dan marker SQLite tidak satu transaksi. Crash di antara keduanya dapat mengulang pengiriman. |
| Mengapa Media tidak boleh menerima raw lalu menyembunyikannya di UI? | Data yang sudah dikirim tetap dapat dibaca client. Server memeriksa scope dan membuat objek baru berdasarkan allowlist. |
| Mengapa JWT asimetris? | Client API hanya membutuhkan public key untuk verifikasi dan tidak mendapat kemampuan menerbitkan token. |
| Apakah refresh membuat JWT lama langsung tidak berlaku? | Tidak. Access JWT diverifikasi offline dan berlaku sampai expired. Refresh reuse mencabut keluarga refresh, bukan langsung semua access JWT aktif. |
| Apa akibat respons refresh hilang di jaringan? | Redis mungkin sudah menyelesaikan rotasi. Penggunaan ulang token lama dapat terdeteksi sebagai reuse dan memerlukan login ulang. |
| Apa beda health dan ready? | Health memeriksa proses hidup. Ready memeriksa dependensi yang relevan. Keduanya tidak otomatis membuktikan semua data sumber baru atau consumer tanpa lag. |
| Apakah 50 VU sama dengan 50 koneksi? | Tidak otomatis. Runner membaca koneksi TCP ESTABLISHED, lalu memeriksa minimal 50 koneksi bertahan selama minimal 60 detik. |
| Mengapa 429 tidak dihitung sebagai error bisnis? | Spesifikasi mengecualikan penolakan terkendali, tetapi jumlahnya wajib dilaporkan. Throughput sukses tetap dihitung hanya dari respons HTTP 200. |

## Tempat membaca kode

| Alur | Berkas awal |
| --- | --- |
| Perakitan Aggregator | `services/aggregator/cmd/aggregator/main.go` |
| Tolerant reader dan pemetaan | `services/aggregator/internal/application/canonicalize/input.go` dan `mapper.go` |
| Transaksi ingest | `services/aggregator/internal/application/ingest/service.go` dan adapter PostgreSQL |
| Polling dan circuit breaker | `services/aggregator/internal/worker/poller/` dan adapter sumber |
| Query | `services/aggregator/internal/application/query/` dan `internal/adapter/outbound/postgres/query_repository.go` |
| JWT dan proyeksi | `services/client-api/internal/authn/verifier.go` dan `internal/projection/projector.go` |
| Rotasi token | `services/auth-service/internal/adapter/outbound/redis/rotate.lua` |
| Relay | `services/aggregator/internal/worker/outbox/relay.go` |
| Efek notifier | `services/notifier/internal/application/apply.go` |
| Consumer, DLQ, dan commit | `services/notifier/internal/consumer/consumer.go` |
| Bukti runtime | `scripts/check/` serta `docs/evidence/` |

## Membaca hasil pengujian dengan benar

Pisahkan tiga hal ketika menjawab penguji. Kode menjelaskan mekanisme, pengujian membuktikan skenario yang memang dijalankan, dan asumsi menjelaskan kondisi tempat hasil itu berlaku. Jangan memperluas satu pengujian lokal menjadi klaim seluruh kegagalan atau semua lingkungan pasti aman.

Sistem ini berjalan pada satu host dengan satu broker. Data mock sintetis dan tersimpan dalam memori. Notifikasi berupa log. TLS, failover, banyak relay, micro-cache, singleflight, LISTEN/NOTIFY, serta propagasi deadline melalui header belum diimplementasikan. Timeout lokal, pembatasan beban, circuit breaker, dan polling outbox sudah tersedia.

Untuk latihan, masing-masing anggota sebaiknya bisa menunjuk satu request client, satu transaksi ingest, dan satu event Kafka pada diagram; menjelaskan apa yang terjadi jika dependensinya mati; lalu menunjukkan bukti yang sesuai. Kontribusi boleh berbeda, tetapi spesifikasi tetap meminta kelompok memahami kode yang dikumpulkan.
