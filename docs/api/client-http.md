# Kontrak HTTP client-api M1

Pemilik: B. Kerangka executable sudah tersedia; data nyata menunggu implementasi API internal Aggregator.
Kontrak upstream: [Aggregator HTTP](aggregator-http.md). Tidak ada akses langsung database atau mock.

## Autentikasi dan rute

Authorization: Bearer JWT EdDSA dari [auth-service](token-http.md).
GET /v1/hazards, /v1/hazards/seismic, /v1/hazards/volcanic.
GET /v1/hazards/{id} dan /v1/hazards/{id}/raw.
GET /health dan /ready tidak memerlukan JWT.
Verifier memeriksa algoritma, signature, issuer, audience, exp, iat, sub, jti, scope.

List mendukung type=SEISMIC|VOLCANIC, severity, since (filter waktu occurred_at),
cursor opaque, limit (default100, maksimum500).
Rute seismic/volcanic menetapkan type sesuai rute.
List: `{"data":[<hazard>],"next_cursor":"optional","sources":[{"source":"BMKG","status":"HEALTHY","stale_since":"optional"}]}`.
Status sumber: HEALTHY, DEGRADED, DOWN.
Detail: objek hazard langsung. sources hanya memuat field metadata yang diizinkan.
Tidak ada data pengganti/fiktif saat Aggregator gagal.

## Proyeksi

Ringkasan allowlist: hazard_id, source, hazard_type, severity, area_name, occurred_at, ingested_at.
Scope hazard:read:raw menambah source_ref_id, latitude, longitude, attributes.
Field baru upstream tidak otomatis diteruskan.

fields adalah daftar field dipisahkan koma; respons hanya memuat field tersebut.
include=raw dan endpoint /raw merupakan permintaan raw eksplisit.
Media yang meminta field raw melalui fields, include, atau endpoint raw mendapatkan403 sebelum request upstream.
Scope raw mencakup izin data mentah; konfigurasi client juga wajib menyertakan scope summary.
Field tidak dikenal400. Tidak ada parameter role dari client.

## Error dan proteksi

Error JSON `{"error":"<code>"}`.
401 invalid_token;403 insufficient_scope;400 invalid_query/invalid_fields/invalid_limit/invalid_cursor/invalid_type/invalid_include;
404 not_found;429 rate_limited/upstream_overloaded dengan Retry-After:1;503 dependency_unavailable.
Timeout upstream lokal maksimum1500ms; pool HTTP dibatasi; tidak ada cache.
Readiness mencoba /ready Aggregator; gagal/tidak tersedia503. /health tetap200.
Response detail dan list selalu melewati allowlist; error upstream tidak diteruskan mentah.

## Konfigurasi

HTTP_ADDR=:8080; JWT_PUBLIC_KEY_FILE dan INTERNAL_KEY wajib;
JWT_ISSUER=bnpb-auth; JWT_AUDIENCE=bnpb-api; AGGREGATOR_URL=http://aggregator:9000;
AGGREGATOR_TIMEOUT=1500ms; MAX_CONCURRENT=100; RATE_LIMIT_RPS=100 per client terverifikasi;
RATE_LIMIT_BURST=200; PAGE_DEFAULT=100; PAGE_MAX=500.
Kredensial upstream X-Internal-Key dan X-Correlation-ID diteruskan; credential tidak dicatat.
Verifikasi integrasi Aggregator serta load test P2/P3 masih menunggu jalur inti.
