# pvmbg-mock

Mock sumber independen untuk fondasi M1. **Status: implementasi fondasi tersedia; belum menjadi bukti lulus skenario tugas.**

Default port `8082` dapat diubah melalui `HTTP_ADDR`. Kontrak endpoint, autentikasi, environment, seed, dan simulasi dijelaskan dalam [kontrak mock HTTP](../../docs/api/mock-http.md).

## Menjalankan

Siapkan environment wajib melalui bootstrap secret repository atau konfigurasi lokal, lalu dari direktori service:

```sh
go run ./cmd/pvmbg-mock
```

Build container dengan context direktori service:

```sh
docker build -t pvmbg-mock:local .
```

Go 1.24.2, standard library saja; tidak memerlukan database atau broker. Dockerfile menjalankan binary sebagai user non-root. Credential kosong/salah format menggagalkan startup. Liveness `/health` dan readiness `/ready` tidak memerlukan credential.

## Komponen

- [cmd/pvmbg-mock](cmd/pvmbg-mock/README.md)
- [internal/auth](internal/auth/README.md)
- [internal/config](internal/config/README.md)
- [internal/domain](internal/domain/README.md)
- [internal/generator](internal/generator/README.md)
- [internal/http](internal/http/README.md)
- [internal/simulation](internal/simulation/README.md)
- [internal/store](internal/store/README.md)
- [seed](seed/README.md)

## Verifikasi dan batas

Pemeriksaan kompilasi dijalankan dengan `go test ./...`; belum ada test case perilaku. Docker/integrasi, race, credential lintas domain, since, schema drift, dan outage recovery masih perlu diuji sebelum fondasi dinyatakan tervalidasi. Tidak ada implementasi ingest BNPB di service ini.

Seed sintetis tetap dimuat setiap startup. Data runtime disimpan di memori, hilang saat restart, dan belum memiliki retensi; sesuai penggunaan demo terbatas.
