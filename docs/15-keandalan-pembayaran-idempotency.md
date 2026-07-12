# 15 — Keandalan Pembayaran: Idempotency, Race Condition, Stok, Cek Status Manual

Dokumen ini merangkum perbaikan keandalan pada alur pembayaran & inventori
Komerce (RajaOngkir Payment Service) setelah ditemukan beberapa celah
nyata: order bisa tercipta dobel kalau checkout di-submit dua kali,
notifikasi "paid" bisa dobel kalau webhook Komerce mengirim event lebih
dari sekali, tidak ada cara memulihkan status kalau webhook tidak pernah
sampai (localhost dev, atau server down saat webhook dikirim), dan stok
produk tidak pernah berkurang dari order sama sekali (overselling mungkin
terjadi). Kode lengkap tidak disalin ulang di sini — cukup rujuk path file
dan status tiap bagian.

## Kronologi masalah

Order `debe2cdd...` (Wahyudin Saudin, Rp 7.637.000, VA BCA) sudah "Paid" di
halaman hosted Komerce, tapi status di admin panel tetap "Belum Dibayar".
Penyebabnya: webhook Komerce (`POST /api/v1/webhooks/komerce/payment`) tidak
bisa reach `localhost:8080` dari server Komerce di internet — bukan bug,
memang begitu cara kerja webhook (butuh URL publik). Diskusi ini kemudian
membuka tiga pertanyaan lanjutan yang dijawab lewat perubahan di dokumen
ini: bagaimana memulihkan status tanpa webhook, dan apakah kode yang ada
sudah aman dari race condition & double payment.

## 1. Cek Status Pembayaran manual (fallback tanpa webhook)

`komercepay.Client.GetStatus()` ([pkg/komercepay/client.go](../pkg/komercepay/client.go))
sudah ada sejak integrasi awal tapi belum pernah dipakai di mana pun. Ini
dijadikan endpoint admin yang tanya langsung ke Komerce, tidak menunggu
webhook.

| File | Perubahan |
|---|---|
| [internal/usecase/payment_usecase.go](../internal/usecase/payment_usecase.go) | `CheckStatus(ctx, orderID)` — ambil order, panggil `GetStatus` ke Komerce pakai `order.PaymentID`, kalau hasilnya `PAID` tandai order paid |
| [internal/delivery/http/handler/payment_handler.go](../internal/delivery/http/handler/payment_handler.go) | `CheckStatus` handler, map `ErrNoPaymentToCheck`/`ErrPaymentUnavailable`/`ErrOrderNotFound` ke status HTTP yang sesuai |
| [internal/delivery/http/router/router.go](../internal/delivery/http/router/router.go) | `POST /admin/orders/:id/check-payment-status` (admin-only) |
| `src/entities/order/api/order.api.ts` | `checkPaymentStatus(id)` |
| `src/app/admin/orders/page.tsx` | Tombol "🔄 Cek Status Pembayaran ke Komerce" di modal detail order, muncul kalau `payment_status !== "paid"` |

**Kapan dipakai:** develop di localhost (webhook memang tidak bisa nyampe),
atau production kalau ada delivery webhook yang hilang/telat.

**Status:** ✅ Selesai & **diverifikasi dengan data order sungguhan** —
order `debe2cdd` (contoh kasus di atas) berhasil direkonsiliasi dari
`unpaid` → `paid` lewat tombol ini, terkonfirmasi lewat response API
(`payment_status: "paid"`) dan tampilan list order ("Sudah Dibayar").

## 2. Race condition di webhook — check-then-act → atomic update

**Sebelum:** `HandleKomerceCallback` baca status order, cek `if paid return`,
baru `UpdatePaymentStatus`. Kalau dua delivery webhook untuk order yang sama
datang nyaris bersamaan (lazim — kebanyakan payment gateway retry webhook
sampai dapat `200 OK`), keduanya bisa lolos pengecekan "belum paid" sebelum
salah satu sempat nulis → `NotifyOrderPaid` terpanggil dua kali → notifikasi
admin dobel.

