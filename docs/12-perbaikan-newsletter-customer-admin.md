# 12 — Perbaikan Admin Panel: Newsletter & Customer

Modul Newsletter dan Customer sempat dibangun di luar sesi dokumentasi
utama — [08-roadmap-checklist.md](08-roadmap-checklist.md) baris 56-78
secara eksplisit mencatatnya sebagai "sengaja tidak dibahas". Akibatnya
kualitasnya tidak direview seperti modul Product/Order: tab Subscribers
di Newsletter tidak pernah pakai pagination/filter server-side
walaupun backend sudah mendukungnya, dan halaman Customer tidak punya
tampilan detail sama sekali walaupun fungsi `getCustomer()` sudah
ditulis di API client tapi tidak pernah dipanggil. Dokumen ini
merangkum perbaikan yang dilakukan untuk menutup gap tersebut. Kode
lengkap tidak disalin ulang di sini — cukup rujuk path file dan status
tiap bagian.

## Cakupan & keputusan

**Dikerjakan:**
1. Newsletter tab Subscribers — pagination + search/filter status
   beneran ke server (sebelumnya fetch sekali `per_page=200` lalu
   filter di client, jadi kalau subscriber > 200 sisanya tidak pernah
   muncul).
2. Modal detail Customer — profil, riwayat order, daftar alamat.
3. Dua bug backend: kolom `role` hilang saat update user, dan N+1
   query di endpoint list customer.

**Sengaja tidak disentuh (lihat bagian 7):** tab Campaign (email
sending) di Newsletter, dan pagination/filter status untuk tabel list
Customer.

## 1. Bug backend — kolom `role` hilang saat update user

[internal/infrastructure/mysql/user_repo.go:105-118](../internal/infrastructure/mysql/user_repo.go)
— query `UPDATE users SET ...` tidak menyertakan kolom `role`, jadi
kalau ada fitur di masa depan yang mengubah role user (mis. promote ke
admin), perubahannya tidak akan pernah tersimpan. Semua caller
`Update()` (`UpdateProfile`, `ChangePassword`, `UpdateCustomerStatus`)
selalu load full `entity.User` dulu lewat `FindByID`, jadi menambah
`role = :role,` ke query aman — tidak mengubah perilaku yang sudah ada,
cuma menutup celah untuk fitur berikutnya.

**Status:** ✅ Selesai. Diverifikasi: toggle status customer lewat
`PATCH /admin/customers/:id/status`, lalu cek langsung ke tabel
`users` — kolom `role` tetap `customer`, tidak berubah jadi kosong
atau ter-reset.

## 2. Bug backend — N+1 query di `ListCustomers`

Sebelumnya `UserUsecase.ListCustomers` memanggil
`orderRepo.GetCustomerStats(ctx, user.Email)` di dalam loop per baris
customer — kalau ada 50 customer di satu halaman, itu 50 query
terpisah ke tabel `orders`.

| File | Perubahan |
|---|---|
| [internal/repository/order_repository.go:16](../internal/repository/order_repository.go) | Tambah struct `CustomerStats{OrderCount, TotalSpent}` dan method `GetCustomerStatsBulk(ctx, emails []string) (map[string]CustomerStats, error)` ke interface `OrderRepository` |
| [internal/infrastructure/mysql/order_repo.go:218-247](../internal/infrastructure/mysql/order_repo.go) | Implementasi pakai `sqlx.In` — satu query `SELECT customer_email, COUNT(*), SUM(total) ... WHERE customer_email IN (?) GROUP BY customer_email`, bukan N query |
| [internal/usecase/user_usecase.go:86-96](../internal/usecase/user_usecase.go) | `ListCustomers` kumpulkan semua email dari hasil `FindAll`, panggil `GetCustomerStatsBulk` sekali, lalu map hasilnya ke tiap customer. `GetCustomer` (single, ambil 1 row) tetap pakai `GetCustomerStats` yang lama — tidak perlu diubah |

**Status:** ✅ Selesai, `go build ./...` dan `go vet ./...` sukses.
Diverifikasi lewat `curl`: list customer dengan beberapa order paid,
`order_count`/`total_spent` di response tetap akurat setelah perubahan.

## 3. Endpoint baru — riwayat order & alamat per customer

Dibutuhkan untuk modal detail Customer di frontend (bagian 5).

| Endpoint | Cara kerja |
|---|---|
| `GET /admin/orders?customer_email=x@y.com` | Reuse endpoint list order yang sudah ada — cukup tambah filter, bukan endpoint baru |
| `GET /admin/customers/:id/addresses` | Endpoint baru, reuse `AddressRepository.FindByUserID` yang sudah dipakai untuk profil user sendiri |

