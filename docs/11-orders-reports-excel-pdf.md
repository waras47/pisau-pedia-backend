# 11 — Modul Order & Reports (Excel/PDF)

Dokumen ini merangkum penambahan modul Order (backend + checkout
storefront + admin panel) dan modul Reports (Sales Report & Inventory
Report, exportable ke Excel/PDF). Kode lengkap tidak disalin ulang di
sini — cukup rujuk path file dan status tiap bagian.

## Cakupan & keputusan desain

- **Guest checkout** — order tidak mewajibkan login. `user_id` di
  tabel `orders` nullable, data customer (nama/email/telepon/alamat)
  di-snapshot langsung ke baris order, bukan cuma FK ke `addresses`.
- **Belum ada payment gateway sungguhan** — sengaja dirancang lewat
  interface `PaymentGateway` di usecase layer, dengan implementasi
  `DummyGateway` sekarang. Order dibuat dengan status `unpaid`, admin
  tandai `paid` manual lewat panel selama belum ada gateway asli.
  Begitu ada gateway (Xendit atau lainnya), tinggal buat implementasi
  baru dari interface yang sama — usecase & handler tidak perlu diubah.
- **Harga dihitung ulang dari katalog produk asli di backend**, tidak
  pernah dipercaya dari input client — pola yang sama seperti API
  checkout Next.js yang lama (sebelum dihapus).
- **Reports digenerate di backend** (Go), bukan di browser — agregasi
  data (SUM revenue, GROUP BY status/kategori, top produk) lebih rapi
  dikerjakan di SQL/Go, dan konsisten dengan pola project ini (semua
  logic data lewat backend).

## 1. Migration & Entity Order

| File | Isi |
|---|---|
| [migrations/000003_create_orders.up.sql](../migrations/000003_create_orders.up.sql) | Tabel `orders` (status pesanan + status pembayaran terpisah, data customer & alamat di-snapshot, kolom `xendit_external_id`/`xendit_invoice_url` buat integrasi pembayaran nanti) dan `order_items` (snapshot nama/harga/slug produk per baris) |
| [internal/entity/order.go](../internal/entity/order.go) | `Order`, `OrderItem`, enum `OrderStatus` (`pending`/`processing`/`ready_for_delivery`/`delivered`/`cancelled`) dan `PaymentStatus` (`unpaid`/`paid`/`expired`/`failed`) |

**Status:** ✅ Migration dijalankan ke database lokal, tabel terverifikasi.

## 2. Backend — modul Order

Pola sama seperti modul Product (entity → repository interface →
mysql repo → usecase → DTO → handler → route):

| File | Peran |
|---|---|
| [internal/repository/order_repository.go](../internal/repository/order_repository.go) + [internal/infrastructure/mysql/order_repo.go](../internal/infrastructure/mysql/order_repo.go) | CRUD + filter status/rentang tanggal, create transactional (order + items sekaligus) |
| [internal/usecase/payment_gateway.go](../internal/usecase/payment_gateway.go) | Interface `PaymentGateway` — `CreateInvoice(ctx, input) (*InvoiceResult, error)` |
| [internal/infrastructure/payment/dummy_gateway.go](../internal/infrastructure/payment/dummy_gateway.go) | `DummyGateway` — langsung kasih URL ke `/checkout/success`, tidak ada charge sungguhan |
| [internal/usecase/order_usecase.go](../internal/usecase/order_usecase.go) | `CreateOrder` (hitung ulang harga dari katalog), `ListOrders`, `GetOrder`, `UpdateOrderStatus`, `UpdatePaymentStatus` |
| [internal/delivery/http/dto/order_dto.go](../internal/delivery/http/dto/order_dto.go) + [internal/delivery/http/handler/order_handler.go](../internal/delivery/http/handler/order_handler.go) | Request/response mapping + handler |

**Endpoint:**

| Endpoint | Akses |
|---|---|
| `POST /api/v1/orders` | Publik (guest checkout) |
| `GET /api/v1/admin/orders` | Admin, filter `?status=` |
| `GET /api/v1/admin/orders/:id` | Admin |
| `PATCH /api/v1/admin/orders/:id/status` | Admin |
| `PATCH /api/v1/admin/orders/:id/payment-status` | Admin (tandai `paid` manual) |

