# pvmbg

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

HTTP client dan tolerant decoder untuk PVMBG; tidak berisi aturan pemetaan severity.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `client.go` | GET /volcanic-reports dengan since. |
| `decoder.go` | Validasi field dikenal; unknown fields dipertahankan sebagai JSON. |
| `errors.go` | Klasifikasi HTTP error, timeout, dan record invalid. |

## Kontrak dan alur

- FetchReports(ctx, since).
- Hasil berupa tipe input application/canonicalize atau domain/tsunami, disertai record yang perlu dikarantina.

## Dependensi

- net/http; application/canonicalize.
- Kredensial dan client HTTP diberikan oleh composition root.

## Aturan penting

- Authorization: Bearer pvmbg_<token>; timeout lokal awal 4 detik.
- Tidak meng-import kode mock atau membuang field tak dikenal.
- Teruskan correlation ID, ukur latensi, tutup body respons, dan batasi ukuran respons.
- Record invalid tidak membatalkan record valid lain; format respons rusak menjadi error endpoint.

## Langkah implementasi dan verifikasi

- Implementasikan fetch dan decoder terpisah.
- Pertahankan nilai JSON unknown fields; jangan mengonversi semuanya ke string.