| File | Perubahan |
|---|---|
| [internal/repository/order_repository.go:16](../internal/repository/order_repository.go) | Tambah field `CustomerEmail string` ke `OrderFilter` |
| [internal/infrastructure/mysql/order_repo.go:41-44](../internal/infrastructure/mysql/order_repo.go) | `FindAll` tambah kondisi `customer_email = ?` kalau filter diisi |
| [internal/usecase/order_usecase.go:34-38, 116](../internal/usecase/order_usecase.go) | `OrderListInput` tambah `CustomerEmail`, diteruskan ke `OrderFilter` |
| [internal/delivery/http/handler/order_handler.go:78](../internal/delivery/http/handler/order_handler.go) | `List` baca `c.QueryParam("customer_email")` |
| [internal/usecase/user_usecase.go:121-123](../internal/usecase/user_usecase.go) | `ListCustomerAddresses(ctx, userID)` — delegasi langsung ke `addressRepo.FindByUserID`, sama seperti `ListAddresses` yang dipakai endpoint `/users/me/addresses` |
| [internal/delivery/http/handler/user_handler.go:153-158](../internal/delivery/http/handler/user_handler.go) | Handler `GetCustomerAddresses`, pakai `dto.ToAddressResponses` yang sudah ada |
| [internal/delivery/http/router/router.go:82](../internal/delivery/http/router/router.go) | Route `admin.GET("/customers/:id/addresses", ...)` |

**Status:** ✅ Selesai & terverifikasi lewat `curl` — order dibuat untuk
satu email, lalu `GET /admin/orders?customer_email=...` cuma
mengembalikan order milik email itu; alamat ditambahkan lewat endpoint
customer sendiri lalu muncul benar di `GET
/admin/customers/:id/addresses`.

## 4. Frontend — helper `apiFetchPaginated`

Sebelumnya `apiFetch<T>()` di `src/shared/api/client.ts` cuma
mengembalikan `data`, membuang field `meta` (`page`/`total`/
`total_pages`) yang sebenarnya sudah dikirim backend di setiap endpoint
list. Ini alasan utama kenapa tidak ada halaman admin yang punya
pagination nyata sebelumnya.

Refactor: logic fetch/auth/refresh-token diekstrak ke helper internal
`apiFetchEnvelope`, dipakai bareng oleh `apiFetch` (ambil `.data` saja,
signature tidak berubah — semua caller lama tetap jalan) dan fungsi
baru `apiFetchPaginated<T>()` yang mengembalikan `{ data, meta }`.

**Status:** ✅ Selesai, `npx tsc --noEmit` dan `eslint` bersih.

## 5. Newsletter — tab Subscribers pakai pagination & filter server-side

Backend tidak perlu diubah — `NewsletterHandler.ListSubscribers` sudah
lama menerima `page`, `per_page`, `status`, `search`.

| File | Perubahan |
|---|---|
| `src/entities/newsletter/api/newsletter.api.ts` | `listSubscribers()` diganti terima `{page, perPage, status, search}`, pakai `apiFetchPaginated`, return `{items, meta}` |
| `src/app/admin/newsletter/page.tsx` | Tab Subscribers: search di-debounce 300ms, filter status & page dikirim ke API (bukan lagi filter di client atas 200 baris fetch), tambah kontrol pagination Prev/Next, hapus `filteredSubs` (tidak perlu lagi karena hasil sudah difilter server). Kartu stats "Total/Active/Unsubscribed" diambil dari 3 request `per_page=1` terpisah (independen dari filter aktif) supaya angkanya tetap benar walau tabel sedang difilter/dipaginasi — kartu "New This Month" **dihapus** karena backend tidak punya filter tanggal untuk subscriber, jadi tidak ada cara akurat menghitungnya tanpa fetch semua data. Tab Campaigns (mock) tidak disentuh |

**Status:** ✅ Selesai & terverifikasi manual — test dengan 3 subscriber
dummy dan `per_page=2`: halaman 1 dan 2 menampilkan baris berbeda,
pencarian mengirim request `search=` ke backend (dicek lewat network
tab), stats tetap akurat saat tabel difilter.

## 6. Customer — modal detail (profil + riwayat order + alamat)

Ikut pola modal detail yang sudah ada di `src/app/admin/orders/page.tsx`
(`openDetail` fetch on click) — bukan route `/admin/customers/[id]`
baru, supaya konsisten dengan codebase yang sudah ada (belum ada satu
pun halaman detail dengan route dinamis di admin panel ini).

