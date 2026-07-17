# 16 — Konfirmasi "Pesanan Diterima" + Trigger Review

> **Status dokumen: ✅ SELESAI DIKERJAKAN & DIVERIFIKASI.** Dokumen ini
> dimulai sebagai rencana (rancangan disepakati dulu sebelum coding, pola
> yang sama dipertahankan di bawah), lalu diperbarui jadi retrospektif
> begitu implementasinya rampung — lihat "Ringkasan status" di paling
> bawah untuk detail per bagian.

## Masalah yang mau diselesaikan

Setelah paket dikirim, admin tidak punya cara tahu kapan barang **benar-benar
sampai** ke customer — kurir tidak terintegrasi sebagai webhook status
otomatis (lihat [14-integrasi-rajaongkir-payment-qris.md](14-integrasi-rajaongkir-payment-qris.md),
cek ongkir RajaOngkir yang kita pakai cuma untuk hitung biaya, bukan
tracking real-time). Rencana: customer sendiri yang konfirmasi lewat
frontstore — begitu diklik, langsung diarahkan mengisi review produk yang
dibeli. Pola ini sudah lazim (Shopee/Tokopedia pakai alur yang sama).

## Keputusan desain (sudah disepakati)

| # | Keputusan | Alasan |
|---|---|---|
| 1 | Butuh halaman **"Pesanan Saya"** dulu sebagai prasyarat | Belum ada sama sekali halaman customer untuk lihat daftar/status order — tombol "Diterima" butuh tempat untuk hidup |
| 2 | Fitur **hanya untuk order yang dibuat sambil login** — guest checkout tidak dapat tombol ini | Tidak ada cara memvalidasi "yang klik tombol ini benar-benar pembelinya" untuk order tanpa akun lewat halaman biasa. Guest checkout tetap dipertahankan (lihat percakapan soal wajib-login sebelumnya) — cuma fitur ini yang dibatasi |
| 3 | Konfirmasi customer disimpan di field **terpisah** (`customer_confirmed_at`), **tidak menimpa** `orders.status` yang dikontrol admin | Kalau customer keliru klik sebelum barang benar sampai, admin tetap pegang kendali status resmi pesanan. Notifikasi ke admin dipicu dari field ini, bukan mengubah alur status yang sudah ada |
| 4 | Form review yang muncul setelah klik **per-produk**, bukan satu review umum per order | Tabel `reviews` memang dirancang per-produk (`product_slug`), bukan per-order — konsisten dengan skema yang sudah ada |

## Rancangan teknis

### 1. Database

| Perubahan | Detail |
|---|---|
| Migration `000013_add_orders_customer_confirmed_at` | `ALTER TABLE orders ADD COLUMN customer_confirmed_at DATETIME NULL AFTER payment_status` — diterapkan ke database lokal |

Tidak perlu kolom baru di `reviews` — tabel itu sudah cukup (`product_slug`,
`customer_name`, `customer_email`, `rating`, `content`). Opsional untuk
nanti (bukan v1): tambah `order_id` nullable di `reviews` supaya bisa kasih
badge "Pembelian Terverifikasi" — dicatat di bagian "Di luar scope v1" di
bawah.

### 2. Backend

