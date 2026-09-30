# reference

[Panduan service](../README.md) · [Peta repository](../../../README.md)

Referensi statis nama serta koordinat gunung api milik BNPB.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** jalur A sudah diimplementasikan. Berkas tersedia: `embed.go`, `volcanoes.json`. Cakupan pengujian ada di [bukti ingest](../../../docs/evidence/ingest/README.md). Tabel rencana di bawah adalah panduan pemecahan tanggung jawab; sebagian operasi digabung dalam file yang tersedia.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `volcanoes.json` | Daftar volcano_id, nama, latitude, dan longitude. |
| `embed.go` | Embed referensi ke binary. |
| `load.go` | Validasi dan indeks pencarian berdasarkan ID. |

## Kontrak dan alur

- Lookup(volcanoID) menghasilkan referensi gunung atau hasil not-found.
- Daftar ID untuk fixture disepakati dengan pembuat mock melalui dokumentasi.

## Dependensi

- Dipakai canonicalize/ingest melalui data yang dirangkai main; bukan import kode mock.

## Aturan penting

- Unknown volcano dikarantina, bukan dipetakan ke koordinat null.
- Referensi mock harus diberi label sintetis jika bukan data aktual tervalidasi.
- Tidak membaca filesystem atau database mock lintas service.

## Langkah implementasi dan verifikasi

- Siapkan referensi minimum untuk seed PVMBG.
- Verifikasi ID unik dan koordinat valid.
