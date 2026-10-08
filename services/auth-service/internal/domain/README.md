# domain

[Panduan service](../../README.md) · [Peta repository](../../../../README.md)

Identitas client, scope server, dan lifecycle keluarga refresh token.

**Status:** diimplementasikan. Cakupan verifikasi mengikuti pengujian yang dirujuk di bawah.

## Kontrak dan alur

- Tiga identitas berbeda: media, field-team, bnpb-ops.
- TokenPair adalah kontrak penerbitan; RefreshRecord hanya state internal.

## Dependensi

- Standard library; digunakan application dan adapter auth-service.

## Aturan penting

- Tidak menyimpan refresh token mentah sebagai record DB.
- Scope pengganti tidak boleh melebihi scope grant awal.
- Jangan mengimpor Redis, JWT library, atau HTTP ke model domain.

## Berkas implementasi

| Berkas | Tanggung jawab |
| --- | --- |
| [types.go](types.go) | Tipe identitas client, pasangan token, dan record internal refresh token. |

## Verifikasi

Jalankan `go test ./...` dan `go vet ./...` dari root module service. Pengujian lintas service memerlukan stack aktif dan dijalankan terpisah dari unit test. Lihat [audit persyaratan](../../../../docs/requirements-audit.md) untuk pemetaan ke spesifikasi, lokasi bukti, dan batas yang belum terpenuhi.
