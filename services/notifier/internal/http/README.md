# http

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Endpoint health internal notifier.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `router.go` | GET /health dan /ready. |
| `health.go` | Status proses, broker, dan dedup store. |

## Kontrak dan alur

- Endpoint tersedia di jaringan internal; tidak memerlukan port publik.

## Dependensi

- Status worker dan store lokal.

## Aturan penting

- /health bukan bukti backlog sudah nol; freshness/lag dilaporkan terpisah.
- Tidak menyediakan endpoint untuk menerbitkan notifikasi bebas.
- Health dan log wajib tersedia meski tidak ada UI.

## Langkah implementasi dan verifikasi

- Sepakati semantik readiness saat Kafka tidak tersedia tetapi view lama masih bisa dibaca.
- Buat respons health ringkas tanpa secret.
