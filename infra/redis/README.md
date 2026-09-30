# redis

[Peta repository](../../README.md)

Konfigurasi Redis auth-store milik auth-service.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** Redis 7.4.2 tersedia dalam Compose dengan AOF, password env, named volume, dan healthcheck. Rotasi/reuse, refresh bersamaan, serta sesi bertahan setelah stop/start telah diuji. Expiry refresh 8 jam belum diuji dengan waktu nyata. Lihat [hasil fondasi](../../docs/evidence/foundation/README.md).

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `redis.conf` | AOF, direktori persistensi, dan konfigurasi server non-secret. |

## Kontrak dan alur

- Redis hanya ada di auth_net bersama auth-service; data pada named volume.

## Dependensi

- Compose memasok secret terpisah; key schema dan Lua milik auth-service.

## Aturan penting

- Password tidak ditulis ke redis.conf yang ter-commit.
- Jangan memakai Redis ini sebagai cache client-api.
- Dokumentasikan pilihan fsync/persistensi dan keterbatasannya.
- AOF memakai default appendfsync everysec: crash host dapat kehilangan sekitar satu detik penulisan terakhir. Ini bukan jaminan durability penuh.

## Langkah implementasi dan verifikasi

- Konfigurasikan AOF dan password dari mekanisme secret runtime.
- Verifikasi restart dan TTL tanpa membocorkan key/token.