**Sesudah:** satu `UPDATE ... WHERE payment_status != 'paid'` atomic,
`RowsAffected` menentukan apakah *request ini* yang benar-benar mengubah
status — hanya request itu yang boleh mengirim notifikasi.

| File | Perubahan |
|---|---|
| [internal/repository/order_repository.go](../internal/repository/order_repository.go) | `MarkPaidIfUnpaid(ctx, id) (bool, error)` ditambah ke interface |
| [internal/infrastructure/mysql/order_repo.go](../internal/infrastructure/mysql/order_repo.go) | Implementasi: `UPDATE orders SET payment_status='paid' WHERE id=? AND payment_status != 'paid'`, balikin `affected > 0` |
| [internal/usecase/payment_usecase.go](../internal/usecase/payment_usecase.go) | `HandleKomerceCallback` dan `CheckStatus` sama-sama pakai `MarkPaidIfUnpaid` — notifikasi cuma terkirim kalau `changed == true` |

**Status:** ✅ Selesai, `go build ./...` sukses.

## 3. Idempotency key di pembuatan order

**Sebelum:** `CreateOrder` generate `uuid.New()` baru tiap dipanggil, tanpa
cara mendeteksi request duplikat dari client. Double-klik tombol "Buat
Pesanan", retry jaringan, atau multi-tab bisa bikin **2 order + 2 invoice
pembayaran terpisah** untuk transaksi yang sama.

**Sesudah:** frontend generate satu `idempotency_key` (`crypto.randomUUID()`)
sekali per mount `CheckoutForm`, dipakai ulang di setiap retry submit dalam
sesi checkout yang sama. Backend cek dulu apakah key itu sudah pernah
dipakai sebelum memproses apa pun (hitung harga, panggil payment gateway,
dll) — kalau sudah ada, order yang lama langsung dikembalikan, bukan bikin
baru.

| File | Perubahan |
|---|---|
| [migrations/000011_add_order_idempotency_key.up.sql](../migrations/000011_add_order_idempotency_key.up.sql) | `orders.idempotency_key VARCHAR(100) NULL` + `UNIQUE KEY` (MySQL izinkan banyak `NULL` di unique key, jadi order lama tanpa key tetap valid) |
| [internal/entity/order.go](../internal/entity/order.go) | `Order.IdempotencyKey *string` |
| [internal/repository/order_repository.go](../internal/repository/order_repository.go) + [internal/infrastructure/mysql/order_repo.go](../internal/infrastructure/mysql/order_repo.go) | `FindByIdempotencyKey`; `Create` sisipkan kolom baru, deteksi MySQL error 1062 (duplicate key) → `repository.ErrDuplicateEntry` |
| [internal/usecase/order_usecase.go](../internal/usecase/order_usecase.go) | `CreateOrderInput.IdempotencyKey`; cek di awal `CreateOrder` (return order lama kalau ketemu); kalau `Create` gagal karena `ErrDuplicateEntry` (dua request *benar-benar* bersamaan lolos pre-check yang sama), fallback re-fetch by key alih-alih return error |
| `internal/delivery/http/dto/order_dto.go` | `CreateOrderRequest.IdempotencyKey` (opsional, `max=100`) |
| `src/entities/order/api/order.api.ts` + `src/features/checkout/ui/CheckoutForm.tsx` | `CreateOrderInput.idempotency_key`; `idempotencyKeyRef` (`useRef`, sekali per mount) dikirim di setiap panggilan `createOrder` |

**Kenapa dua lapis (pre-check + UNIQUE constraint), bukan cuma satu:**
pre-check (`FindByIdempotencyKey` sebelum mulai proses) menghindari kerja
sia-sia (hitung ongkir, bikin invoice Komerce) untuk 99.9% kasus double-klik
biasa. `UNIQUE` index di DB adalah jaring pengaman terakhir untuk kasus dua
request yang *benar-benar* nyampe bersamaan di milidetik yang sama, lolos
pre-check yang sama-sama melihat "belum ada" sebelum salah satu commit.

