# 05 — Modul Product: Urutan Implementasi

Dikerjakan **setelah** modul User & Auth selesai, karena endpoint admin
produk butuh middleware `JWTAuth()` + `RequireRole("admin")` yang sudah
dibangun di modul sebelumnya. Urutan langkah mengikuti pola yang sama
(entity → repository interface → usecase → infrastructure → delivery).

## Langkah 1 — Migration database

Jalankan migration `000002_create_products_and_categories` dari
[02-database-design.md](02-database-design.md) (5 tabel: `categories`,
`products`, `product_images`, `product_specs`, `product_highlights`).

## Langkah 2 — Entity (`internal/entity/`)

Buat `category.go`, `product.go`. Karena produk punya relasi 1-ke-banyak ke
images/specs/highlights, entity `Product` boleh punya field slice untuk
data yang sudah di-JOIN:

```go
type Product struct {
    ID              string
    CategoryID      *string
    Name            string
    Slug            string
    Description     *string
    Price           int64
    CompareAtPrice  *int64
    Currency        string
    Maker           *string
    Badge           *string
    Stock           uint
    RatingAvg       float64
    ReviewCount     uint
    IsActive        bool
    CreatedAt       time.Time
    UpdatedAt       time.Time

    // Diisi hanya saat query detail (GetBySlug), bukan saat listing.
    Images     []ProductImage
    Specs      []ProductSpec
    Highlights []ProductHighlight
}
```

Keputusan ini (field relasi ada di entity, hanya diisi kondisional) dipilih
supaya `usecase.List()` tidak perlu N+1 query untuk listing — cukup query
`products` saja tanpa join ke 3 tabel anak. Detail query dijelaskan di
Langkah 5.

## Langkah 3 — Repository interface (`internal/repository/`)

`product_repository.go`:

- `FindAll(ctx, filter ProductFilter) ([]entity.Product, int64, error)` —
  `ProductFilter` berisi `Page`, `PerPage`, `CategorySlug`, `Search`,
  `Sort`. Return juga total count untuk pagination meta.
- `FindBySlug(ctx, slug string) (*entity.Product, error)` — termasuk
  images/specs/highlights.
- `Create(ctx, *entity.Product) error`
- `Update(ctx, *entity.Product) error`
- `Delete(ctx, id string) error`

`category_repository.go`: `FindAll`, `FindBySlug`, `Create`, `Update`,
`Delete`.

## Langkah 4 — Usecase (`internal/usecase/product_usecase.go`, `category_usecase.go`)

1. **`ListProducts(ctx, filter)`** — validasi/normalisasi filter (default
   `page=1`, `per_page=12`, clamp `per_page` maksimal mis. 50 supaya tidak
   ada yang minta 100000 row sekaligus), delegasikan ke repository,
   hitung `total_pages` dari total count.
2. **`GetProductBySlug(ctx, slug)`** — return `ErrProductNotFound` kalau
   tidak ada atau `is_active = false` (produk nonaktif tidak boleh terlihat
   dari endpoint publik).
3. **`CreateProduct(ctx, input)`** *(admin)* — generate `slug` dari `name`
   kalau tidak dikirim eksplisit (slugify + cek unik, tambahkan suffix
   angka kalau bentrok), generate UUID, `Create` ke repository beserta
   images/specs/highlights dalam **satu transaksi database** (lihat
   Langkah 5 — kalau insert produk sukses tapi insert specs gagal, semua
   harus rollback).
4. **`UpdateProduct(ctx, id, patch)`** *(admin)*, **`DeleteProduct(ctx, id)`**
   *(admin)*.
5. **`ListCategories(ctx)`**, **`CreateCategory`**, **`UpdateCategory`**,
   **`DeleteCategory`** *(admin)* — pola sama, lebih sederhana (tanpa
   relasi anak).

## Langkah 5 — Infrastructure MySQL (`internal/infrastructure/mysql/`)

`product_repo.go` — poin teknis penting:

- **`FindAll`**: satu query `SELECT ... FROM products LEFT JOIN categories
  ...` dengan `WHERE is_active = 1` + filter dinamis (`category`,
  `search` pakai `LIKE '%...%'` pada `name`), `ORDER BY` sesuai `Sort`,
  `LIMIT`/`OFFSET` dari `Page`/`PerPage`. Query kedua terpisah untuk
  `COUNT(*)` dengan filter yang sama (tanpa `LIMIT`), dipakai untuk
  `meta.total`.
- **`FindBySlug`**: 4 query terpisah (product, images, specs, highlights)
  lebih sederhana dibanding satu query JOIN besar yang menghasilkan baris
  duplikat — pilih pendekatan 4 query untuk fase ini demi keterbacaan kode.
