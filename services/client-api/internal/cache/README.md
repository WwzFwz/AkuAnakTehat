# cache

Fitur opsional yang belum diimplementasikan. Folder ini belum memuat package Go dan tidak dirangkai dalam runtime.

Singleflight dan micro-cache merupakan alternatif pengembangan untuk mengurangi fetch identik. Keduanya bukan syarat wajib M1. Jika dikerjakan, ukur dampaknya, batasi TTL dan kapasitas, serta pertahankan otorisasi dan proyeksi per request. Tidak ada sakelar cache aktif yang perlu diatur pengguna saat ini.
