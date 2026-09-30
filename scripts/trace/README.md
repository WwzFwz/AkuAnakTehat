# trace

[Peta repository](../../README.md)

Mencari log satu correlation ID di seluruh container.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `trace.sh` | Filter log Compose berdasarkan correlation ID persis. |
| `trace.ps1` | Alternatif Windows bila diperlukan. |

## Kontrak dan alur

- Rencana pemanggilan: scripts/trace/trace.sh <id>.

## Dependensi

- Log JSON semua service dan Docker Compose.

## Aturan penting

- Validasi argumen; hindari evaluasi shell dari ID pengguna.
- Jangan menganggap correlation ID unik untuk seluruh waktu hidup aplikasi; satu alur punya ID sendiri.
- Output tidak boleh memuat secret/header sensitif.

## Langkah implementasi dan verifikasi

- Tentukan parser log yang diperlukan dan dokumentasikan prasyarat.
- Verifikasi alur ingest→outbox→consumer dapat dilacak.
