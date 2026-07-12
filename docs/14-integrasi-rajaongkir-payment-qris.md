# 14 — Integrasi RajaOngkir (Ongkir + Payment Service + QRIS)

Dokumen ini merangkum integrasi tiga API dari ekosistem Komerce/RajaOngkir
(https://rajaongkir.com/docs): **Cek Ongkir Realtime**, **Payment Service**
(VA/bank transfer), dan **QRIS** — dipakai bareng di alur checkout
storefront (hitung ongkir per kurir + pilih metode bayar). Kode lengkap
tidak disalin ulang di sini — cukup rujuk path file dan status tiap bagian.

## Cakupan & keputusan desain

- **Ongkir dihitung dari berat produk asli di database**, bukan dari
  input client — pola yang sama dengan harga di modul Order
  ([11-orders-reports-excel-pdf.md](11-orders-reports-excel-pdf.md)).
  Kolom `products.weight` (gram) ditambahkan lewat migration, default
  500g untuk produk yang belum diisi manual.
- **QRIS dan bank transfer (VA) dilayani lewat satu client yang sama**
  (`pkg/komercepay`), bukan client terpisah — API Payment Service
  Komerce menerima `payment_type: "qris" | "bank_transfer"` di endpoint
  create-payment yang sama. Config `KOMERCE_QRISLY_*` (base URL, API
  key, callback key, `qris_id`) sudah disiapkan di `pkg/config` untuk
  skenario QRIS statis/terpisah, **tapi belum dipakai** — lihat bagian
  "Yang belum dikerjakan" di bawah.
- **Graceful degradation tanpa API key.** Baik `rajaongkir.Client`
  maupun `komercepay.Client` punya method `Enabled()` (`true` kalau
  `apiKey != ""`). Kalau key kosong: pencarian tujuan & cek ongkir
  balas `503 shipping service is not configured`, dan pembuatan order
  otomatis jatuh ke `DummyGateway` (dev tanpa payment sungguhan) —
  bukan checkout yang error total.
- **Webhook memverifikasi HMAC-SHA256 atas raw body**, bukan
  percaya begitu saja payload JSON — signature dibaca dari header
  `X-Callback-Api-Key` dan dibandingkan pakai `hmac.Equal` (constant-time).

## 1. Migration

| File | Isi |
|---|---|
| [migrations/000009_add_weight_and_shipping.up.sql](../migrations/000009_add_weight_and_shipping.up.sql) | `products.weight` (gram, default 500); `orders.shipping_cost`, `shipping_courier`, `shipping_service`, `shipping_etd`, `destination_id` |

**Status:** ✅ Dijalankan ke database lokal.

> Catatan: kolom `payment_type`/`payment_channel`/`payment_va_number`/
> `payment_qr_string`/`payment_url`/`payment_expiry` di `entity.Order`
> dipetakan lewat migration order yang sudah ada sebelumnya — cek
> `internal/entity/order.go` kalau perlu menambah kolom baru di area ini.

## 2. Config & environment

| File | Isi |
|---|---|
| [pkg/config/config.go](../pkg/config/config.go) | `RajaOngkirConfig` (BaseURL/APIKey/OriginID), `KomercePaymentConfig` (BaseURL/APIKey/CallbackKey), `QrislyConfig` (BaseURL/APIKey/CallbackKey/QrisID) |
| [.env.example](../.env.example) | `RAJAONGKIR_*`, `KOMERCE_PAYMENT_*`, `KOMERCE_QRISLY_*` |

**Variabel yang wajib diisi untuk mengaktifkan tiap fitur:**

| Fitur | Variabel |
|---|---|
| Cek ongkir | `RAJAONGKIR_API_KEY`, `RAJAONGKIR_ORIGIN_ID` (ID lokasi toko, didapat dari endpoint search domestic-destination) |
| Payment VA/QRIS | `KOMERCE_PAYMENT_API_KEY`, `KOMERCE_PAYMENT_CALLBACK_KEY` (buat verifikasi webhook) |

**Status:** ✅ Selesai. API key **test** yang diberikan untuk sesi
development disimpan di `.env` lokal (tidak di-commit — lihat bagian
gitignore di bawah), bukan di `.env.example`.

## 3. Backend — Shipping (RajaOngkir)

| File | Peran |
|---|---|
| [pkg/rajaongkir/client.go](../pkg/rajaongkir/client.go) | HTTP client murni: `SearchDomesticDestination`/`SearchInternationalDestination` (header `key`), `CalculateDomesticCost`/`CalculateInternationalCost` (form-encoded POST, `price=lowest`) |
| [internal/usecase/shipping_usecase.go](../internal/usecase/shipping_usecase.go) | `SearchDestinations`, `CalculateOptions` — total berat dihitung dari `product.Weight * quantity` per item (query katalog asli), minimum 1000g |
| [internal/delivery/http/dto/shipping_dto.go](../internal/delivery/http/dto/shipping_dto.go) + [internal/delivery/http/handler/shipping_handler.go](../internal/delivery/http/handler/shipping_handler.go) | Request/response mapping + handler |

**Endpoint:**

| Endpoint | Akses |
|---|---|
| `GET /api/v1/shipping/destinations?search=` | Publik, min. 3 karakter (di bawah itu balas array kosong tanpa panggil API) |
| `POST /api/v1/shipping/cost` | Publik — body `{destination_id, items:[{product_slug, quantity}]}` |

**Status:** ✅ Selesai, `go build ./...` sukses.

## 4. Backend — Payment (VA + QRIS)

| File | Peran |
|---|---|
| [pkg/komercepay/client.go](../pkg/komercepay/client.go) | HTTP client: `GetMethods` (daftar metode aktif), `CreatePayment` (VA/QRIS, header `x-api-key`), `GetStatus` |
| [internal/usecase/payment_gateway.go](../internal/usecase/payment_gateway.go) | Interface `PaymentGateway.CreateInvoice` — abstraksi supaya `OrderUsecase` tidak tahu vendor spesifik |
| [internal/infrastructure/payment/komerce_gateway.go](../internal/infrastructure/payment/komerce_gateway.go) | Implementasi asli — fallback ke `DummyGateway` kalau client belum `Enabled()` atau `payment_type` kosong |
| [internal/usecase/payment_usecase.go](../internal/usecase/payment_usecase.go) | `GetMethods`, `HandleKomerceCallback` (verifikasi HMAC → tandai order `paid` secara idempotent) |
| [internal/delivery/http/dto/payment_dto.go](../internal/delivery/http/dto/payment_dto.go) + [internal/delivery/http/handler/payment_handler.go](../internal/delivery/http/handler/payment_handler.go) | Request/response mapping + handler |

**Endpoint:**

| Endpoint | Akses |
|---|---|
| `GET /api/v1/payment/methods` | Publik — balas array kosong (bukan error) kalau gateway belum dikonfigurasi |
| `POST /api/v1/webhooks/komerce/payment` | Publik, diverifikasi via header `X-Callback-Api-Key` |

**Alur create order dengan pembayaran** (`internal/usecase/order_usecase.go`):
1. Hitung ulang subtotal dari katalog + ongkir dari RajaOngkir (kalau
   `destination_id`/`courier`/`service` diisi).
2. Panggil `PaymentGateway.CreateInvoice` dengan total akhir.
3. Simpan `payment_type`, `payment_channel`, `payment_va_number`,
   `payment_qr_string`, `payment_url`, `payment_expiry` ke baris order.
4. Webhook Komerce memanggil `/webhooks/komerce/payment` saat status
   berubah → `payment_status` di-update ke `paid`, notifikasi admin
   dikirim (`NotifyOrderPaid`).

**Status:** ✅ Selesai, `go build ./...` sukses.

## 5. DI wiring & routes

| File | Perubahan |
|---|---|
| [cmd/api/main.go](../cmd/api/main.go) | Inisialisasi `rajaongkir.New(...)`, `komercepay.New(...)`, pemilihan `DummyGateway` vs `KomercePaymentGateway` berdasar `Enabled()` |
| [internal/delivery/http/router/router.go](../internal/delivery/http/router/router.go) | Registrasi 4 endpoint publik: `/shipping/destinations`, `/shipping/cost`, `/payment/methods`, `/webhooks/komerce/payment` |

**Status:** ✅ Selesai.

## 6. Storefront — checkout

| File | Perubahan |
|---|---|
| [src/entities/shipping/api/shipping.api.ts](../../pisau-pedia/src/entities/shipping/api/shipping.api.ts) | `searchDestinations`, `calculateShippingCost` |
| [src/entities/payment/api/payment.api.ts](../../pisau-pedia/src/entities/payment/api/payment.api.ts) | `getPaymentMethods` |
| [src/entities/order/api/order.api.ts](../../pisau-pedia/src/entities/order/api/order.api.ts) | `CreateOrderInput` ditambah `payment_type`/`payment_channel`; `OrderResponse` ditambah field pembayaran (`payment_type`, `payment_channel`, `payment_va_number`, `payment_qr_string`, `payment_url`, `payment_expiry`) — sebelumnya field ini dikirim/dipakai tapi belum dideklarasikan di tipe TypeScript |
| [src/features/checkout/model/CheckoutProvider.tsx](../../pisau-pedia/src/features/checkout/model/CheckoutProvider.tsx) | Context untuk `destination`, `shippingOption`, `freeShipping` (dipakai bareng `CheckoutForm` dan ringkasan order) |
| [src/features/checkout/ui/CheckoutForm.tsx](../../pisau-pedia/src/features/checkout/ui/CheckoutForm.tsx) | Autocomplete tujuan (debounce 350ms, min 3 karakter) → pilih kurir (list harga per opsi) → **pilih metode pembayaran** → submit `createOrder` |

**Status:** ✅ Selesai & `npx tsc --noEmit` bersih. Bagian pemilihan
metode pembayaran (radio-style button per metode dari
`GET /payment/methods`, membedakan QRIS vs VA per `bank_code`)
sebelumnya **belum ada di JSX** — form sudah fetch `paymentMethods` dan
punya state `selectedPayment`, tapi tidak ada UI untuk mengisinya,
sehingga submit checkout selalu gagal validasi "Pilih metode
pembayaran terlebih dahulu." Sudah dilengkapi.

> Belum diverifikasi di browser end-to-end (perlu backend Go + MySQL +
> `.env` berisi API key asli berjalan bareng; saat pengecekan, port
> 3000 lokal terpakai proyek lain). Verifikasi manual disarankan:
> jalankan `docker-compose up` + backend + `npm run dev`, lalu tambah
> produk ke keranjang → checkout → pastikan daftar kurir & metode
> pembayaran muncul dan order berhasil dibuat.

## 7. Yang belum dikerjakan (follow-up)

- **`KOMERCE_QRISLY_*` config belum dipakai.** QRIS saat ini jalan
  lewat `payment_type: "qris"` di client `komercepay` yang sama dengan
  VA (pakai `KOMERCE_PAYMENT_API_KEY`). Kalau QRIS statis/terpisah
  (endpoint `upload-qris` + `qris_id` sendiri) memang dibutuhkan
  sebagai jalur berbeda, perlu client baru di `pkg/qrisly` dan
  percabangan di `komerce_gateway.go`.
- **Halaman `/checkout/success` belum menampilkan instruksi bayar**
  (nomor VA / string QRIS) — saat ini redirect ke `order.invoice_url`
  yang untuk gateway Komerce asli adalah halaman hosted milik Komerce
  (jadi instruksi bayar tetap muncul, hanya di luar domain sendiri).
  Kalau nanti mau tampilan custom, data sudah tersedia di
  `OrderResponse.payment_va_number` / `payment_qr_string`.
- **Cek ongkir internasional** (`SearchInternationalDestination`,
  `CalculateInternationalCost`) sudah ada di client Go tapi belum
  dipakai di usecase/handler manapun — baru domestik yang jalan.

## Ringkasan status

| Bagian | Status |
|---|---|
| Migration `weight` & kolom shipping di `orders` | ✅ Selesai |
| Backend Shipping (RajaOngkir) — search + cost | ✅ Selesai |
| Backend Payment (Komerce) — methods + create + webhook | ✅ Selesai |
| DI wiring & routes | ✅ Selesai |
| Storefront checkout (tujuan → kurir → metode bayar) | ✅ Selesai |
| Verifikasi end-to-end di browser | ⏳ Belum — perlu backend+DB berjalan |
| QRIS statis via `pkg/qrisly` terpisah | ⏳ Belum dikerjakan (opsional) |
| Instruksi bayar custom di `/checkout/success` | ⏳ Belum dikerjakan (opsional) |
| Cek ongkir internasional dipakai di endpoint | ⏳ Belum dikerjakan |
