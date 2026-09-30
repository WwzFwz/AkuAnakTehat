# HTTP internal Aggregator — kontrak baseline

**Status:** kontrak query tahap B. Proses Aggregator ingest A sudah berjalan, tetapi route bisnis masih 503 `query_not_implemented`. `/ready/ingest` memeriksa DB; `/ready` tetap 503 sampai query dibuat. Client-api tetap mengembalikan 503 untuk pembacaan data.

## Transport dan autentikasi

- Base URL Compose yang direncanakan: `http://aggregator:9000`.
- Request bisnis wajib membawa `X-Internal-Key`. Nilai kredensial hanya berada pada konfigurasi lokal client-api dan verifikator Aggregator.
- `X-Correlation-ID` diteruskan dalam satu alur request dan dikembalikan pada respons.
- Timeout lokal client-api menuju Aggregator: default 1,5 detik. Propagasi deadline lewat header merupakan tambahan, bukan prasyarat.
- Error JSON: `error`, `message`, dan `correlation_id`. Tidak mengandung detail SQL atau secret.

## Daftar hazard

`GET /internal/hazards`

| Parameter | Aturan |
| --- | --- |
| `type` | Opsional: `SEISMIC` atau `VOLCANIC`; kosong berarti gabungan. |
| `severity` | Opsional: `NORMAL`, `WASPADA`, `SIAGA`, `AWAS`. |
| `since` | Opsional: timestamp RFC 3339 UTC, batas bawah inklusif `occurred_at`. |
| `limit` | Default 100, maksimum 500; harus positif. |
| `cursor` | Opaque cursor pasangan `occurred_at` dan `hazard_id`, dengan urutan turun. |

```json
{
  "data": [],
  "next_cursor": "",
  "sources": [
    {"source": "BMKG", "status": "HEALTHY"},
    {"source": "PVMBG", "status": "DOWN", "stale_since": "2026-09-30T01:00:00Z"}
  ]
}
```

`next_cursor` boleh kosong atau tidak disertakan bila halaman berikutnya tidak ada. `data` selalu array. `sources` memuat status sumber yang relevan dengan permintaan; status yang disepakati: `HEALTHY`, `DEGRADED`, `DOWN`.

Objek data berisi seluruh field HazardEvent dalam [kontrak event](hazard-event.md). Proyeksi hak akses client dilakukan oleh client-api. Metadata internal seperti content hash dan version storage tidak ditambahkan ke HazardEvent publik.

Hasil filter kosong tetap 200 dengan array kosong. Apabila sumber belum mempunyai data yang dapat digunakan dan sedang tidak tersedia, endpoint khusus sumber boleh mengembalikan 503 `source_unavailable`. Endpoint gabungan mempertahankan data sumber lain beserta metadata per sumber.

## Detail

`GET /internal/hazards/{hazard_id}` mengembalikan satu objek HazardEvent langsung, atau 404 `not_found`. Endpoint ini tidak melakukan panggilan ke mock sumber.

## Error

| Status | Makna |
| --- | --- |
| 400 | Parameter/filter/cursor tidak valid. |
| 401/403 | Kredensial internal tidak valid. |
| 404 | Hazard tidak ditemukan. |
| 429 | Penolakan beban terkendali; sertakan `Retry-After`. |
| 503 | Sumber tanpa data yang bisa digunakan atau dependensi internal belum tersedia. |

Health/readiness berada di luar endpoint bisnis. Detail status readiness ditentukan saat server Aggregator dibuat; client-api tidak boleh menganggap respons health proses sebagai bukti semua query siap.
