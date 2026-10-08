# Port internal Aggregator — kontrak baseline

**Status:** port UnitOfWork/Tx, checkpoint, dan status sudah diimplementasikan untuk jalur A. Store/Publisher pada worker/outbox sudah diimplementasikan untuk C; HazardQuery, repository baca, dan HTTP inbound sudah diimplementasikan untuk tahap B.

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

Hazard, warning terkait, dan outbox satu record memakai Tx yang sama. Record yang ditolak dicatat pada transaksi karantina. Checkpoint disimpan pada transaksi terakhir setelah seluruh respons ditangani. Callback error membatalkan transaksi record tersebut; record sebelumnya tetap committed dan aman dipoll ulang. Error commit diteruskan ke pemanggil. Replay input sama tidak menambah outbox bila hash sama. Keputusan ID/version/hash dibuat konsisten dengan unique constraint dan transaksi.

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
    FindWarning(ctx context.Context, warningID string) (Warning, bool, error)
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
- Nama konkret Go: `hazard.Event`, `hazard.Record`, `tsunami.Warning`; tipe payload outbox/karantina dan interface berada di `application/ingest/ports.go`.
- FindWarning menjaga related_event_id immutable untuk warning_id yang sama; perubahan relasi dikarantina agar hazard lama tidak kehilangan korelasi diam-diam.
- `Warning`: field kontrak TsunamiWarning serta unknown fields yang diperlukan untuk attributes.
- `OutboxEvent`: event_id UUID, hazard_id UUID, version int64, payload JSON envelope lengkap, created_at UTC. Correlation ID berada dalam payload.
- `RejectedRecord`: source, endpoint, payload JSON, reason, correlation_id, observed_at UTC.
- bool dari FindHazard/ReadCheckpoint berarti record ditemukan. Tidak ditemukan bukan error transport/DB.
- PutHazard tidak menaikkan version diam-diam. Application menghitung perubahan dalam Tx; adapter menegakkan constraint dan konsistensi transaksi.
- Nama checkpoint tetap: `bmkg.seismic-events`, `bmkg.tsunami-warnings`, `pvmbg.volcanic-reports`.

`StatusStore.RecordPoll(ctx, endpoint, ok, degraded, code, attemptedAt)` mencatat hasil endpoint dan menghitung status sumber. Source DEGRADED bila hanya sebagian endpoint sehat atau terdapat record karantina; DOWN bila seluruh endpoint gagal. HTTP sukses tetapi commit gagal tidak memajukan checkpoint. Breaker menghitung kegagalan fetch sumber; kegagalan DB tetap dicoba pada siklus berikutnya.

## B — HazardQuery

Pemilik interface: `application/query`.

- `List(ctx, filter)` menghasilkan halaman hazard, cursor berikutnya, dan metadata sumber.
- `Get(ctx, hazardID)` menghasilkan HazardDetail berisi HazardEvent dan metadata sources atau error not-found.

Filter dan representasi wire mengikuti [HTTP internal](aggregator-http.md). Pool baca memiliki batas koneksi dan timeout sendiri. Query tidak melakukan polling sumber.

```go
type HazardQuery interface {
    List(ctx context.Context, filter HazardFilter) (HazardPage, error)
    Get(ctx context.Context, hazardID string) (HazardDetail, error)
}
```

`HazardFilter` memuat type/severity opsional, since opsional, limit, dan cursor. `HazardPage` memuat data, next_cursor, dan sources mengikuti kontrak HTTP. Get yang tidak menemukan ID mengembalikan error terklasifikasi not-found; tidak mengembalikan objek kosong sebagai sukses.

## C — OutboxStore dan Publisher

Pemilik interface: `worker/outbox`.

- `Pending(ctx, limit)` membaca row yang belum di-ACK dan belum ditolak permanen, urut id.
- `MarkPublished(ctx, id, acknowledgedAt)` menandai row setelah ACK Kafka.
- `Reject(ctx, id, reason)` menyimpan penolakan permanen tanpa menandai row sebagai published. Relay baru boleh melanjutkan setelah penolakan tersimpan.
- `DeletePublishedBefore(ctx, cutoff)` hanya menghapus row yang sudah di-ACK.
- `Publisher.Publish(ctx, message)` selesai sukses setelah broker ACK; error mempertahankan row sebagai pending.

A menulis row outbox dalam transaksi ingest. C membaca, mengirim, dan menandainya. Tidak ada transaksi database yang ditahan selama menunggu jaringan Kafka. Polling satu detik adalah jalur inti; NOTIFY hanya opsi tambahan.

```go
type OutboxStore interface {
    Pending(ctx context.Context, limit int) ([]OutboxMessage, error)
    MarkPublished(ctx context.Context, id int64, acknowledgedAt time.Time) error
    Reject(ctx context.Context, id int64, reason string) error
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
