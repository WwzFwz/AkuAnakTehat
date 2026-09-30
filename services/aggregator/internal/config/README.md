# config

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Konfigurasi ingest Aggregator; `config.go` adalah sumber nama dan nilai default yang diterapkan.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A sudah diimplementasikan. Berkas tersedia: `config.go`. Cakupan pengujian ada di [bukti ingest](../../../../docs/evidence/ingest/README.md). Tabel rencana di bawah adalah panduan pemecahan tanggung jawab; sebagian operasi digabung dalam file yang tersedia.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `config.go` | Struct konfigurasi dan pembacaan env/berkas lokal. |
| `validate.go` | Validasi invarian milik service sebelum worker dimulai. |

## Kontrak dan alur

- Load() menghasilkan konfigurasi tervalidasi; HTTP_ADDR memiliki port rencana 9000.
- BMKG_URL, PVMBG_URL: Alamat sumber; secret sumber terpisah.
- DATABASE_URL: Akses Canonical Store khusus Aggregator.
- BMKG_POLL_INTERVAL, PVMBG_POLL_INTERVAL: Default2 s/5 s; overlap10 s.
- BMKG_TIMEOUT, PVMBG_TIMEOUT: Default1 s/4 s; timeout DB lokal.
- BMKG_API_KEY dan PVMBG_TOKEN: kredensial wajib sumber, terpisah.
- POLL_OVERLAP=10s, DB_TIMEOUT=2s, DB_POOL_SIZE=5 (2–20).
- BREAKER_FAILURES=3 dan BREAKER_COOLDOWN=10s, per endpoint; probe berikutnya pulih tanpa restart.
- INTERNAL_KEY dan KAFKA_* yang sudah disiapkan bootstrap baru dipakai saat query/relay B/C dibuat; belum dibaca jalur A.
- Fitur tambahan schema_observations dan LISTEN/NOTIFY belum diimplementasikan.

## Dependensi

- Environment variable dan file lokal yang tidak ter-commit; diteruskan main ke komponen.

## Aturan penting

- Tidak membaca env/file konfigurasi service lain.
- Tidak memakai secret default dan tidak mencetak konfigurasi sensitif.
- Validasi durasi positif, limit, TTL, dan path yang relevan; konfigurasi antarservice dicatat sebagai kontrak deployment.

## Langkah implementasi dan verifikasi

- Finalisasi nama env di .env.example saat file itu dibuat.
- Pisahkan error konfigurasi dari dependensi yang sementara unavailable.
