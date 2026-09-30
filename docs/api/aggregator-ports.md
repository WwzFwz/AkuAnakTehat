# Port internal Aggregator — kontrak baseline

**Status:** kontrak v1 untuk tahap jalur inti. Signature di bawah adalah rancangan dalam dokumentasi, belum file deklarasi Go atau implementasi repository.

Port adalah interface kecil di package pemakainya. Adapter PostgreSQL memenuhi port ingest, query, dan outbox; tidak ada import business logic antarservice.

## A — UnitOfWork dan Tx

Pemilik interface: `application/ingest`.

`WithTx(ctx, callback)` menyediakan Tx yang terikat ke satu transaksi database. Callback mendapatkan operasi untuk:

- Membaca hazard berdasarkan `(source, source_ref_id)`, termasuk metadata version/hash.
- Menyimpan hazard baru/perubahan dan timestamp yang sesuai.
- Menyimpan warning dan membaca semua warning terkait gempa.
- Menambahkan satu snapshot payload outbox untuk perubahan bisnis.
- Mencatat karantina per record beserta alasan dan correlation ID.
- Memajukan checkpoint endpoint setelah hasil batch berhasil ditangani.

Semua penulisan terkait memakai Tx yang sama. Callback error membatalkan transaksi; error commit diteruskan ke pemanggil. Replay input sama tidak menambah outbox bila hash sama. Keputusan ID/version/hash dibuat konsisten dengan unique constraint dan transaksi.

Checkpoint pembacaan dan status sumber boleh memiliki port terpisah. Kegagalan satu endpoint BMKG tidak memajukan checkpoint endpoint tersebut atau membuang hasil valid endpoint lainnya.

Signature yang menjadi acuan A/B/C:

```go
type UnitOfWork interface {
    WithTx(ctx context.Context, fn func(Tx) error) error
}

type Tx interface {
    FindHazard(ctx context.Context, source, sourceRefID string) (StoredHazard, bool, error)
    PutHazard(ctx context.Context, record StoredHazard) error
    PutWarning(ctx context.Context, warning Warning) error
    WarningsFor(ctx context.Context, relatedEventID string) ([]Warning, error)
    AppendOutbox(ctx context.Context, event OutboxEvent) error
    SaveCheckpoint(ctx context.Context, endpoint string, watermark time.Time) error
    Quarantine(ctx context.Context, rejected RejectedRecord) error
}

type CheckpointReader interface {
    ReadCheckpoint(ctx context.Context, endpoint string) (time.Time, bool, error)
}
```

- `StoredHazard`: HazardEvent, version int64, content_hash []byte, updated_at dan last_seen_at UTC.
- `Warning`: field kontrak TsunamiWarning serta unknown fields yang diperlukan untuk attributes.
- `OutboxEvent`: event_id UUID, hazard_id UUID, version int64, payload JSON envelope lengkap, created_at UTC. Correlation ID berada dalam payload.
- `RejectedRecord`: source, endpoint, payload JSON, reason, correlation_id, observed_at UTC.
- bool dari FindHazard/ReadCheckpoint berarti record ditemukan. Tidak ditemukan bukan error transport/DB.
- PutHazard tidak menaikkan version diam-diam. Application menghitung perubahan dalam Tx; adapter menegakkan constraint dan konsistensi transaksi.
- Nama checkpoint tetap: `bmkg.seismic-events`, `bmkg.tsunami-warnings`, `pvmbg.volcanic-reports`.

## B — HazardQuery

Pemilik interface: `application/query`.

- `List(ctx, filter)` menghasilkan halaman hazard, cursor berikutnya, dan metadata sumber.
- `Get(ctx, hazardID)` menghasilkan satu HazardEvent atau error not-found.

Filter dan representasi wire mengikuti [HTTP internal](aggregator-http.md). Pool baca memiliki batas koneksi dan timeout sendiri. Query tidak melakukan polling sumber.

```go
type HazardQuery interface {
    List(ctx context.Context, filter HazardFilter) (HazardPage, error)
    Get(ctx context.Context, hazardID string) (HazardEvent, error)
}
```

`HazardFilter` memuat type/severity opsional, since opsional, limit, dan cursor. `HazardPage` memuat data, next_cursor, dan sources mengikuti kontrak HTTP. Get yang tidak menemukan ID mengembalikan error terklasifikasi not-found; tidak mengembalikan objek kosong sebagai sukses.

## C — OutboxStore dan Publisher

Pemilik interface: `worker/outbox`.

- `Pending(ctx, limit)` membaca row belum di-ACK, urut id.
- `MarkPublished(ctx, id, acknowledgedAt)` menandai row setelah ACK Kafka.
- `DeletePublishedBefore(ctx, cutoff)` hanya menghapus row yang sudah di-ACK.
- `Publisher.Publish(ctx, message)` selesai sukses setelah broker ACK; error mempertahankan row sebagai pending.

A menulis row outbox dalam transaksi ingest. C membaca, mengirim, dan menandainya. Tidak ada transaksi database yang ditahan selama menunggu jaringan Kafka. Polling satu detik adalah jalur inti; NOTIFY hanya opsi tambahan.

```go
type OutboxStore interface {
    Pending(ctx context.Context, limit int) ([]OutboxMessage, error)
    MarkPublished(ctx context.Context, id int64, acknowledgedAt time.Time) error
    DeletePublishedBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

type Publisher interface {
    Publish(ctx context.Context, message OutboxMessage) error
}
```

`OutboxMessage` memuat id int64, event_id, hazard_id, version int64, payload JSON lengkap, dan correlation_id hasil pembacaan envelope. Pending mengembalikan slice kosong bila tidak ada kerja, bukan error. MarkPublished idempoten untuk row yang telah ditandai; ID yang tidak dikenal merupakan error. DeletePublishedBefore mengembalikan jumlah row terhapus.

## Error dan kepemilikan file

Pisahkan invalid input, not-found, conflict/retryable, timeout/cancellation, dan dependency unavailable. Adapter tidak membocorkan error driver mentah ke HTTP.

Di `adapter/outbound/postgres`, A memegang unit of work/ingest, B query, dan C outbox. Main/config/observability/Dockerfile Aggregator dipegang A. Perubahan signature atau semantik error pada kontrak v1 perlu diselaraskan dengan semua pemakai sebelum kode mereka diubah.
