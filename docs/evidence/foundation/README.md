# Verifikasi fondasi lokal — 1 Oktober 2026

Fondasi A/B/C berhasil dijalankan pada Windows + Docker Desktop (Linux containers), Go 1.24.2. Hasil ini mencakup fondasi yang sudah dibuat; Aggregator, relay outbox, dan consumer bisnis belum diimplementasikan.

## Hasil

| Pemeriksaan | Hasil dan batas |
| --- | --- |
| Build/start Compose | Empat image aplikasi berhasil dibangun; PostgreSQL, Redis, Kafka, mock, auth, dan client-api berjalan. Init dua topic selesai exit 0. |
| Health | Semua liveness 200; auth readiness 200. Client-api readiness 503 karena Aggregator belum ada. |
| Kontrak mock | Minimal 20 seed, ID unik, since inklusif, since invalid 400, hasil kosong `[]`, kredensial lintas domain ditolak. |
| Schema/outage PVMBG | Generator menghasilkan confidence_level untuk record v2; record v1 tetap tersedia. Outage error 503, hang dibatalkan deadline client, admin memulihkan layanan. |
| Warning BMKG | Test generator deterministik membuktikan warning sebelum/sesudah gempa dan eskalasi dengan ID tetap terambil melalui watermark waktu perubahan. |
| JWT/otorisasi | Tiga client memperoleh token; token termodifikasi, expired, issuer/audience salah, dan iat masa depan ditolak. Media meminta raw lewat fields/include/route mendapat 403. |
| Proyeksi | Test router dengan upstream stub membuktikan ringkasan Media 7 field, seleksi fields, raw untuk scope berizin, dan field upstream baru tidak bocor. Belum integrasi data Aggregator nyata. |
| Refresh Redis | Rotasi berhasil; reuse mencabut keluarga termasuk token pengganti; dua refresh bersamaan memberi tepat satu 200 dan satu 400, lalu descendant ditolak. |
| Dependency recovery | Redis dihentikan: liveness auth tetap 200, readiness dan refresh 503. Sesudah Redis hidup, sesi yang belum dipakai dapat di-refresh. |
| Persistence | Marker PostgreSQL dan pesan Kafka bertahan setelah restart; refresh token Redis bertahan stop/start dengan AOF. Volume tidak dihapus. Tabel/topic fixture dibersihkan. |
| Log | Correlation ID, method, status, latency dalam JSON; kredensial sumber serta access/refresh token yang diuji tidak muncul dalam sampel log aplikasi. |
| Check lokal | `scripts/check/check.ps1` lulus: test bootstrap, test dan vet keempat module. |
| Secret scan lokal | Gitleaks 8.24.2 pada histori sampai `f73c035` memeriksa 5 commit dan tidak menemukan kebocoran. Hasil scan memiliki cakupan histori tersebut. |

Suite runtime terakhir lulus **7 test dalam 34,941 detik**. [Output yang tidak memuat secret](runtime-test-output.txt). Test module dan bootstrap diperiksa terpisah.

## Mengulang

Ikuti [panduan infra](../../../infra/README.md), lalu [petunjuk suite pengujian](../../../scripts/check/README.md). Perintah dari root:

```text
go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m
```

Suite memodifikasi simulasi mock sementara dan restart infrastruktur lokal. Gunakan saat tidak ada demo lain. Ini uji restart normal, bukan bukti ketahanan terhadap power loss atau broker HA.

## Perbaikan dan batas

- Tambahkan healthcheck HTTP aplikasi supaya Compose dapat menunggu proses siap.
- UID/GID auth/client-api mengikuti pemilik bind mount secret pada Linux melalui Make/CI. Windows memakai default non-root; private key tidak perlu dibuat world-readable pada Linux.
- Workflow CI kini juga menjalankan suite integrasi. Status run GitHub Actions belum dapat dikonfirmasi: endpoint repository pada GitHub API tanpa autentikasi mengembalikan 404 dari sesi ini. Itu tidak membuktikan CI gagal atau lulus.
- Tidak ada klaim P1–P5 lengkap, load/SLO, race detector, expiry refresh 8 jam dengan waktu nyata, restart consumer, atau pengiriman notifikasi. Bagian tersebut memerlukan jalur inti dan pengujian berikutnya.