**Status:** ✅ Selesai, `go build ./...` sukses. Diverifikasi lewat
`curl`: buat order → data tersimpan di `orders`/`order_items` → ubah
status via endpoint admin. 8 order dummy diseed buat testing (variasi
status pending/processing/ready_for_delivery/delivered/cancelled).

## 3. Storefront — checkout disambungkan ke backend asli

| File | Perubahan |
|---|---|
| `src/entities/order/api/order.api.ts` (baru) | `createOrder()` manggil `POST /orders` |
| `src/features/checkout/ui/CheckoutForm.tsx` | Diganti total: dari fetch ke `/api/checkout` (Next.js, hitung harga dari data mock) jadi manggil `createOrder()` ke Go backend (hitung harga dari database asli) |
| `src/app/api/checkout/route.ts` | **Dihapus** — logic-nya pindah total ke Go backend |

Catatan penting:
- Item hasil configurator (blade/handle/accessory custom) **belum
  didukung** backend order — di-skip dari payload saat checkout, kalau
  cart isinya cuma item custom, muncul pesan error yang jelas.
- Cart baru di-clear di halaman `/checkout/success`, **bukan** langsung
  setelah `createOrder()` — supaya begitu payment gateway asli
  dipasang (redirect ke luar dulu), order yang belum kebayar tidak
  langsung menghilangkan isi cart pelanggan.

**Status:** ✅ Selesai, `npx tsc --noEmit` bersih.

## 4. Admin — halaman Orders

| File | Perubahan |
|---|---|
| `src/entities/order/api/order.api.ts` | Ditambah `listOrders`, `getOrder`, `updateOrderStatus`, `updatePaymentStatus` |
| `src/app/admin/orders/page.tsx` | Dari placeholder "coming soon" jadi halaman penuh: baca filter status dari query param (nyambung ke link sidebar "Pending"/"Processing"/"Delivered" yang sudah ada dari awal), tabel order + badge status, modal detail (info customer, alamat, dropdown ubah status pesanan & status pembayaran, daftar item) |

**Status:** ✅ Selesai, `npx tsc --noEmit` bersih, render terverifikasi.

## 5. Reports — helper generate Excel/PDF (generik)

| File | Isi |
|---|---|
| [pkg/export/table.go](../pkg/export/table.go) | Struct generik `Table{Title, Headers, Rows}` — dipakai bareng oleh semua laporan |
| [pkg/export/excel.go](../pkg/export/excel.go) | `ToExcel(Table) ([]byte, error)` pakai `github.com/xuri/excelize/v2` — judul di-merge di baris atas, header bold + fill abu-abu, kolom auto-width |
| [pkg/export/pdf.go](../pkg/export/pdf.go) | `ToPDF(Table) ([]byte, error)` pakai `github.com/go-pdf/fpdf` — landscape A4, tabel bordered dengan header shaded, kolom dibagi rata |

**Status:** ✅ Selesai, `go build ./...` sukses.

## 6. Sales Report

| File | Perubahan |
|---|---|
| `internal/repository/order_repository.go` + `internal/infrastructure/mysql/order_repo.go` | `GetSalesSummary(ctx, from, to)` — total revenue (`payment_status='paid'`), total order, breakdown jumlah order per status, revenue per hari, top 5 produk terlaris |
| `internal/usecase/order_usecase.go` | `GetSalesReport(ctx, from, to)` — gabung `SalesSummary` (buat preview JSON) + daftar order mentah dalam rentang tanggal (buat isi export) |
| `internal/delivery/http/dto/order_dto.go` | `SalesReportResponse` |
| `internal/delivery/http/handler/order_handler.go` | `GetSalesReport` (JSON), `ExportSalesReport` (Excel/PDF, tabel: Order ID, Tanggal, Customer, Status, Pembayaran, Total) |

**Endpoint:**

| Endpoint | Fungsi |
|---|---|
| `GET /admin/reports/sales?from=YYYY-MM-DD&to=YYYY-MM-DD` | JSON preview (default rentang: 30 hari terakhir kalau `from`/`to` tidak diisi) |
| `GET /admin/reports/sales/export?format=xlsx&from=&to=` | Download Excel |
| `GET /admin/reports/sales/export?format=pdf&from=&to=` | Download PDF |

