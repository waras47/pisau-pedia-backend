# 02 — Desain Database (MySQL)

Skema di bawah mencakup **7 tabel** untuk 2 modul fase ini (User & Auth,
Product). Field produk mengikuti bentuk `Product` yang sudah dipakai
frontend di
[`product.types.ts`](../../pisau-pedia/src/entities/product/model/product.types.ts)
supaya mapping DTO → frontend type nanti tinggal 1:1.

## ERD (ringkas)

```
users ──1───N── addresses
users ──1───N── refresh_tokens

categories ──1───N── products
products ──1───N── product_images
products ──1───N── product_specs
products ──1───N── product_highlights
```

Tidak ada relasi antara `users` dan `products` di fase ini (belum ada
cart/order/review — lihat [08-roadmap-checklist.md](08-roadmap-checklist.md)).

## Konvensi umum

- Primary key: `CHAR(36)` berisi UUID v4, **digenerate di Go**
  (`uuid.New().String()`), bukan di database (alasan: lihat
  [01-architecture-and-tech-stack.md](01-architecture-and-tech-stack.md)).
- Charset/collation: `utf8mb4` / `utf8mb4_unicode_ci` di semua tabel &
  kolom teks, supaya mendukung emoji/karakter non-Latin dan pencarian
  case-insensitive.
- Timestamp: `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`,
  `updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`
  bila kolom itu ada.
- Uang (harga) disimpan sebagai `BIGINT` dalam Rupiah (tanpa desimal),
  konsisten dengan keputusan di `kissaki-backend`.
- Setiap tabel FK memakai `ON DELETE CASCADE` kecuali disebutkan lain.

## Migration 1 — `000001_create_users_and_auth`

```sql
-- +migrate Up
CREATE TABLE users (
    id             CHAR(36)     NOT NULL PRIMARY KEY,
    email          VARCHAR(255) NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    full_name      VARCHAR(150) NOT NULL,
    phone          VARCHAR(20)  NULL,
    role           ENUM('customer', 'admin') NOT NULL DEFAULT 'customer',
    avatar_url     VARCHAR(500) NULL,
    is_active      TINYINT(1)   NOT NULL DEFAULT 1,
    email_verified_at DATETIME NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE addresses (
    id           CHAR(36)     NOT NULL PRIMARY KEY,
    user_id      CHAR(36)     NOT NULL,
    label        VARCHAR(50)  NULL,
    full_name    VARCHAR(150) NOT NULL,
    phone        VARCHAR(20)  NULL,
    address_line VARCHAR(500) NOT NULL,
    city         VARCHAR(100) NOT NULL,
    province     VARCHAR(100) NULL,
    postal_code  VARCHAR(10)  NOT NULL,
    is_default   TINYINT(1)   NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    KEY idx_addresses_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE refresh_tokens (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    user_id     CHAR(36)     NOT NULL,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  DATETIME     NOT NULL,
    revoked_at  DATETIME     NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    KEY idx_refresh_tokens_user_id (user_id),
    UNIQUE KEY uq_refresh_tokens_hash (token_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS users;
```

Catatan desain:

- `token_hash`, bukan token mentah, yang disimpan (hash SHA-256 dari
  refresh token) — supaya kalau tabel bocor, token tidak langsung bisa
  dipakai penyerang. Detail alurnya ada di
  [04-module-user-auth.md](04-module-user-auth.md).
- `email_verified_at` disiapkan sejak awal walau flow verifikasi email
  belum diimplementasikan di fase ini — supaya kolom tidak perlu migration
  tambahan nanti (opsional dipakai, boleh diabaikan dulu dan selalu `NULL`).

## Migration 2 — `000002_create_products_and_categories`

