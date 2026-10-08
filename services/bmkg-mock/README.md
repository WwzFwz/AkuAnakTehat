# bmkg-mock

Mock sumber independen untuk fondasi M1. **Status: diimplementasikan dan terhubung ke pengujian integrasi lokal.**

Default port `8081` dapat diubah melalui `HTTP_ADDR`. Kontrak endpoint, autentikasi, environment, seed, dan simulasi dijelaskan dalam [kontrak mock HTTP](../../docs/api/mock-http.md).

## Menjalankan

Siapkan environment wajib melalui bootstrap secret repository atau konfigurasi lokal, lalu dari direktori service:

```sh
go run ./cmd/bmkg-mock
```

Build container dengan context direktori service:

```sh
docker build -t bmkg-mock:local .
```

Go 1.24.2, standard library saja; tidak memerlukan database atau broker. Dockerfile menjalankan binary sebagai user non-root. Credential kosong/salah format menggagalkan startup. Liveness `/health` dan readiness `/ready` tidak memerlukan credential.

## Komponen

- [cmd/bmkg-mock](cmd/bmkg-mock/README.md)
- [internal/auth](internal/auth/README.md)
- [internal/config](internal/config/README.md)
- [internal/domain](internal/domain/README.md)
- [internal/generator](internal/generator/README.md)
- [internal/http](internal/http/README.md)
- [internal/store](internal/store/README.md)
- [seed](seed/README.md)

## Verifikasi dan batas

Kontrak mock diperiksa pada [pengujian fondasi](../../docs/evidence/foundation/README.md). Alur ingest, perubahan skema, dan outage memiliki bukti terpisah pada [audit persyaratan](../../docs/requirements-audit.md). Tidak ada klaim race test atau pengukuran kapasitas maksimum mock. Tidak ada implementasi ingest BNPB di service ini.

Seed sintetis tetap dimuat setiap startup. Data runtime disimpan di memori, hilang saat restart, dan belum memiliki retensi; sesuai penggunaan demo terbatas.