- **`Create`**: pakai `db.Beginx()` → insert `products` → insert semua
  `product_images`/`product_specs`/`product_highlights` dalam loop →
  `tx.Commit()`. Kalau ada error di tengah, `tx.Rollback()`.

`category_repo.go` — implementasi standar, tidak ada kompleksitas
tambahan.

## Langkah 6 — DTO (`internal/delivery/http/dto/`)

`ProductListItemResponse` (ringkas, untuk listing — tanpa
specs/highlights/images penuh, sesuai field yang dipakai `ProductCard` di
frontend) vs `ProductDetailResponse` (lengkap, untuk halaman detail).
Pemisahan ini mencegah payload listing membengkak karena ikut mengirim
specs/highlights yang tidak dipakai di grid.

`CreateProductRequest` dengan validasi: `name` required, `price`
required+min=0, `category_id` opsional (`omitempty,uuid`), `images`
opsional array of URL string, dst.

## Langkah 7 — Handler & Middleware

Tidak ada middleware baru — pakai `JWTAuth()` + `RequireRole("admin")` yang
sudah ada dari modul sebelumnya, dipasang di route group `/admin/products`
dan `/admin/categories`. Endpoint publik (`GET /products`, `GET
/products/:slug`, `GET /categories`) **tanpa** middleware auth sama sekali.

## Langkah 8 — Routing (`cmd/api/main.go`, tambahan)

```
v1.GET("/categories", categoryHandler.List)
v1.GET("/products", productHandler.List)
v1.GET("/products/:slug", productHandler.GetBySlug)

admin := v1.Group("/admin", middleware.JWTAuth(), middleware.RequireRole("admin"))
admin.POST("/products", productHandler.Create)
admin.PATCH("/products/:id", productHandler.Update)
admin.DELETE("/products/:id", productHandler.Delete)
admin.POST("/categories", categoryHandler.Create)
admin.PATCH("/categories/:id", categoryHandler.Update)
admin.DELETE("/categories/:id", categoryHandler.Delete)
```

## Langkah 9 — Seed data untuk development

Tambahkan minimal 5-8 produk contoh (pisau berbagai kategori) via migration
seed atau skrip terpisah, supaya frontend bisa langsung dites terhadap data
nyata alih-alih array kosong. Struktur mengikuti contoh 8 produk seed di
`kissaki-backend` (boleh dipakai sebagai referensi konten, tapi field-nya
harus disesuaikan ke skema MySQL di dokumen ini).

## Langkah 10 — Testing manual

1. `GET /api/v1/products` tanpa param → 12 produk pertama + `meta` benar.
2. `GET /api/v1/products?category=chef-knives&page=2` → filter & pagination
   jalan bareng.
3. `GET /api/v1/products?search=santoku` → hanya produk yang match nama.
4. `GET /api/v1/products/:slug` untuk slug tidak ada → `404`.
5. `POST /api/v1/admin/products` **tanpa token** → `401`.
6. `POST /api/v1/admin/products` dengan token `role=customer` → `403`.
7. `POST /api/v1/admin/products` dengan token `role=admin`, payload
   lengkap termasuk `specs`+`highlights` → `201`, lalu `GET
   /products/:slug` menunjukkan specs/highlights tsb.
8. Matikan network di tengah `Create` (atau simulasikan error di specs
   insert) → pastikan produk **tidak** ke-insert setengah jalan (transaksi
   rollback bekerja).

## Checklist keluar dari modul ini

- [ ] Semua 10 langkah selesai
- [ ] Unit test usecase minimal untuk `ListProducts` (pagination math) dan
      slug generation/uniqueness
- [ ] 8 skenario manual di Langkah 10 lulus semua
- [ ] Response `GET /products` dan `GET /products/:slug` dicek manual
      cocok field-nya dengan `Product` type di frontend
      (`product.types.ts`) — tidak ada field yang namanya beda
      (`camelCase` di frontend vs response API perlu konsisten, putuskan
      di DTO apakah backend ikut `camelCase` di JSON tag atau frontend
      yang mapping `snake_case` → `camelCase`; **rekomendasi**: API pakai
      `snake_case` standar REST, frontend tetap yang mapping, karena
      `get-product.ts` di frontend sudah jadi satu titik mapping)

Setelah ini selesai, lanjut ke
[06-observability-prometheus-grafana.md](06-observability-prometheus-grafana.md)
untuk memastikan kedua modul ini termonitor, lalu
[07-docker-and-deployment.md](07-docker-and-deployment.md) untuk
menjalankan semuanya sebagai satu stack.