**Status:** ✅ Selesai & **diverifikasi dengan pengujian nyata**:
- Submit sekuensial dengan key sama 2x → order ID kembar persis sama.
- **5 request `POST /orders` ditembak benar-benar bersamaan** (bukan
  berurutan) dengan key yang sama → cuma **1 baris order tercipta di
  database**, ke-5 response HTTP mengembalikan order ID yang sama.

## 4. Reservasi stok saat checkout

**Sebelum:** `CreateOrder` tidak pernah menyentuh `products.stock` — stok
murni field manual yang diedit admin. Dua customer bisa checkout unit
terakhir yang sama secara bersamaan, sistem tidak menolak siapa pun.

**Sesudah:** stok **dikunci begitu order dibuat** (bukan menunggu
pembayaran) — pola "reserved stock" seperti marketplace pada umumnya. Kalau
customer tidak menyelesaikan pembayaran sebelum `payment_expiry`, order
otomatis dibatalkan dan stoknya dikembalikan ke etalase.

| File | Perubahan |
|---|---|
| [internal/repository/product_repository.go](../internal/repository/product_repository.go) + [internal/infrastructure/mysql/product_repo.go](../internal/infrastructure/mysql/product_repo.go) | `DecrementStockIfAvailable(ctx, productID, qty) (bool, error)` — `UPDATE products SET stock = stock - ? WHERE id=? AND stock >= ?`, gagal (`false`) tanpa minus kalau stok kurang; `RestoreStock(ctx, productID, qty)` — kebalikannya |
| [internal/usecase/order_usecase.go](../internal/usecase/order_usecase.go) | `CreateOrder`: setelah item dihitung, reservasi stok per item satu-satu. Kalau ada item yang stoknya kurang → `ErrInsufficientStock` (pesan sertakan nama produk), dan item-item yang *sudah* kepotong di request yang sama langsung di-restore (rollback kompensasi lewat `defer` + flag `orderCommitted`) — bukan cuma item yang gagal, tapi juga penciptaan invoice pembayaran Komerce & `orderRepo.Create` yang gagal setelahnya |
| [internal/delivery/http/handler/order_handler.go](../internal/delivery/http/handler/order_handler.go) | `ErrInsufficientStock` → HTTP `409 Conflict` dengan pesan jelas, otomatis muncul di frontend lewat `HttpError` yang sudah dipakai `CheckoutForm.tsx` (**tidak ada perubahan frontend diperlukan**) |
| [internal/repository/order_repository.go](../internal/repository/order_repository.go) + [internal/infrastructure/mysql/order_repo.go](../internal/infrastructure/mysql/order_repo.go) | `FindExpiredUnpaidOrders(ctx, now)` — ambil order `unpaid` yang `payment_expiry`-nya (disimpan sebagai string RFC3339, di-parse & dibandingkan di Go, bukan di SQL, supaya tidak bergantung pada perbandingan string yang rapuh) sudah lewat; `ExpireIfUnpaid(ctx, id)` — atomic conditional update sama seperti `MarkPaidIfUnpaid`, supaya order yang *kebetulan* dibayar tepat saat sweep jalan tidak ke-cancel keliru |
| [internal/usecase/order_usecase.go](../internal/usecase/order_usecase.go) | `ReleaseExpiredOrders(ctx) (int, error)` — orkestrasi: cari order kedaluwarsa → **`reconcileIfActuallyPaid`** (tanya langsung ke Komerce, bukan cuma percaya webhook yang mungkin hilang — kalau ternyata sudah dibayar, tandai `paid` & lewati, jangan di-expire) → `ExpireIfUnpaid` → restore stok tiap item → `UpdateStatus(cancelled)` |
| [cmd/api/main.go](../cmd/api/main.go) | `runExpiredOrderSweep` — goroutine `time.Ticker` tiap 5 menit, jalan sekali juga di startup (tidak nunggu interval pertama); `OrderUsecase` sekarang juga menerima `komercepay.Client` buat keperluan `reconcileIfActuallyPaid` |

