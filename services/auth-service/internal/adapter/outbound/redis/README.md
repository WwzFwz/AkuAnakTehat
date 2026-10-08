# redis

[Panduan service](../../../../README.md) · [Peta repository](../../../../../../README.md)

Persistensi refresh token dan status keluarga dengan TTL.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Memenuhi port Store application untuk create, rotate, dan family revocation.
- Operasi rotasi mengembalikan hasil terklasifikasi, bukan error string mentah ke client.

## Dependensi

- domain/application, go-redis; hanya auth-service menjangkau auth-store.

## Aturan penting

- Timeout lokal awal200 ms.
- Token used mempertahankan family_id; periksa status revoked sebelum menerbitkan pengganti.
- Validasi argumen/key type sebelum mutasi; atomisitas Lua tidak berarti rollback otomatis atas semua runtime error.
- AOF/volume dan timeout perlu dikonfigurasi; tidak mengklaim zero data loss untuk semua crash.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [rotate.lua](rotate.lua) | Rotasi atomik state refresh token dan pencabutan keluarga ketika reuse terdeteksi. |
| [store.go](store.go) | Namespace key, hash token, TTL, create/get, pemanggilan Lua, dan probe Redis. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
