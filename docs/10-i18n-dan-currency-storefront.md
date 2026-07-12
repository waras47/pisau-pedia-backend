# 10 — Bahasa (EN/ID) & Mata Uang (USD/IDR) di Storefront

Dokumen ini merangkum penambahan dua bahasa (Inggris/Indonesia) dan
konversi mata uang real-time (USD/IDR) di frontend `pisau-pedia`.
Kode lengkap tidak disalin ulang di sini — cukup rujuk path file dan
status tiap bagian.

## Cakupan & keputusan desain

- **Hanya storefront** (`src/app/(store)/*` dan widget/feature yang
  dipakai di situ) — **admin panel sengaja tidak disentuh**, tetap
  Bahasa Indonesia saja sesuai permintaan.
- Bahasa = toggle client-side (state + `localStorage`), **bukan**
  routing `/en/`, `/id/` — polanya sama seperti `ThemeToggle` yang
  sudah ada (`next-themes`), supaya konsisten dengan codebase dan
  tidak perlu restrukturisasi route.
- Pilihan bahasa **otomatis menentukan mata uang**: EN → USD, ID →
  IDR. Default: **ID / IDR** (tidak perlu aksi apa pun dari user).
- Kurs USD⇄IDR diambil dari backend (proxy), bukan dipanggil langsung
  dari browser — supaya bisa di-cache di server dan tidak setiap
  browser user memanggil API luar sendiri-sendiri.
- Konversi dirancang **generik** (dari currency asli produk ke
  currency target), bukan asumsi satu currency tetap — supaya tetap
  benar baik untuk data mock storefront (saat ini `EUR`) maupun nanti
  kalau storefront sudah pakai data asli dari backend (`IDR`).

## 1. Backend — endpoint kurs

| File | Isi |
|---|---|
| [pkg/exchangerate/client.go](../pkg/exchangerate/client.go) | Fetch `https://open.er-api.com/v6/latest/USD` (API publik, tanpa key), cache in-memory 6 jam, fallback ke cache lama kalau API luar gagal |
| [internal/delivery/http/handler/exchange_rate_handler.go](../internal/delivery/http/handler/exchange_rate_handler.go) | `GET /exchange-rate` → `{"base":"USD","rates":{...},"updated_at":"..."}` |
| [internal/delivery/http/router/router.go](../internal/delivery/http/router/router.go) | Route publik (tanpa auth) `v1.GET("/exchange-rate", ...)` |
| [cmd/api/main.go](../cmd/api/main.go) | Wiring `exchangerate.New()` → `handler.NewExchangeRateHandler` → `router.Dependencies` |

Rumus konversi (dipakai di frontend, basis kurs selalu USD dari API):
```
amountTarget = amount / rates[fromCurrency] * rates[toCurrency]
```

**Status:** ✅ Selesai, `go build ./...` sukses. Sudah dicoba lewat
`curl` (404 sebelum restart backend — perlu `make run` ulang setelah
kode ini ditambahkan, karena `go run` tidak hot-reload).

## 2. Frontend — plumbing bahasa & currency

| File | Peran |
|---|---|
| `src/shared/i18n/dictionaries.ts` | Kamus teks UI EN/ID (cart, checkout, subtotal, sold out, dst) |
| `src/features/locale-currency/model/LocaleProvider.tsx` | Context: `locale`, `currency` (turunan otomatis), `setLocale`, `t(key)`, `formatPrice(amount, sourceCurrency)`. Fetch `/exchange-rate` sekali saat mount, simpan pilihan ke `localStorage` (`pp_locale`), default `"id"` |
| `src/features/locale-currency/ui/LocaleToggle.tsx` | Tombol switch "ID / EN" di header, pola sama seperti `ThemeToggle` |
| `src/app/(store)/layout.tsx` | Dibungkus `<LocaleProvider>`, **hanya route group `(store)`** — admin panel di luar cakupan ini |
| `src/widgets/header/Header.tsx` | Tambah `<LocaleToggle />` di sebelah `<ThemeToggle />` |