```sql
-- +migrate Up
CREATE TABLE categories (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(120) NOT NULL,
    description TEXT NULL,
    image_url   VARCHAR(500) NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products (
    id                CHAR(36)     NOT NULL PRIMARY KEY,
    category_id       CHAR(36)     NULL,
    name              VARCHAR(200) NOT NULL,
    slug              VARCHAR(220) NOT NULL,
    description       TEXT NULL,
    price             BIGINT UNSIGNED NOT NULL,
    compare_at_price  BIGINT UNSIGNED NULL,
    currency          VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    maker             VARCHAR(150) NULL,
    badge             ENUM('new', 'sale', 'sold-out') NULL,
    stock             INT UNSIGNED NOT NULL DEFAULT 0,
    rating_avg        DECIMAL(2,1) NOT NULL DEFAULT 0.0,
    review_count      INT UNSIGNED NOT NULL DEFAULT 0,
    is_active         TINYINT(1)   NOT NULL DEFAULT 1,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_products_slug (slug),
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    KEY idx_products_category_id (category_id),
    KEY idx_products_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_images (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    url         VARCHAR(500) NOT NULL,
    alt_text    VARCHAR(200) NULL,
    sort_order  INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    KEY idx_product_images_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_specs (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    label       VARCHAR(100) NOT NULL,
    value       VARCHAR(200) NOT NULL,
    sort_order  INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_product_specs_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    KEY idx_product_specs_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_highlights (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    highlight   VARCHAR(255) NOT NULL,
    sort_order  INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_product_highlights_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    KEY idx_product_highlights_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS product_highlights;
DROP TABLE IF EXISTS product_specs;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
```

Mapping ke frontend `Product` type (untuk referensi saat menulis DTO di
[03-api-contract.md](03-api-contract.md)):

| Field frontend (`product.types.ts`) | Sumber di database |
|---|---|
| `id`, `name`, `price`, `compareAtPrice`, `currency`, `slug`, `maker`, `badge` | Kolom langsung di `products` |
| `category` | Nama diambil via JOIN `categories.name` (frontend pakai string, bukan object) |
| `rating`, `reviewCount` | `products.rating_avg`, `products.review_count` (denormalisasi, di-update saat ada review — review module di luar scope fase ini, kolom disiapkan saja) |
| `description`, `highlights[]`, `specs[]` | `products.description`, agregasi `product_highlights`, `product_specs` |
| `galleryLabels[]` / gambar | `product_images` (frontend belum pakai gambar asli — field ini akan dipetakan ke `url` dari `product_images` saat integrasi) |
| `component` (blade/handle/accessory dari configurator) | **Tidak dibuat di fase ini** — itu domain `entities/configurator`, di luar scope User/Auth + Product |

## Migration 3 (opsional, seed data untuk development)

```sql
-- +migrate Up
INSERT INTO users (id, email, password_hash, full_name, role, is_active)
VALUES (UUID(), 'admin@pisaupedia.com', '<bcrypt-hash-generated-di-go>', 'Admin Pisau Pedia', 'admin', 1);

INSERT INTO categories (id, name, slug) VALUES
  (UUID(), 'Chef Knives', 'chef-knives'),
  (UUID(), 'Santoku', 'santoku'),
  (UUID(), 'Nakiri', 'nakiri');
-- Produk seed ditambahkan manual setelah migration jalan (butuh category_id hasil insert di atas)

-- +migrate Down
DELETE FROM categories WHERE slug IN ('chef-knives', 'santoku', 'nakiri');
DELETE FROM users WHERE email = 'admin@pisaupedia.com';
```

> Catatan: `password_hash` seed **tidak boleh** ditulis sebagai plaintext di
> file SQL. Generate hash bcrypt-nya lewat skrip Go kecil atau `htpasswd`
> setara, lalu tempel hasilnya. Detail proses seeding dibahas di
> [04-module-user-auth.md](04-module-user-auth.md) langkah terakhir.

## Index yang sengaja dipasang untuk performa query yang sering dipakai frontend

- `products.slug` (unique) — halaman detail produk `/products/[slug]`
- `products.category_id` + `products.is_active` — listing/filter kategori
- `categories.slug` (unique) — halaman koleksi `/collections/[handle]`

Lanjut ke [03-api-contract.md](03-api-contract.md) untuk kontrak endpoint
yang akan dibangun di atas skema ini.
