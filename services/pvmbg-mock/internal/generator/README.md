# generator

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Pembangkitan data baru PVMBG untuk membuktikan polling berkala.

**Pemilik rencana:** A. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `generator.go` | Loop periodik yang dapat dihentikan. |
| `factory.go` | Membangkitkan laporan dari ID gunung yang terdaftar. |

## Kontrak dan alur

- Run(ctx) menambahkan minimal satu record baru per10 detik per instansi melalui store lokal.

## Dependensi

- domain dan store lokal; snapshot state simulation menentukan schema version.

## Aturan penting

- Mock tidak push ke BNPB; hanya membuat data agar dapat dipoll.
- ID runtime unik lintas restart; ID seed tetap.
- Tambahkan confidence_level hanya pada laporan baru ketika schema version aktif.

## Langkah implementasi dan verifikasi

- Pastikan interval dapat dipercepat untuk demo.
- Pisahkan penggunaan waktu/acak agar skenario penting mudah direproduksi.
