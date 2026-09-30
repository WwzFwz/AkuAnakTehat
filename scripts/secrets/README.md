# secrets

[Peta repository](../../README.md)

Bootstrap kredensial pengembangan per service, tanpa hardcoded secret.

**Pemilik rencana:** C. **Tahap:** Baseline / pendukung baseline.

**Status:** generator diimplementasikan. Tests idempotensi, penolakan konfigurasi parsial, dan penolakan pasangan key tidak cocok lulus. Generator sudah dijalankan lokal; hasilnya diabaikan Git.

## File

| File | Tanggung jawab |
| --- | --- |
| `generate.go` | Generator stdlib Go memakai crypto/rand; tidak merotasi secret yang sudah ada. |
| `generate.sh` | Entrypoint Linux/macOS, dengan cache Go lokal. |
| `generate.ps1` | Entrypoint Windows PowerShell, dengan cache Go lokal. |
| `generate_test.go` | Memverifikasi idempotensi dan penolakan konfigurasi yang tidak konsisten. |

## Kontrak dan alur

- Menghasilkan env/<service>.env, data client, dan pasangan Ed25519 di env/keys.
- Private key memakai Ed25519 PKCS8 PEM dan public key memakai PKIX PEM.
- env/clients.json adalah array {client_id, secret_hash, scopes}; secret_hash SHA-256 lowercase hex. Media mendapat hazard:read:summary; field-team dan bnpb-ops juga mendapat hazard:read:raw.
- env/demo-clients.json menyimpan client_id, client_secret, scopes untuk demo lokal. env/demo.env memuat BMKG_API_KEY, PVMBG_TOKEN, PVMBG_ADMIN_KEY, dan INTERNAL_KEY.

## Dependensi

- Go 1.24.2. Jalankan `powershell -NoProfile -File scripts/secrets/generate.ps1` atau `sh scripts/secrets/generate.sh` dari root repository.

## Aturan penting

- .gitignore mengecualikan env/kunci, termasuk kredensial demo lokal.
- Tidak merotasi/menimpa secret yang sudah ada tanpa tindakan eksplisit.
- Private key hanya auth-service; public key client-api; mock verifikator menerima hash kredensial.
- Jangan mencetak secret ke stdout, log, atau docs/evidence.
- Jika konfigurasi parsial, generator berhenti tanpa mengubah berkas yang tersisa. Pulihkan backup lokal atau koordinasikan reset seluruh secret secara eksplisit; env PostgreSQL baru tidak mengubah password volume lama.
- Lock env/.bootstrap.lock mencegah generator paralel. Hapus lock tertinggal hanya setelah memastikan tidak ada generator aktif.
- Izin berkas 0600 dan direktori 0700 berlaku pada OS pendukung; Windows memakai ACL workspace pengguna.

## Langkah implementasi dan verifikasi

- Nama env dan format key sudah disepakati dengan pemilik mock/auth/client-api.
- Runtime integrasi Compose masih perlu diuji; bootstrap saja belum membuktikan layanan berhasil start.