**Kenapa sweep juga tanya ke Komerce, bukan cuma andalkan waktu:**
kalau cuma "waktu habis = expired", order yang **beneran dibayar** tapi
webhook-nya hilang (server down, jaringan putus, dll — persis skenario yang
memicu seluruh dokumen ini) bisa salah dibatalkan dan stoknya dilepas lagi
ke etalase padahal sudah terjual. `reconcileIfActuallyPaid` menutup celah
itu — setiap kandidat kedaluwarsa dicek dulu ke Komerce sebelum
benar-benar dibatalkan. Ini otomatis juga menutup celah "tidak ada
reconciliation job otomatis" yang tadinya masuk daftar belum-dikerjakan.

**Kenapa reservasi di `CreateOrder`, bukan nunggu `paid`:** kalau baru
dikurangi saat `paid`, dua order `pending` untuk unit terakhir yang sama
bisa *dua-duanya* lanjut dibayar customer dan lolos — overselling di
titik paling nyata. Mengunci stok begitu order dibuat menutup celah itu;
konsekuensinya butuh mekanisme pelepasan otomatis (sweep di atas) supaya
stok tidak "hilang" selamanya kalau customer batal bayar.

**Status:** ✅ Selesai & **diverifikasi dengan data produk sungguhan**
(`Shirobana Sujihiki 240mm`, stok awal 25):
- Checkout 1 unit → stok `25 → 24`.
- Checkout 999 unit (melebihi stok) → **HTTP 409**,
  `"insufficient stock: Shirobana Sujihiki 240mm"`, stok **tidak berubah**.
- Order 2 item (item pertama cukup stok, item kedua melebihi) → seluruh
  order ditolak, **item pertama yang sempat kepotong ikut di-restore**
  (rollback kompensasi teruji, bukan cuma diasumsikan).
- Order dengan `payment_expiry` dimundurkan manual ke masa lalu → restart
  backend (memicu sweep) → log `"released":1`, order jadi
  `status=cancelled, payment_status=expired`, stok **kembali ke 25**.
- Order yang **sudah `paid`** dengan `payment_expiry` juga di masa lalu →
  sweep berjalan tapi **tidak menyentuhnya** (query awal `WHERE
  payment_status='unpaid'` sudah menyaring, dan `ExpireIfUnpaid` jadi
  jaring pengaman kedua) — stok tetap terkurangi karena unitnya memang
  terjual sah.
- Order sungguhan yang **belum benar-benar dibayar di Komerce** (bukan
  cuma di database kita), `payment_expiry` dimundurkan → sweep memanggil
  `GetStatus` ke Komerce beneran (bukan mock), dapat jawaban "belum bayar",
  lanjut expire seperti biasa — membuktikan `reconcileIfActuallyPaid`
  tidak salah menahan order yang memang belum dibayar.
- *Catatan jujur:* skenario "order beneran sudah dibayar di Komerce tapi
  webhook hilang" tidak sempat diuji end-to-end (perlu benar-benar
  menyelesaikan pembayaran lewat halaman hosted Komerce secara manual,
  tidak bisa disimulasikan lewat API) — tapi kedua komponen yang
  membentuknya (`GetStatus` mendeteksi "PAID" dengan benar, dan
  `MarkPaidIfUnpaid` menandai order dengan benar) sudah teruji terpisah
  di bagian 1.

## 5. Kegagalan post-commit tidak lagi dilaporkan sebagai error

**Sebelum:** setelah `orderRepo.Create` sukses (order, invoice pembayaran,
dan reservasi stok semuanya sudah beneran ada), kalau `IncrementUsage`
kupon gagal, `CreateOrder` tetap `return nil, err`. Customer lihat
"Checkout failed" padahal order-nya sah — membingungkan, walau tidak
sampai duplikat/kehilangan data berkat idempotency key (bagian 3).

**Sesudah:** titik commit (`orderCommitted = true`, tepat setelah
`orderRepo.Create` sukses) dijadikan garis tegas — apa pun yang gagal
*setelah* itu di-log sebagai warning/error server-side, bukan lagi
menggagalkan response ke client.

