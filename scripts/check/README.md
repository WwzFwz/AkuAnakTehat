# Pemeriksaan fondasi

`check.ps1` / `check.sh`: test bootstrap, `go test ./...`, dan `go vet ./...` pada delapan module service serta module `tools/field-cli`, masing-masing dengan `GOWORK=off`. Tidak memerlukan stack aktif.

`foundation_test.go`: suite integrasi terpisah, memakai Go standard library, HTTP lokal, dan Docker CLI. Jalankan setelah bootstrap serta `docker compose up -d --build --wait --wait-timeout 180`:

```text
go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m
```

Alternatif POSIX: `make smoke`. Gunakan konfigurasi bootstrap default (port 8080/8081/8082/8090, TTL access 60s, generator 10s). Suite memverifikasi readiness client-api terhadap query Aggregator yang sudah aktif.

Suite mengubah state simulasi PVMBG lalu memulihkannya, membuat tabel/topic uji unik lalu menghapusnya, dan stop/start atau restart PostgreSQL, Redis, Kafka, serta auth-service. Jalankan tanpa demo lain yang bersamaan. Tidak menghapus volume. Jika proses dihentikan paksa, cleanup mungkin tidak berjalan; periksa state simulasi dan fixture berprefix `foundation_smoke_` / `foundation-smoke-` sebelum mengulang.

Kredensial dibaca dari file lokal hasil bootstrap dan tidak dicetak. Test negatif JWT memakai private key lokal untuk membuat claim invalid dengan signature yang sah. Sesi login uji kedaluwarsa mengikuti TTL normal Redis. Hasil, cakupan, dan keterbatasan: [verifikasi fondasi](../../docs/evidence/foundation/README.md).

## Jalur ingest A

`make ingest-check` menjalankan dua lapis verifikasi berikut:

```text
docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test
go test ./scripts/check/foundation_test.go ./scripts/check/ingest_test.go -run TestIngestPipeline -v -count=1 -timeout=5m
```

Test PostgreSQL membuat schema unik `ingest_test_<timestamp>` dan membersihkannya; tidak mengubah tabel produksi. Test live memakai mock asli, menyalakan schema v2 dan outage hang sementara, kemudian memulihkan state dan me-restart Aggregator. Data sintetis hasil ingest serta outbox dipertahankan. Jangan jalankan bersamaan dengan demo/test lain yang mengubah mock. [Hasil ingest](../../docs/evidence/ingest/README.md).

## Jalur query dan event

`make events-check` memeriksa relay, replay, dedup, DLQ, consumer independen, subscriber baru, serta outage/recovery Kafka. `make query-check` memeriksa filter/cursor, pembatasan Media, raw/proyeksi, trace, drift skema, stale/recovery, expiry token nyata dan refresh CLI, serta rebuild notifier tanpa restart service lain.

PowerShell, setelah semua image tersedia:

```powershell
docker compose --profile demo up -d --build --wait --wait-timeout 180
$env:GOWORK='off'
go test ./scripts/check/foundation_test.go ./scripts/check/ingest_test.go ./scripts/check/events_test.go ./scripts/check/query_test.go ./scripts/check/large_events_test.go -v -count=1 -timeout=15m
py scripts/loadtest/run.py
```

Suite query menggunakan TTL bootstrap 60s dan menunggu 65s dalam satu sesi CLI. Jangan memperpendek TTL atau mengganti jam untuk bukti expiry alami. Suite juga stop/build/start notifier; sediakan akses Docker dan source lengkap.

`large_events_test.go` membuat fixture outbox administratif untuk menguji envelope 4 MiB melalui relay, broker, tiga consumer, dan API. Fixture 4 MiB ditambah satu byte harus tetap tersimpan sebagai penolakan, sedangkan event berikutnya tetap terkirim. Pengujian batas saat ingest, pagination byte, dan backlog 600 record berada dalam suite PostgreSQL. `py scripts/demo/demo.py verify all` juga memasukkan pengujian event besar; `make events-check` hanya menjalankan suite event dasarnya.

`make load-check` menjalankan runner Python/k6, mengukur 50 koneksi TCP selama minimal 60s, dan memulihkan konfigurasi PVMBG setelah pengujian. Detail metrik di [loadtest](../loadtest/README.md), hasil di [integrasi](../../docs/evidence/integration/README.md). Jalankan seluruh suite secara berurutan, tanpa demo lain yang mengubah state.
