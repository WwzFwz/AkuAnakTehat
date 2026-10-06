# Trace correlation ID

Dari root repository, jalankan `sh scripts/trace/trace.sh <id>` (POSIX) atau
`powershell -NoProfile -File scripts/trace/trace.ps1 <id>` (Windows). Bisa juga
langsung `py scripts/trace/trace.py <id> --tail 5000`.

Membutuhkan Python 3 dan Docker Compose. Script membaca log JSON, mencocokkan
field `correlation_id` secara persis, lalu mengurutkan hasil menurut timestamp.
Default 2.000 baris terakhir per container; exit 1 bila tidak ada kecocokan.
Tidak mengubah container atau konfigurasi, dan tidak mengevaluasi ID sebagai shell.

Jalur query dapat ditelusuri dari client-api ke Aggregator. Jalur event memakai
ID batch poller yang dibawa ke outbox dan consumer. Ini dua alur berbeda; query
bukan penyebab ingest. `record_completed` berarti pemrosesan/penanganan DLQ dan
commit offset selesai, bukan selalu sebuah notifikasi baru terkirim.

Jangan menambahkan token, secret, atau header autentikasi ke log aplikasi.
Contoh hasil aktual ada di [bukti integrasi](../../docs/evidence/integration/README.md).