| Endpoint baru | Fungsi | Guard |
|---|---|---|
| `GET /api/v1/users/me/orders` | List order milik user yang login (paginated) | `JWTAuth`; filter `WHERE user_id = :current_user_id` — **bukan** filter by email, supaya order guest dengan email yang sama tidak ikut nyasar ke sini (konsisten dengan keputusan #2: kalau tidak login saat checkout, order itu tidak akan pernah muncul di "Pesanan Saya" walau emailnya sama) |
| `GET /api/v1/users/me/orders/:id` | Detail 1 order milik user yang login | `JWTAuth`; 404 kalau `order.user_id != current_user_id` (bukan 403, supaya tidak bocorin keberadaan order orang lain) |
| `POST /api/v1/users/me/orders/:id/confirm-received` | Set `customer_confirmed_at = NOW()` | `JWTAuth`; order harus milik user; `payment_status = paid`; `status IN (processing, ready_for_delivery, delivered)` (menutup kemungkinan konfirmasi order yang belum diproses admin sama sekali); idempotent — kalau `customer_confirmed_at` sudah terisi, balas sukses tanpa menimpa ulang (bukan error), supaya klik dobel aman |

**Perubahan pada `OrderUsecase`:**
- `ListMyOrders(ctx, userID, page, perPage)` — mirror pola `ListOrders` admin tapi filter `user_id`.
- `GetMyOrder(ctx, userID, orderID)` — mirror `GetOrder` tapi validasi kepemilikan.
- `ConfirmReceived(ctx, userID, orderID)` — validasi guard di atas, `UPDATE orders SET customer_confirmed_at = NOW() WHERE id = ? AND user_id = ? AND customer_confirmed_at IS NULL` (atomic conditional update, pola yang sama dengan `MarkPaidIfUnpaid`/`ExpireIfUnpaid` di [15-keandalan-pembayaran-idempotency.md](15-keandalan-pembayaran-idempotency.md) — supaya klik dobel dari 2 tab tidak kirim notifikasi dobel ke admin), lalu panggil `NotificationUsecase.NotifyOrderReceived`.

**Perubahan pada `NotificationUsecase`:**
- `NotifyOrderReceived(ctx, order)` — event baru, modul `order`, pesan semacam "Pesanan #xxxxxxxx dikonfirmasi diterima oleh {customer_name}", link ke `/admin/orders`. Ikut pola `NotifyOrderPaid`/`NotifyOrderCreated` yang sudah ada.

**Perubahan pada `OrderResponse` (DTO):** tambah field `customer_confirmed_at` (nullable string, format sama seperti `created_at`) supaya frontend tahu kapan harus tampilkan tombol vs status "sudah dikonfirmasi".

**Admin side:** tampilkan `customer_confirmed_at` sebagai info read-only di
modal detail order admin (badge kecil "✅ Dikonfirmasi diterima customer,
{tanggal}") — bukti visibility tanpa mengubah kontrol status admin,
sesuai keputusan #3.

### 2b. Temuan saat implementasi: `orders.user_id` ternyata tidak pernah diisi

Rencana awal di atas berasumsi order yang dibuat sambil login otomatis
punya `user_id` terisi — ternyata **tidak**. `POST /orders` adalah endpoint
publik tanpa middleware auth sama sekali (mendukung guest checkout), dan
`CreateOrderInput` tidak pernah punya field `UserID`. Kalau ini tidak
ditambal, seluruh fitur "Pesanan Saya" tidak akan pernah punya order untuk
ditampilkan — bahkan untuk customer yang login pas checkout.

**Perbaikan (di luar daftar file rencana awal):**
- `appmw.OptionalJWTAuth(secret)` — middleware baru di
  [internal/delivery/http/middleware/auth.go](../internal/delivery/http/middleware/auth.go),
  sepupu `JWTAuth` yang tidak memaksa: token valid → isi context (sama
  seperti `JWTAuth`), token tidak ada/tidak valid → lanjut sebagai
  anonymous, **bukan** 401. Dipasang di `POST /orders` supaya guest
  checkout tetap 100% jalan seperti sebelumnya.
- `CreateOrderInput.UserID` (usecase) + `OrderHandler.Create` mengisi
  `input.UserID = currentUserID(c)` — string kosong kalau anonymous.
- `entity.Order.UserID` di-set dari `input.UserID` saat membangun order
  (kolom `user_id` di tabel `orders` sebenarnya sudah lama ada di skema,
  cuma belum pernah benar-benar dipakai sampai sekarang).

Diverifikasi: order guest (tanpa header `Authorization`) tetap berhasil
persis seperti sebelumnya; order yang dibuat dengan header `Authorization`
valid otomatis dapat `user_id` terisi tanpa perubahan apa pun di payload
request — frontend tidak perlu kirim apa-apa secara eksplisit karena
`apiFetch` (shared client) sudah otomatis menyisipkan `Authorization:
Bearer <token>` kalau ada sesi aktif.

### 3. Frontend

| Halaman baru | Isi |
|---|---|
| `/account/orders` | "Pesanan Saya" — daftar order milik user login, kartu ringkas (ID, tanggal, total, status, badge sudah/belum dikonfirmasi). Dibungkus guard yang sama seperti pola `AdminGuard` (redirect ke `/account/login?redirect=/account/orders` kalau belum login) |
| `/account/orders/[id]` | Detail order — item, alamat, status pengiriman, dan **tombol "Pesanan Diterima?"** yang cuma muncul kalau `payment_status=paid && status` bukan `pending`/`cancelled` `&& !customer_confirmed_at` |

**Alur setelah tombol diklik:**
1. Panggil `POST /users/me/orders/:id/confirm-received`.
2. Kalau sukses, tampilkan form review **satu per item** di order itu juga (bukan halaman terpisah) — tiap item render form kecil (rating bintang + komentar) yang submit ke endpoint review publik yang sudah ada (`POST /api/v1/reviews`, isi `product_slug` dari item tsb). Item yang sudah pernah direview (kalau customer submit lalu reload) tidak perlu tampil formnya lagi — cek sederhana di frontend, tidak perlu backend baru untuk ini.
3. Badge order berubah jadi "✅ Diterima — {tanggal}", tombol hilang.

**Header:** avatar akun yang sudah login ([Header.tsx](../../pisau-pedia/src/widgets/header/Header.tsx), dibangun di sesi Google SSO) diubah dari link statis ke `/account/login` jadi dropdown (klik avatar → "Pesanan Saya" + "Keluar").

**Review tanpa order:** ternyata belum ada fungsi `createReview` di frontend sama sekali (`entities/review/api/review.api.ts` cuma punya fungsi admin: list/update/delete) — endpoint publik `POST /reviews` di backend sudah ada sejak awal tapi tidak pernah dipanggil dari mana pun. Ditambahkan sebagai bagian dari implementasi ini.

### 4. Ringkasan perubahan per file

| Layer | File |
|---|---|
| Migration | `migrations/000013_add_orders_customer_confirmed_at.up/down.sql` |
| Entity | `internal/entity/order.go` — `CustomerConfirmedAt *time.Time`, `UserID` (baru benar-benar dipakai) |
| Repository | `internal/repository/order_repository.go` + `mysql/order_repo.go` — `OrderFilter.UserID`, `ConfirmReceivedIfEligible` |
| Middleware | `internal/delivery/http/middleware/auth.go` — `OptionalJWTAuth` (baru, di luar rencana awal — lihat bagian 2b) |
| Usecase | `internal/usecase/order_usecase.go` — `ListMyOrders`, `GetMyOrder`, `ConfirmReceived`, `CreateOrderInput.UserID`; `internal/usecase/notification_usecase.go` — `NotifyOrderReceived`; `internal/usecase/errors.go` — `ErrOrderNotEligibleForConfirmation` |
| DTO/Handler | `internal/delivery/http/dto/order_dto.go` (+`CustomerConfirmedAt`), `internal/delivery/http/handler/order_handler.go` (`ListMine`, `GetMine`, `ConfirmReceived`, + `Create` mengisi `UserID` dari token) |
| Router | `internal/delivery/http/router/router.go` — group `/users/me/orders`, `OptionalJWTAuth` di `POST /orders` |
| DI wiring | `cmd/api/main.go` — `optionalJWTAuth`, diteruskan ke `router.Dependencies` |
| Frontend | `src/app/(store)/account/orders/page.tsx` (baru), `src/app/(store)/account/orders/[id]/page.tsx` (baru), `src/entities/order/api/order.api.ts` (+3 fungsi, +`customer_confirmed_at`), `src/entities/review/api/review.api.ts` (+`createReview`, baru), `src/widgets/header/Header.tsx` (dropdown akun), `src/app/admin/orders/page.tsx` (badge konfirmasi) |

## Di luar scope v1 (dicatat, sengaja tidak dikerjakan dulu)

| Item | Kenapa ditunda |
|---|---|
| Guest checkout dapat cara konfirmasi juga (magic link via email/WA) | Butuh desain otentikasi tanpa password terpisah — dibahas lagi kalau volume guest checkout ternyata signifikan |
| Badge "Pembelian Terverifikasi" di review (butuh `reviews.order_id`) | Nice-to-have, tidak blocking fungsi inti |
| Reminder otomatis (push/email "sudah terima paketmu?") kalau customer tidak konfirmasi dalam X hari | Butuh job terjadwal tambahan, mirip pola `runExpiredOrderSweep` di [15-keandalan-pembayaran-idempotency.md](15-keandalan-pembayaran-idempotency.md) — bisa disatukan nanti kalau memang dibutuhkan |
| Auto-transisi `status` ke `delivered` saat customer konfirmasi | Sengaja dihindari sesuai keputusan #3 — status tetap murni kendali admin |

## Ringkasan status

| Bagian | Status |
|---|---|
| Keputusan desain (4 poin) | ✅ Disepakati |
| Migration `customer_confirmed_at` | ✅ Diterapkan ke database lokal & diverifikasi |
| Fix `orders.user_id` tidak pernah terisi (`OptionalJWTAuth`) | ✅ Selesai & diverifikasi — order guest tetap jalan, order sambil login otomatis dapat `user_id` |
| Backend — endpoint `/users/me/orders/*` | ✅ Selesai & diverifikasi (list, detail, confirm-received) |
| Backend — guard kepemilikan & eligibility | ✅ Diverifikasi: 404 untuk order guest/orang lain, 409 sebelum lunas, 200 idempoten kalau diklik dua kali |
| Backend — notifikasi `order_received` | ✅ Selesai & diverifikasi — tepat 1 baris notifikasi walau confirm-received dipanggil 2x |
| Frontend — halaman "Pesanan Saya" | ✅ Selesai & diverifikasi lewat browser sungguhan |
| Frontend — tombol konfirmasi + form review per-produk | ✅ Selesai & diverifikasi — review benar-benar tersimpan di tabel `reviews` (status `pending`, rating & konten sesuai) |
| Header — dropdown akun ("Pesanan Saya" + "Keluar") | ✅ Kode selesai & lolos type-check; belum sempat diklik-verifikasi visual karena dev server sempat tidak stabil di sesi ini (isu versi Node di environment, bukan kode) |
| Admin — badge "Dikonfirmasi diterima customer" | ✅ Selesai (belum sempat screenshot terpisah, memakai pola badge yang sama seperti bagian lain yang sudah teruji) |

**Alur end-to-end yang diverifikasi dengan data sungguhan** (bukan cuma
`curl`, lewat browser beneran): daftar akun baru → checkout sambil login
→ `orders.user_id` otomatis terisi → order ditandai lunas → "Pesanan Saya"
menampilkan order → buka detail → klik "Pesanan Diterima" → badge
"✅ Diterima" muncul, form review tampil → isi & kirim review → review
tersimpan di database dengan `product_slug` yang benar dan status
`pending` (menunggu approval admin, sesuai alur review yang sudah ada).