| File | Perubahan |
|---|---|
| `src/entities/customer/api/customer.api.ts` | Tambah `getCustomerAddresses(id)` dan tipe `AddressApiItem` (belum ada tipe address di frontend sama sekali sebelumnya) |
| `src/entities/order/api/order.api.ts` | `listOrders(status?, customerEmail?)` — tambah parameter kedua opsional |
| `src/app/admin/customers/page.tsx` | Tombol "Detail" per baris → `openDetail(id)` manggil `getCustomer`, `listOrders(undefined, email)`, `getCustomerAddresses(id)` paralel lewat `Promise.all`. Modal tampilkan profil (nama/email/telepon/tanggal join/status), stats (order_count/total_spent), tabel riwayat order (id, tanggal, status berwarna, total), dan daftar alamat (label, penerima, alamat lengkap, badge "Default") |

**Status:** ✅ Selesai & terverifikasi manual dengan data nyata (1
customer, 1 order, 1 alamat) — modal menampilkan ketiganya dengan
benar, termasuk network request `GET
/admin/customers/:id/addresses` dan `GET /admin/orders?customer_email=`
yang terkonfirmasi lewat network tab browser.

## 7. Bug ditemukan saat verifikasi (belum diperbaiki)

Saat mengisi data uji untuk Reviews (lihat bagian 8), ditemukan bug
tidak terkait pekerjaan di atas: halaman `src/app/admin/reviews/page.tsx`
selalu menampilkan ★★★★★ (5 bintang) untuk semua review, padahal
`GET /admin/reviews` sudah mengembalikan `rating` yang benar (dicek
lewat `curl`, nilainya bervariasi 2-5). Murni bug rendering di
frontend, bukan masalah data. Sudah di-spawn sebagai task terpisah,
belum dikerjakan.

## 8. Data uji

Untuk memverifikasi dan mengisi tampilan admin panel, ditambahkan data
contoh lewat API (bukan lewat DB langsung, supaya jalur validasi
backend ikut teruji):

- 5 customer (`auth/register`) dengan beberapa order berstatus `paid`
- 5 newsletter subscriber dengan variasi source (`footer`/`checkout`/
  `manual`) dan status (`active`/`unsubscribed`)
- 6 review produk dengan rating 2-5, campuran status `approved`/
  `rejected`/`pending` untuk mendemokan alur moderasi
- 5 kupon dengan tipe berbeda (`percentage`/`fixed`/`free_shipping`),
  termasuk satu dengan `max_uses` dan satu nonaktif

Data uji dari sesi verifikasi pagination (subscriber `sub1-3@test.com`,
customer `customer1@test.com`) sudah dibersihkan lagi setelah
diverifikasi. Data contoh di atas (nama-nama Indonesia) sengaja
dibiarkan untuk keperluan demo admin panel.

## 9. Yang belum dikerjakan (follow-up)

- **Campaign email sending** — tab Campaign di Newsletter masih 100%
  mock, sudah ditandai jelas di UI (banner kuning + tombol "Compose
  Campaign" disabled). Butuh keputusan provider email (SMTP/SendGrid/
  Resend/dll), tabel `campaigns` baru, dan worker pengiriman. Sengaja
  tidak dikerjakan sesi ini.
- **Pagination & filter status untuk tabel list Customer** —
  `listCustomers()` masih hardcode `per_page=50` tanpa kontrol
  halaman. Pola yang sama seperti perbaikan Subscribers di bagian 5
  bisa dipakai ulang (`apiFetchPaginated` sudah tersedia).
- **Bug star rating di admin Reviews** — lihat bagian 7.
- **Bug `checkout`/`popup`/`blog` sebagai `SubscriberSource` yang tidak
  pernah dipakai** — cuma `footer` dan `manual` yang benar-benar
  diproduksi di storefront/admin. Tidak ditangani sesi ini karena
  menambah opt-in newsletter di checkout adalah perubahan scope
  terpisah (halaman checkout, bukan admin panel).

## Ringkasan status

| Bagian | Status |
|---|---|
| Fix kolom `role` hilang di `Update()` | ✅ Selesai |
| Fix N+1 query `ListCustomers` (bulk stats) | ✅ Selesai & terverifikasi |
| Endpoint filter order per `customer_email` | ✅ Selesai & terverifikasi |
| Endpoint `GET /admin/customers/:id/addresses` | ✅ Selesai & terverifikasi |
| Helper `apiFetchPaginated` di frontend | ✅ Selesai |
| Newsletter Subscribers — pagination & filter server-side | ✅ Selesai & terverifikasi |
| Customer — modal detail (profil/order/alamat) | ✅ Selesai & terverifikasi |
| Data uji untuk demo admin panel | ✅ Ditambahkan |
| Campaign email sending | ⏳ Belum dikerjakan — masih mock |
| Pagination & filter tabel list Customer | ⏳ Belum dikerjakan |
| Bug star rating admin Reviews | ⏳ Belum dikerjakan — sudah di-spawn task terpisah |
