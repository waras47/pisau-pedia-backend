# Seeds

## `production_products_seed.sql`

Data katalog **asli** (19 produk pisau, dari `template-produk-pisau.xlsx` yang diisi user),
digenerate otomatis dari spreadsheet — bukan ditulis manual, supaya tidak ada salah ketik dari
19 deskripsi produk yang panjang.

**Isi:** 10 kategori baru (Slicer, K-Tip Gyuto, Gyuto*, Deba, Petty*, Honesuki, Yanagiba, Taiwan,
Kiritsuke, Bunka* — yang ditandai `*` sudah ada di skema, di-skip otomatis kalau sudah ada) +
19 produk (nama, harga, harga coret, stok, berat, badge, deskripsi) + 149 baris `product_specs`
(Blade Shape/Steel Type/Blade Length/dst per produk).

**Cara pakai di production:**

```bash
mysql -h <host> -P <port> -u <user> -p <database> < seeds/production_products_seed.sql
```

Aman dijalankan berkali-kali — tiap `INSERT` dibungkus `WHERE NOT EXISTS (...)` berdasarkan
`slug`, jadi kalau sebagian data sudah pernah masuk, tidak akan dobel. Sudah divalidasi lewat
dry-run (`START TRANSACTION` + `ROLLBACK`) di database lokal, tanpa error.

**Belum termasuk — foto produk.** Kolom "Gambar" di spreadsheet cuma berisi link halaman listing
Tokopedia (`tokopedia.com/pisaupedia/...`), bukan URL file gambar langsung — dan sesuai kebijakan
proyek ini, gambar tidak pernah di-hotlink/scrape otomatis dari situs lain, termasuk marketplace
sendiri. Daftar link referensinya ada di `tokopedia_reference_links.md` — dipakai manual untuk
unduh foto asli lalu upload lewat form admin (Products → Edit → 4 slot foto sudut), sama seperti
alur upload foto produk yang sudah ada.
