# Pemeriksaan fondasi

`check.ps1` / `check.sh`: test bootstrap, `go test ./...`, dan `go vet ./...` setiap module yang telah diimplementasikan. Tidak memerlukan stack aktif.

`foundation_test.go`: suite integrasi terpisah, memakai Go standard library, HTTP lokal, dan Docker CLI. Jalankan setelah bootstrap serta `docker compose up -d --build --wait --wait-timeout 180`:

```text
go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m
```

Alternatif POSIX: `make smoke`. Gunakan konfigurasi bootstrap default (port 8080/8081/8082/8090, TTL access 60s, generator 10s). Suite saat ini mengharapkan Aggregator belum ada.

Suite mengubah state simulasi PVMBG lalu memulihkannya, membuat tabel/topic uji unik lalu menghapusnya, dan stop/start atau restart PostgreSQL, Redis, Kafka, serta auth-service. Jalankan tanpa demo lain yang bersamaan. Tidak menghapus volume. Jika proses dihentikan paksa, cleanup mungkin tidak berjalan; periksa state simulasi dan fixture berprefix `foundation_smoke_` / `foundation-smoke-` sebelum mengulang.

Kredensial dibaca dari file lokal hasil bootstrap dan tidak dicetak. Test negatif JWT memakai private key lokal untuk membuat claim invalid dengan signature yang sah. Sesi login uji kedaluwarsa mengikuti TTL normal Redis. Hasil, cakupan, dan keterbatasan: [verifikasi fondasi](../../docs/evidence/foundation/README.md).
