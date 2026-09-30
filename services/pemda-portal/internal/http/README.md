# http

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Endpoint pembacaan view lokal dan health consumer.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `router.go` | GET /view, /health, dan /ready. |
| `handler.go` | Pembacaan view lokal dengan limit. |
| `health.go` | Status proses, broker, dan view store. |

## Kontrak dan alur

- /view mengembalikan data dari SQLite milik consumer, bukan Canonical Store.

## Dependensi

- Port pembacaan application/store lokal.

## Aturan penting

- /health bukan bukti backlog sudah nol; freshness/lag dilaporkan terpisah.
- Data view dapat memuat field raw; dashboard hanya bind127.0.0.1 pada host demo dan tetap dianggap internal.
- Health dan log wajib tersedia meski tidak ada UI.

## Langkah implementasi dan verifikasi

- Sepakati semantik readiness saat Kafka tidak tersedia tetapi view lama masih bisa dibaca.
- Buat respons health ringkas tanpa secret.
