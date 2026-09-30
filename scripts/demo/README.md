# demo

[Peta repository](../../README.md)

Skenario demonstrasi P1–P5 yang dapat dijalankan ulang dan menghasilkan bukti nyata.

**Pemilik rencana:** A/B/C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `p1-schema-change.sh` | Aktifkan confidence_level tanpa restart; periksa pemetaan dan atribut. |
| `p2-outage.sh` | Picu/pulihkan outage; buktikan BMKG tetap dilayani dan vulkanik bertanda stale. |
| `p3-auth.sh` | Kredensial silang, Media raw403, serta refresh Tim Lapangan otomatis. |
| `p4-independent-restart.sh` | Rebuild satu service tanpa restart lainnya serta bukti storage ownership. |
| `p5-consumer-catchup.sh` | Stop/start consumer dan tambah pemda tanpa perubahan producer. |

## Kontrak dan alur

- Setiap script menulis output tersanitasi ke docs/evidence/pN.
- Script gagal dengan exit code nonzero bila perilaku wajib tidak terpenuhi.

## Dependensi

- Sistem yang sudah diimplementasikan; endpoint dan konfigurasi yang didokumentasikan.

## Aturan penting

- Tidak mengganti bukti nyata dengan log contoh.
- Simpan/kembalikan state simulasi jika script mengubahnya; admin explicit enabled/version lebih mudah diulang.
- P3 menunggu expiry alami dan melakukan refresh serial; jangan mengubah jam sistem.
- Restart P4 gunakan target tepat dan --no-deps; jangan menghapus volume untuk demo restart.

## Langkah implementasi dan verifikasi

- A menangani P1 dan ingest/outage P2; B P3 dan load test; C P4/P5.
- Sediakan prasyarat dan cleanup setiap skenario.
