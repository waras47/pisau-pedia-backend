# 13 — Notifikasi Aktivitas Admin (Product, Service, Order, Customer)

Admin panel sebelumnya punya ikon lonceng di header yang murni
dekoratif — tidak ada dropdown, tidak ada data, titik merahnya statis.
Dokumen ini merangkum pembangunan sistem notifikasi sungguhan dari nol
untuk 4 modul: **Product**, **Service** (sharpening/engraving — satu
entity `ServiceRequest` di backend), **Order**, dan **Customer**. Kode
lengkap tidak disalin ulang di sini — cukup rujuk path file dan status
tiap bagian.

## Cakupan & keputusan desain

- **Storage**: tabel `notifications` baru, bukan derive dari kolom
  `updated_at` yang sudah ada di tabel lain — supaya bisa punya pesan
  yang jelas per event, status baca/belum-baca, dan tidak berubah kalau
  row aslinya diedit lagi setelahnya.
- **Trigger**: dipanggil eksplisit dari usecase terkait setelah mutasi
  sukses (pola yang sama seperti `ProductRepository.UpdateRatingStats`
  yang dipanggil eksplisit dari `ReviewUsecase`) — bukan trigger
  database, tetap "no magic" sesuai gaya codebase ini.
- **Delivery**: polling tiap 30 detik dari frontend, bukan WebSocket/SSE
  — tidak ada infrastruktur real-time sama sekali di project ini, dan
  skala single-instance saat ini tidak butuh kompleksitas tambahan itu.
- **Kegagalan insert notifikasi tidak boleh menggagalkan aksi utama** —
  order/produk/dst tetap berhasil dibuat walau penyimpanan notifikasi
  gagal; error-nya diabaikan (`_ = err`) di titik pemanggilan, bukan
  di-propagate.
- **Event yang memicu notifikasi**: order baru + perubahan status,
  service request baru + perubahan status, customer baru mendaftar,
  produk dibuat/dihapus admin, dan stok produk menipis (pakai
  `lowStockThreshold = 5` yang sudah ada di
  [product_usecase.go](../internal/usecase/product_usecase.go) — sudah
  diantisipasi lewat komentar di kode sejak modul Reports dibangun,
  lihat [docs/11](11-orders-reports-excel-pdf.md)).

## 1. Backend — tabel & layer baru

| File | Isi |
|---|---|
| [migrations/000008_create_notifications.up.sql](../migrations/000008_create_notifications.up.sql) | Tabel `notifications`: `module` (enum `product`/`order`/`customer`/`service`), `type` (string bebas per event, mis. `order_created`), `title`/`message`, `reference_id` & `link` nullable, `is_read`, `created_at` |
| [internal/entity/notification.go](../internal/entity/notification.go) | Struct `Notification` + tipe `NotificationModule` |
| [internal/repository/notification_repository.go](../internal/repository/notification_repository.go) + [internal/infrastructure/mysql/notification_repo.go](../internal/infrastructure/mysql/notification_repo.go) | `Create`, `FindAll` (filter unread-only + pagination), `CountUnread`, `MarkAsRead`, `MarkAllAsRead` |
| [internal/usecase/notification_usecase.go](../internal/usecase/notification_usecase.go) | CRUD dasar (`ListNotifications`, `CountUnread`, `MarkAsRead`, `MarkAllAsRead`) + 7 helper per-event (`NotifyOrderCreated`, `NotifyOrderStatusChanged`, `NotifyServiceRequestCreated`, `NotifyServiceRequestStatusChanged`, `NotifyCustomerRegistered`, `NotifyProductCreated`, `NotifyProductDeleted`, `NotifyProductLowStock`) — masing-masing menyusun title/message berbahasa Indonesia + link ke halaman admin terkait |
| [internal/delivery/http/dto/notification_dto.go](../internal/delivery/http/dto/notification_dto.go) + [internal/delivery/http/handler/notification_handler.go](../internal/delivery/http/handler/notification_handler.go) | Request/response mapping + handler |

**Endpoint** (semua admin-only, di [router.go](../internal/delivery/http/router/router.go)):

| Endpoint | Fungsi |
|---|---|
| `GET /admin/notifications?page=&per_page=&unread_only=` | List notifikasi, terpaginasi |
| `GET /admin/notifications/unread-count` | Endpoint ringan khusus buat polling badge — tidak perlu fetch list penuh tiap 30 detik |
| `PATCH /admin/notifications/:id/read` | Tandai satu notifikasi sudah dibaca |
| `PATCH /admin/notifications/read-all` | Tandai semua sudah dibaca |

**Status:** ✅ Selesai, `go build ./...` dan `go vet ./...` sukses.

## 2. Backend — titik integrasi

Constructor 4 usecase menerima parameter baru `notificationUsecase
*NotificationUsecase`, di-wire di
[cmd/api/main.go](../cmd/api/main.go):

