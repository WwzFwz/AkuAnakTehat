# Penyimpanan — batas dan skema baseline

**Status:** DDL Canonical Store dan repository ingest sudah tersedia pada jalur A. SQLite consumer sudah tersedia pada masing-masing service. Migrasi SQL adalah sumber skema executable; dokumen ini merangkum kontraknya.

| Store | Pemilik akses langsung | Pemakai tidak langsung |
| --- | --- | --- |
| canonical-db / PostgreSQL | Aggregator | client-api melalui HTTP Aggregator |
| auth-store / Redis | auth-service | client melalui token endpoint |
| Kafka log dan offsets | Broker melalui protokol Kafka | Producer dan consumer melalui protokol broker |
| SQLite view/dedup | Masing-masing consumer | Endpoint view melalui service pemilik |
| Data mock in-memory | Masing-masing mock | Aggregator melalui HTTP |

## Tabel Canonical Store

| Tabel | Isi dan constraint utama |
| --- | --- |
| hazard_events | UUID hazard_id PK; unique(source,source_ref_id); field kanonik bertipe; attributes JSONB; version/content_hash/updated_at/last_seen_at internal; koordinat NOT NULL. |
| outbox | ID urut; event_id UUID unique; hazard_id/version; snapshot payload JSONB; created_at; published_at nullable untuk ACK. |
| source_status | Status, waktu sukses/percobaan terakhir, dan awal stale; sumber yang sedang down tetap dapat mempunyai data historis. |
| source_endpoint_status | State sehat, kegagalan berurutan, dan waktu polling masing-masing endpoint; mencegah keberhasilan satu endpoint menutupi kegagalan endpoint lain. |
| checkpoints | PK endpoint; watermark UTC, maju setelah seluruh record respons selesai; transaksi record terdahulu dapat sudah committed saat record berikutnya gagal. |
| tsunami_warnings | warning_id unik, related_event_id, seluruh data warning yang diperlukan untuk korelasi ulang. |
| quarantine | Payload invalid, sumber/endpoint, alasan, correlation_id, dan waktu pencatatan. |
| schema_migrations | Dikelola migration runner; bukan tabel yang dibuat ulang oleh service lain. |

Migrasi tersedia di `services/aggregator/migrations/001_initial.up.sql` dan `002_source_status_stale_since.up.sql`, dan `003_outbox_rejections.up.sql` beserta pasangan `.down.sql`, dijalankan dengan golang-migrate/iofs. Referensi `VOLCANO-DEMO-01`/`VOLCANO-DEMO-02` tersedia pada `reference/volcanoes.json`; nama/koordinat berlabel sintetis.

## Detail field untuk kontrak repository

Semua timestamp menggunakan TIMESTAMPTZ/UTC. Nama SQL menggunakan snake_case. Ini adalah acuan migrasi pertama; setelah DDL dibuat, perubahan berikutnya dilacak melalui migrasi dan dokumentasi ini diperbarui.

Waktu kanonik dinormalisasi ke mikrodetik agar hash stabil setelah round-trip PostgreSQL. `quarantine.payload` berisi `{"raw_json":"<teks record asli>"}` supaya record dengan angka/escape yang tidak dapat direpresentasikan JSONB tetap dapat dikarantina. `outbox` menegakkan unique(hazard_id,version) dan FK ke hazard.

| Tabel | Field kontrak |
| --- | --- |
| hazard_events | hazard_id UUID PK; source TEXT; source_ref_id TEXT; hazard_type TEXT; severity TEXT; area_name TEXT; latitude/longitude DOUBLE PRECISION; occurred_at/ingested_at/updated_at/last_seen_at TIMESTAMPTZ; attributes JSONB default `{}`; version BIGINT default 1; content_hash BYTEA. Semua NOT NULL. |
| outbox | id BIGSERIAL PK; event_id UUID UNIQUE; hazard_id UUID; version BIGINT; payload JSONB; created_at TIMESTAMPTZ; published_at TIMESTAMPTZ NULL; rejected_at TIMESTAMPTZ NULL; rejection_reason TEXT NULL. Ketiganya nullable; kolom lain NOT NULL. |
| checkpoints | endpoint TEXT PK; watermark TIMESTAMPTZ NOT NULL. |
| tsunami_warnings | warning_id TEXT PK; related_event_id TEXT NOT NULL; payload JSONB NOT NULL; updated_at TIMESTAMPTZ NOT NULL. Payload mempertahankan seluruh field warning dan unknown fields. |
| quarantine | id BIGSERIAL PK; source/endpoint/reason/correlation_id TEXT NOT NULL; payload JSONB NOT NULL; observed_at TIMESTAMPTZ NOT NULL. |
| source_status | source TEXT PK; status TEXT NOT NULL; last_success_at TIMESTAMPTZ NULL; last_attempt_at TIMESTAMPTZ NULL; stale_since TIMESTAMPTZ NULL; consecutive_failures INTEGER NOT NULL default 0; last_error TEXT NULL. |

Constraint: source hanya BMKG/PVMBG; hazard_type hanya SEISMIC/VOLCANIC; severity hanya NORMAL/WASPADA/SIAGA/AWAS; version positif; unique(source,source_ref_id). Warning boleh mendahului gempa sehingga related_event_id tidak memakai FK yang mewajibkan hazard sudah tersedia.

Indeks awal: `(hazard_type, occurred_at DESC, hazard_id DESC)` untuk daftar per tipe, `(occurred_at DESC, hazard_id DESC)` untuk daftar gabungan, `(related_event_id)` pada warning, dan indeks pending outbox berdasarkan id dengan syarat published_at IS NULL AND rejected_at IS NULL. Tambahan indeks mengikuti hasil query nyata.

`schema_observations` merupakan tambahan opsional. Kemunculan field sumber seperti confidence_level tidak memerlukan DDL; nilainya masuk attributes JSONB.

## Redis dan volume

Auth-store menyimpan hash refresh token, metadata client/scope/family, status used/revoked, dan expiry. Script rotasi memeriksa state lama sebelum mutasi. Private key tidak disimpan di Redis. Detail aktual ada pada README adapter auth-service.

Volume per store dipertahankan pada restart biasa. Menghapus volume menghapus state; tindakan itu tidak boleh dipakai sebagai cara mendemokan pemulihan restart. Tidak ada service yang mendapat kredensial store milik service lain.
