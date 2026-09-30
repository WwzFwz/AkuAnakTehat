# domain

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Identitas client, scope server, dan lifecycle keluarga refresh token.

**Pemilik rencana:** B. **Tahap:** Baseline / pendukung baseline.

**Status:** rancangan saja, belum diimplementasikan. Nama file dan operasi di bawah adalah usulan; file tersebut belum dibuat. Sesuaikan signature saat kontrak tim disepakati.

## Rencana file

| File yang akan dibuat | Tanggung jawab |
| --- | --- |
| `client.go` | Client ID, hash secret, dan scope yang diizinkan. |
| `refresh_token.go` | Hash token, family_id, client_id, status used, scope, dan expiry. |
| `token_pair.go` | Respons access token dan refresh token. |

## Kontrak dan alur

- Tiga identitas berbeda: media, field-team, bnpb-ops.
- TokenPair adalah kontrak penerbitan; RefreshRecord hanya state internal.

## Dependensi

- Standard library; digunakan application dan adapter auth-service.

## Aturan penting

- Tidak menyimpan refresh token mentah sebagai record DB.
- Scope pengganti tidak boleh melebihi scope grant awal.
- Jangan mengimpor Redis, JWT library, atau HTTP ke model domain.

## Langkah implementasi dan verifikasi

- Sepakati masa hidup family dan token bekas agar deteksi reuse tetap dapat bekerja.
