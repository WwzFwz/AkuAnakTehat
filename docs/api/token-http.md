# Kontrak token M1

Pemilik: B. Kontrak baseline; implementasi awal tersedia di auth-service, belum bukti demo P3.
Ini protokol tugas berbentuk grant OAuth, bukan klaim kepatuhan OAuth penuh.

## Endpoint

`POST /oauth/token`, Content-Type `application/x-www-form-urlencoded`, body maksimum 8192 byte.

- Login: `grant_type=client_credentials&client_id=<id>&client_secret=<secret>`.
- Refresh: `grant_type=refresh_token&refresh_token=<opaque>`; tidak membutuhkan secret ulang.
- Scope dari CLIENTS_FILE; parameter scope request tidak dapat menaikkan izin.
- Respons 200: `{"access_token":"...","token_type":"Bearer","expires_in":60,"refresh_token":"...","scope":"hazard:read:summary"}`.
- Semua respons token: Cache-Control no-store dan Pragma no-cache.
- Error JSON `{"error":"<code>"}`: 401 invalid_client, 400 invalid_grant/invalid_request/unsupported_grant_type, 413 body terlalu besar, 415 media type salah, 429 rate_limited/overloaded, 503 temporarily_unavailable.

## Identitas dan kriptografi

| Client | Scope |
| --- | --- |
| media | hazard:read:summary |
| field-team | hazard:read:summary hazard:read:raw |
| bnpb-ops | hazard:read:summary hazard:read:raw |

JWT memakai EdDSA/Ed25519, iss=bnpb-auth, aud=bnpb-api, sub=client_id, scope string dipisahkan spasi, exp, iat, jti acak.
Default access TTL60s. Client-api memverifikasi lokal dan menolak token kedaluwarsa tanpa leeway.
Private key PKCS8 PEM hanya di auth-service; public key PKIX PEM di client-api.

CLIENTS_FILE adalah array JSON objek `{client_id,secret_hash,scopes}`.
secret_hash adalah SHA256 secret acak berentropi tinggi, hex lowercase. Bukan format penyimpanan password manusia.

## Lifecycle refresh

Token opaque 32 byte acak base64url; Redis hanya menyimpan hash SHA256.
Family memiliki lifetime tetap 8h sejak login (tidak diperpanjang setiap refresh).
Rotasi menyimpan pengganti dan menandai token lama used secara atomik melalui Lua.
Token used mempertahankan family_id sampai expiry; reuse mencabut keluarga.
Dua refresh bersamaan atas token sama menyebabkan satu rotasi berhasil dan request berikutnya mencabut keluarga.
Client harus menserialisasi refresh.

Signing dan pembangkitan pengganti selesai sebelum rotasi. Script memvalidasi key/state sebelum mutasi.
Atomisitas Redis tidak meliputi respons HTTP: respons hilang setelah rotasi bisa memaksa autentikasi ulang.
Pencabutan family tidak membatalkan access JWT yang sudah terbit; JWT tetap berlaku sampai exp.

## Konfigurasi dan operasi

HTTP_ADDR=:8090; JWT_PRIVATE_KEY_FILE, CLIENTS_FILE, REDIS_PASSWORD wajib.
JWT_ISSUER=bnpb-auth; JWT_AUDIENCE=bnpb-api; REDIS_ADDR=auth-store:6379;
REDIS_TIMEOUT=200ms; ACCESS_TOKEN_TTL=60s; REFRESH_TOKEN_TTL=8h;
TOKEN_RATE_LIMIT=20 request/detik global; TOKEN_RATE_BURST=40; maksimum 100 request token bersamaan.
Tanpa Redis, proses tetap hidup; request/readiness mencoba koneksi lagi dengan timeout lokal.
GET /health =200 liveness; GET /ready =200 saat Redis bisa PING, selain itu503.
Log JSON tidak memuat credential/token; X-Correlation-ID diterima jika valid atau dibangkitkan.