| Usecase | Event | Titik integrasi |
|---|---|---|
| `AuthUsecase.Register` | Customer baru | Setelah `userRepo.Create` sukses |
| `ProductUsecase.CreateProduct` | Produk dibuat | Setelah create sukses |
| `ProductUsecase.DeleteProduct` | Produk dihapus | Produk di-`FindByID` sebelum delete, dipakai untuk notifikasi setelah delete sukses |
| `ProductUsecase.UpdateProduct` | Stok menipis | Simpan `oldStock` sebelum di-overwrite; fire **hanya** saat stok baru ≤5 **dan** stok lama masih >5 (baru saja nembus batas, bukan tiap kali diedit — mencegah notifikasi spam) |
| `OrderUsecase.CreateOrder` | Order baru | Setelah order + invoice sukses dibuat |
| `OrderUsecase.UpdateOrderStatus` | Status order berubah | Status lama diambil dari `FindByID` sebelum update; notifikasi cuma dikirim kalau status benar-benar berubah |
| `ServiceRequestUsecase.CreateRequest` | Service request baru | Setelah create sukses |
| `ServiceRequestUsecase.UpdateRequest` | Status service berubah | Status lama disimpan sebelum di-overwrite; notifikasi cuma dikirim kalau status benar-benar berubah |

**Status:** ✅ Selesai & terverifikasi lewat `curl` — tiap event
ditrigger manual (register, buat order, ubah status order, buat
service request, ubah status service request, buat produk, hapus
produk, ubah stok produk lewat & keluar dari ambang batas menipis) dan
dicek entrinya muncul benar di `GET /admin/notifications`. Kasus
khusus diverifikasi: mengedit produk dua kali berturut-turut sambil
tetap di bawah `lowStockThreshold` **tidak** menghasilkan notifikasi
duplikat — cuma satu entri saat pertama kali menembus ambang batas.

## 3. Frontend — API & polling

| File | Isi |
|---|---|
| `src/entities/notification/api/notification.api.ts` | `listNotifications`, `getUnreadCount`, `markAsRead`, `markAllAsRead` — pola sama seperti `newsletter.api.ts`, list pakai `apiFetchPaginated` |
| `src/widgets/admin/admin-header/AdminHeader.tsx` | Ikon lonceng dekoratif diganti fungsional |

Detail `AdminHeader.tsx`:
- Polling `getUnreadCount()` tiap 30 detik (`setInterval`), badge merah
  jadi kondisional dengan angka asli (bukan titik statis lagi),
  tampilkan `9+` kalau lebih dari 9.
- Klik lonceng membuka dropdown (pola sama seperti dropdown profil yang
  sudah ada di file yang sama) — list penuh (`listNotifications({
  perPage: 10 })`) cuma di-fetch saat dropdown dibuka, bukan tiap
  polling tick, biar polling tetap ringan.
- Klik satu notifikasi → `markAsRead` + navigasi ke `notification.link`
  (mis. order baru → `/admin/orders?status=pending`, sudah nyambung ke
  filter status yang sudah ada di halaman Orders).
- Tombol "Tandai semua dibaca" → `markAllAsRead`, update optimistik di
  UI (badge langsung ke 0, list langsung ditandai terbaca) sebelum
  request selesai.

**Status:** ✅ Selesai, `npx tsc --noEmit` dan `eslint` bersih,
`npm run build` sukses.

## 4. Verifikasi end-to-end (browser)

Ditrigger 2 event lewat `curl` (order baru + service request baru)
sambil admin panel terbuka di browser:
- Polling `GET /admin/notifications/unread-count` terkonfirmasi jalan
  lewat network tab.
- Badge lonceng menampilkan angka `2` yang benar.
- Dropdown menampilkan kedua notifikasi dengan title/message/waktu
  relatif ("1 menit lalu") yang benar.
- Klik notifikasi order → berhasil pindah ke `/admin/orders?status=pending`
  dan tabel benar-benar terfilter ke order yang baru dibuat.
- Unread count di backend turun dari 2 ke 1 setelah diklik (dicek
  lewat `curl` langsung ke API, bukan cuma asumsi dari UI).
- Setelah re-login (sesi sempat expired di tengah testing — tidak
  terkait fitur ini), badge menampilkan `1` yang cocok persis dengan
  `GET /admin/notifications/unread-count`.

## Ringkasan status

| Bagian | Status |
|---|---|
| Migration `notifications` | ✅ Selesai |
| Backend — repository, usecase, handler, route | ✅ Selesai & terverifikasi |
| Wiring notifikasi ke 8 titik di 4 usecase | ✅ Selesai & terverifikasi (termasuk kasus no-spam low-stock) |
| Frontend — API client | ✅ Selesai |
| Frontend — lonceng fungsional (polling, dropdown, mark read) | ✅ Selesai & terverifikasi di browser |

**Catatan follow-up (belum dikerjakan, di luar scope sesi ini):**
- Belum ada UI untuk melihat riwayat notifikasi lama di luar 10 yang
  ditampilkan di dropdown (tidak ada halaman `/admin/notifications`
  tersendiri) — backend sudah mendukung pagination penuh lewat `GET
  /admin/notifications?page=`, tinggal dibuatkan halamannya kalau
  dibutuhkan nanti.
- Notifikasi bersifat global (satu status baca/belum-baca untuk semua
  admin), belum per-akun admin — cukup untuk saat ini karena baru ada
  satu akun admin, tapi perlu didesain ulang kalau nanti ada banyak
  admin yang masing-masing ingin status baca terpisah.
