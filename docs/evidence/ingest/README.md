# Verifikasi ingest Aggregator A — 1 Oktober 2026

Jalur yang diuji: mock BMKG/PVMBG → polling independen → tolerant reader → pemetaan/korelasi → transaksi PostgreSQL → snapshot outbox. Query B, relay Kafka, dan consumer C belum diimplementasikan.

## Hasil lokal

| Pemeriksaan | Hasil |
| --- | --- |
| Module | `go test ./...` dan `go vet ./...` lulus; test database dilewati pada run unit dan dijalankan terpisah di bawah. |
| Pemetaan | Ambang magnitude 5.0/6.5, warning mengoverride magnitude, warning maksimum, ordering warning deterministik, unknown JSON/angka besar tetap utuh. |
| Hash | Metadata observasi dan format angka ekuivalen tidak mengubah hash; perubahan bisnis mengubah hash. Timestamp kanonik memakai mikrodetik PostgreSQL. |
| Transport/poller | Timeout request hang, batas respons, header credential/correlation, overlap checkpoint, keberhasilan sebagian BMKG, serta open/probe/recovery breaker lulus unit test. |
| Transaksi nyata | `TestIngestPostgres` lulus dalam 1,55s pada schema sementara. Warning sebelum gempa, eskalasi terlambat, replay tanpa outbox tambahan, snapshot versi, karantina record invalid, dan health per endpoint terbukti. |
| Rollback | Constraint sengaja menggagalkan checkpoint setelah penulisan hazard/outbox; semua perubahan rollback dan watermark tetap. Batas waktu transaksi/repository serta rollback callback juga lulus. |
| Nilai invalid | Angka dengan exponent ekstrem dan escape NUL dikarantina sebagai raw text dalam JSONB sehingga tidak menggagalkan record valid. |
| Live ingest | `TestIngestPipeline` lulus: schema v2 mencapai JSONB tanpa migrasi, PVMBG hang mencapai tiga kegagalan, BMKG terus maju, watermark PVMBG tidak maju saat gagal, lalu pulih otomatis. |
| Restart | Restart Aggregator mempertahankan identitas hazard, versi migrasi, dan checkpoint persisten; polling dilanjutkan. |
| Regresi fondasi | Tujuh test fondasi tetap lulus setelah Aggregator ditambahkan (29,423s). |

Output: [live ingest](runtime-test-output.txt), [transaksi](postgres-test-output.txt), [regresi fondasi](foundation-regression.txt). Tidak ada credential atau payload sensitif dalam output ini.

## Mengulang

Dari root dengan konfigurasi bootstrap default dan Docker Compose aktif:

```text
docker compose up -d --build --wait --wait-timeout 180
docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test
go test ./scripts/check/foundation_test.go ./scripts/check/ingest_test.go -run TestIngestPipeline -v -count=1 -timeout=5m
```

Alternatif POSIX: `make up`, lalu `make ingest-check`. Test live mengubah mock sementara dan me-restart Aggregator; jangan jalankan saat demo lain. Test transaksi hanya membuat/menghapus schema berprefix `ingest_test_`, tanpa menghapus volume. `--env-from-file` mengatasi Compose lokal yang tidak meneruskan env_file service pada `run`.

## Keputusan implementasi

- Decoder typed/tolerant berada pada `canonicalize/input.go`; adapter sumber hanya transport + pemilihan decoder, sehingga tidak ada import cycle.
- `source_endpoint_status` ditambahkan untuk status BMKG yang mempunyai dua endpoint. Satu endpoint berhasil tidak boleh membuat sumber terlihat sepenuhnya sehat.
- `Tx.FindWarning` ditambahkan untuk menolak perubahan related_event_id atas warning_id yang sama.
- Quarantine menyimpan teks asli dalam `payload.raw_json`. Batas sumber 8 MiB; envelope rusak atau respons terlalu besar tidak memajukan checkpoint.
- Fetch di luar transaksi; transaksi serializable memakai pool terbatas dan timeout lokal. Konflik/commit gagal diulang melalui polling berikutnya, tanpa memajukan watermark.
- `/ready/ingest` memeriksa DB; `/ready` dan route query masih 503 agar client-api tidak menganggap B sudah selesai.

## Batas

Satu instance Aggregator; belum ada leader coordination multi-replica. Watermark memakai waktu mulai request dan overlap 10s, mengasumsikan jam sumber selaras serta semantik since warning sebagai waktu perubahan. Backfill lebih tua daripada overlap memerlukan kebijakan tersendiri. Outbox pending bertambah sampai relay C tersedia; tidak ada publish Kafka dalam jalur A. Race/load test, SLO ingest lag, crash/power-loss, dan P1–P5 lengkap belum dibuktikan. Status run CI GitHub masih belum dapat dikonfirmasi dari sesi ini.
