# aggregator

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Composition root aggregator; tempat merangkai seluruh dependensi runtime.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A, wiring relay C, dan wiring query HTTP B sudah diimplementasikan. Main membuka pool ingest, relay, dan baca terpisah, lalu menjalankan seluruh worker serta readiness query. Berkas tersedia: `main.go`. Cakupan pengujian ada di [bukti ingest](../../../../docs/evidence/ingest/README.md).

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `main.go` | Load config, bangun dependensi, jalankan HTTP/worker, dan graceful shutdown. |

## Kontrak dan alur

- main memulai service dan menghentikannya saat SIGTERM; bukan tempat logika bisnis.

## Dependensi

- Package milik service ini; tidak meng-import module service lain.

## Aturan penting

- Secret wajib tidak ada atau migrasi gagal membuat startup gagal. Compose menunggu database healthy dan memakai restart policy. Sesudah startup, polling yang gagal dicoba lagi sesuai interval/breaker; kegagalan DB membuat readiness ingest 503.
- Jangan menjadikan semua dependency wajib hidup untuk liveness.
- Selesaikan pekerjaan sesuai deadline shutdown; commit consumer hanya pekerjaan yang sudah selesai.

## Langkah implementasi dan verifikasi

- Aggregator: pemilik Canonical Store, ingest, query internal, dan outbox.
- Rangkai jalur inti sebelum fitur tambahan. Pastikan health tersedia tanpa port publik bila service internal.