**Status:** ✅ Selesai, terverifikasi lewat browser otomatis: default
render `Rp 248`/"Stok Habis", klik toggle → instan berubah jadi
`$248.00`/"Sold Out", klik lagi → balik ke Rupiah. `npx tsc --noEmit`
sukses tanpa error.

## 3. Migrasi pemanggil `formatPrice` (7 file)

Sebelumnya `formatPrice(amount, currency)` adalah fungsi statis
(`entities/product/lib/format-price.ts`, locale `de-DE` di-hardcode).
Sekarang diganti jadi hasil hook `useLocaleCurrency()` yang locale
& currency-aware. File lama sudah **dihapus** (tidak dipakai lagi).

| File | Perubahan |
|---|---|
| `src/entities/product/ui/ProductCard.tsx` | Server → **client component** (butuh hook); teks "No reviews yet"/"Sold Out" ikut dipetakan ke `t()` |
| `src/widgets/monthly-pick/MonthlyPick.tsx` | Server → **client component**; teks "Knife of the month"/"Shop This Knife"/"Save" ikut `t()` |
| `src/widgets/product-detail/ProductDetail.tsx` | Server → **client component** |
| `src/features/knife-configurator/ui/KnifeConfigurator.tsx` | Sudah client; ada 2 tempat pemanggilan (`KnifeConfigurator` & sub-komponen `OptionCard`) — keduanya perlu instance hook masing-masing |
| `src/features/cart/ui/CartDrawer.tsx` | Sudah client; teks cart kosong/checkout/subtotal ikut `t()` |
| `src/app/(store)/cart/page.tsx` | Sudah client; sama seperti di atas |
| `src/app/(store)/checkout/page.tsx` | Sudah client; sama seperti di atas |

**Catatan trade-off:** 3 komponen yang tadinya server component
(`ProductCard`, `MonthlyPick`, `ProductDetail`) sekarang jadi client
component karena butuh akses ke context React. `ProductCard` dipakai
berulang di grid produk — bundle JS jadi sedikit lebih besar, tapi ini
trade-off yang wajar untuk komponen harga yang memang perlu interaktif
(currency toggle real-time tanpa reload halaman).

## 4. Cakupan penerjemahan teks (apa yang tercakup, apa yang belum)

**Sudah diterjemahkan** (string UI umum yang sering muncul):
cart/checkout label, subtotal, tombol add-to-cart/checkout/continue
shopping, "sold out", "no reviews yet", teks terkait Monthly Pick.

**Belum diterjemahkan** (sengaja di luar cakupan tahap ini): deskripsi
produk, copy marketing (hero, testimonials, trust badges, footer,
nav menu dari `entities/navigation/model/navigation.data.ts`, dll).
Ini konten, bukan UI chrome, dan datanya belum punya varian
bahasa — kalau nanti dibutuhkan, itu pekerjaan konten terpisah yang
lebih besar (perlu field `*_en`/`*_id` di database, bukan cuma
tambahan dictionary key).

## Ringkasan status

| Bagian | Status |
|---|---|
| Endpoint kurs backend (`GET /exchange-rate`) | ✅ Selesai, compile sukses |
| `LocaleProvider` + `LocaleToggle` frontend | ✅ Selesai & terverifikasi |
| Wiring ke storefront (`(store)/layout.tsx`) | ✅ Selesai — admin panel sengaja tidak disentuh |
| Migrasi 7 pemanggil `formatPrice` | ✅ Selesai, `tsc --noEmit` bersih |
| Terjemahan UI chrome umum | ✅ Set awal selesai |
| Terjemahan konten (deskripsi produk, marketing copy) | ⏳ Di luar cakupan, perlu kerja terpisah |

**Yang perlu dilakukan manual:** restart `make run` di backend supaya
route `/exchange-rate` aktif — begitu itu jalan, konversi kurs
sungguhan (bukan cuma reformat label) otomatis kepakai tanpa ubah
kode apa pun lagi.