**Status:** ✅ Selesai & terverifikasi lewat `curl` — JSON preview
mengembalikan total revenue, breakdown status, revenue per hari, top 5
produk; file `.xlsx` dan `.pdf` yang dihasilkan dicek valid (`file`
command: `Microsoft Excel 2007+` dan `PDF document, version 1.3`).

## 7. Inventory Report

| File | Perubahan |
|---|---|
| `internal/repository/product_repository.go` + `internal/infrastructure/mysql/product_repo.go` | `GetInventorySummary(ctx, lowStockThreshold)` — total produk, total nilai stok (`SUM(price*stock)`), jumlah stok menipis & habis, breakdown jumlah produk/stok per kategori |
| `internal/usecase/product_usecase.go` | `GetInventoryReport(ctx)` — gabung summary + daftar produk lengkap (buat isi export). Konstanta `lowStockThreshold = 5`, **sengaja disamakan** dengan threshold yang direncanakan buat fitur alert stok menipis (lihat task terpisah yang sudah di-spawn sebelumnya) |
| `internal/delivery/http/dto/product_dto.go` | `InventoryReportResponse` |
| `internal/delivery/http/handler/product_handler.go` | `GetInventoryReport` (JSON), `ExportInventoryReport` (Excel/PDF, tabel: Nama Produk, Kategori, Stok, Harga, Nilai Stok) |

**Endpoint:**

| Endpoint | Fungsi |
|---|---|
| `GET /admin/reports/inventory` | JSON preview |
| `GET /admin/reports/inventory/export?format=xlsx` | Download Excel |
| `GET /admin/reports/inventory/export?format=pdf` | Download PDF |

**Status:** ✅ Selesai & terverifikasi lewat `curl` — dengan 31 produk
di database lokal: `total_products: 30` (1 produk `is_active=0` tidak
terhitung), `total_stock_value: 970875000`, `out_of_stock_count: 1`,
breakdown per kategori (Aksesoris 22 produk, Gyuto 3, Nakiri 2,
Bunka/Petty/Santoku masing-masing 1). File `.xlsx` dan `.pdf` valid.

## 8. Yang belum dikerjakan (follow-up)

**Frontend admin — halaman Reports UI belum dibuat.** Backend
(endpoint JSON preview + export) sudah lengkap dan siap dipakai, tapi
`admin/reports/sales/page.tsx` dan `admin/reports/inventory/page.tsx`
masih placeholder "coming soon". Rencana ketika dikerjakan:

- Stat card (total revenue/order untuk Sales; total produk/nilai
  stok/stok menipis/habis untuk Inventory)
- Tabel ringkas (top produk / breakdown kategori)
- Date range picker (khusus Sales Report)
- Tombol "Export Excel" / "Export PDF" — perlu fetch dengan header
  `Authorization` (endpoint admin butuh token), terima response
  sebagai blob, baru trigger download — **tidak bisa** cuma pakai
  `<a href="...">` biasa karena butuh auth header.

## Ringkasan status

| Bagian | Status |
|---|---|
| Migration `orders`/`order_items` | ✅ Selesai |
| Backend modul Order (CRUD + dummy payment gateway) | ✅ Selesai & terverifikasi |
| Checkout storefront disambungkan ke backend asli | ✅ Selesai |
| Admin panel — halaman Orders | ✅ Selesai & terverifikasi |
| Helper export Excel/PDF generik (`pkg/export`) | ✅ Selesai |
| Sales Report (backend + export) | ✅ Selesai & terverifikasi |
| Inventory Report (backend + export) | ✅ Selesai & terverifikasi |
| Halaman admin Reports (frontend UI) | ⏳ Belum dikerjakan — backend sudah siap dipakai |

**Catatan operasional:** sepanjang sesi ini berkali-kali ketemu isu
"kode sudah diubah tapi belum kepakai" karena `go run` (lewat `make
run`) tidak hot-reload — setiap perubahan backend butuh restart manual.
Juga sempat ketemu isu terpisah: cache route Next.js (`.next`) yang
korup setelah banyak file dihapus/ditambah dalam satu sesi, sampai
seluruh `/admin/*` sempat 404 — solusinya hapus folder `.next` lalu
restart `npm run dev`.
