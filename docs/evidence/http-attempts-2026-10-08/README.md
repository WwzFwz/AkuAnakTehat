# Verifikasi latency per percobaan HTTP

Pengujian lokal 8 Oktober 2026 untuk Client API. Cakupan perubahan adalah log outbound HTTP ke Aggregator dan pengujian regresi dalam module Client API.

## Perintah dan hasil

Dijalankan dari `services/client-api`.

| Perintah | Hasil | Bukti |
| --- | --- | --- |
| `go test ./... -count=1 -v` | Lulus, exit code 0 | [tests.txt](tests.txt) |
| `go vet ./...` | Lulus, exit code 0 tanpa diagnostik | [vet.txt](vet.txt) |

`TestRetryLogsEachAttempt` memakai transport terkontrol yang menggagalkan panggilan pertama lalu mengembalikan respons valid pada panggilan kedua. Tes memeriksa tepat dua log, nomor percobaan 1 dan 2, correlation ID yang sama, status 0 lalu 200, hasil `transport_error` lalu `success`, serta latency nonnegatif. Kredensial dan correlation ID diteruskan ke kedua panggilan dengan deadline yang sama. Log diperiksa agar tidak memuat key, URL, nilai query, atau pesan error transport mentah.

`TestAttemptFailuresPreserveRetryPolicy` memeriksa batas dua percobaan, pembatalan tanpa retry tambahan, dan tidak adanya retry atas status HTTP atau JSON invalid. `TestAttemptLogsTimeoutAndClosesBody` memeriksa timeout, hasil sukses satu percobaan, dan penutupan body respons. Tes module lainnya memeriksa perilaku API, scope, autentikasi, pagination, serta pembatasan beban.

## Definisi dan cakupan

Satu percobaan adalah satu pemanggilan `HTTP.Do` oleh adapter. Timer baru mencakup panggilan tersebut hingga pembacaan, decode, dan penutupan body selesai. Status 0 berarti tidak ada respons HTTP. Hasil memakai kategori tetap tanpa isi galat atau respons upstream. Retry internal `net/http` tidak dihitung sebagai panggilan adapter terpisah.

Transkrip disimpan sebagai UTF-8 dengan akhir baris LF. `vet.txt` kosong karena perintah tidak menghasilkan diagnostik. Pemeriksaan ini menutup temuan G1 tentang penggabungan latency retry. Regresi Docker dan load test tidak dijalankan ulang pada perubahan ini; angka performa tetap mengacu pada [hasil load yang sudah direkam](../reliability-2026-10-08/README.md).