| File | Perubahan |
|---|---|
| [internal/usecase/order_usecase.go](../internal/usecase/order_usecase.go) | `OrderUsecase` sekarang menerima `zerolog.Logger`; `IncrementUsage` yang gagal → `log.Error()` + tetap `return order, nil`; `NotifyOrderCreated` yang gagal (sebelumnya `_ = ...` diam-diam) → sekarang juga di-log sebagai `log.Warn()`, bukan cuma dibuang |
| [cmd/api/main.go](../cmd/api/main.go) | `NewOrderUsecase(...)` diberi `log` (instance yang sama dipakai middleware & sweep) |

**Status:** ✅ Selesai. Diverifikasi lewat regresi jalur normal (order +
kupon `HEMAT50K` → sukses, diskon Rp 50.000 terhitung benar, `used_count`
naik jadi 1) — memastikan refactor ini tidak mengubah perilaku jalur
sukses. Jalur kegagalan (`IncrementUsage` gagal tepat setelah `Create`
sukses) **tidak** diuji dengan fault-injection sungguhan — perlu
mem-break DB di detik yang presisi, di luar proporsi untuk perubahan
sesederhana ini — tapi perubahan logikanya sendiri (mengganti `return nil,
err` jadi `log.Error()` lalu lanjut) cukup jelas benar lewat pembacaan
kode & `go vet` yang bersih.

> **Sudah ditutup di bagian 4:** "tidak ada reconciliation job otomatis"
> dan "`payment_status` tidak otomatis jadi `expired`" — dua-duanya
> ditangani sekaligus oleh `runExpiredOrderSweep` +
> `reconcileIfActuallyPaid`.

## 6. Celah lain yang ditemukan (belum dikerjakan — untuk diputuskan)

| Celah | Dampak | Rekomendasi |
|---|---|---|
| **Race kecil di penggunaan kupon.** `ValidateCoupon` (baca `used_count < max_uses`) dan `IncrementUsage` (`used_count = used_count + 1`, atomic di level SQL) adalah dua operasi terpisah, bukan satu transaksi. Beberapa order dengan kupon yang sama, tepat di ambang batas `max_uses`, bisa lolos validasi bersamaan sebelum salah satu increment. | Sangat rendah risiko (butuh banyak request bersamaan tepat di detik yang sama, tepat di angka batas) dan dampaknya cuma kupon kepakai sedikit lebih banyak dari `max_uses`, bukan kebocoran data/uang. | Bisa dibiarkan; kalau mau ditutup total, ganti jadi `UPDATE coupons SET used_count = used_count + 1 WHERE id=? AND used_count < max_uses` + cek `RowsAffected` sebelum `CreateOrder` melanjutkan. |

## Ringkasan status

| Bagian | Status |
|---|---|
| Migration `orders.idempotency_key` | ✅ Diterapkan ke database lokal & diverifikasi |
| Cek Status Pembayaran manual (usecase + endpoint + tombol admin) | ✅ Selesai & diverifikasi dengan order sungguhan |
| Fix race condition webhook (`MarkPaidIfUnpaid`) | ✅ Selesai |
| Idempotency key order creation (pre-check + UNIQUE constraint) | ✅ Selesai & diverifikasi lewat 5 request paralel sungguhan |
| Reservasi stok saat checkout (`DecrementStockIfAvailable`) | ✅ Selesai & diverifikasi dengan produk sungguhan |
| Auto-cancel order kedaluwarsa + restore stok (sweep 5 menit) | ✅ Selesai & diverifikasi (termasuk race guard `ExpireIfUnpaid`) |
| Reconciliation ke Komerce sebelum expire (`reconcileIfActuallyPaid`) | ✅ Selesai & diverifikasi jalur "belum bayar"; jalur "sudah bayar tapi webhook hilang" teruji lewat komponen penyusunnya (bagian 1), bukan end-to-end |
| Kegagalan post-commit tidak lagi dilaporkan gagal ke customer | ✅ Selesai & diverifikasi lewat regresi jalur normal (order + kupon) |
| Race kecil pemakaian kupon di ambang `max_uses` | ⏳ Belum — risiko sangat rendah, opsional |
