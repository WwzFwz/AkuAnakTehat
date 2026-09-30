# redis

[Peta repository](../../README.md)

Konfigurasi Redis auth-store milik auth-service.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

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

## Langkah implementasi dan verifikasi

- Konfigurasikan AOF dan password dari mekanisme secret runtime.
- Verifikasi restart dan TTL tanpa membocorkan key/token.
