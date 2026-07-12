# Pisau Pedia — Backend API (Dokumentasi Perencanaan)

> **Status repo ini: dokumentasi saja.** Belum ada kode. Isi folder `docs/`
> adalah rencana pengembangan step-by-step untuk backend **Pisau Pedia**,
> mulai dari inisiasi project sampai implementasi modul **User & Authentication**
> serta **Product**. Modul lain (cart, order, payment, review, dsb) sengaja
> **belum** dibahas — lihat [docs/08-roadmap-checklist.md](docs/08-roadmap-checklist.md).

## Kenapa backend ini dibuat

Frontend [`pisau-pedia`](../pisau-pedia) (Next.js App Router, Feature-Sliced
Design) saat ini masih 100% data statis (`entities/*/model/*.data.ts`) dan
tidak punya autentikasi nyata (ikon akun masih dekoratif). Backend ini dibuat
supaya frontend punya API sungguhan untuk data produk & akun user.

Arsitektur backend **mengikuti pola project referensi `kissaki-backend`**
(Clean Architecture 4 layer, Go + Echo + JWT), dengan 2 perbedaan sengaja:

| Aspek | `kissaki-backend` (referensi) | `pisau-pedia-backend` (project ini) |
|---|---|---|
| Database | PostgreSQL | **MySQL 8** |
| Cache/session | Redis | Tidak dipakai dulu (refresh token disimpan di tabel MySQL) |
| Observability | Tidak ada | **Prometheus + Grafana** |
| Payment gateway | Xendit | Belum dibahas (di luar scope fase ini) |

## Tech Stack

| Layer | Pilihan |
|---|---|
| Bahasa | Go 1.22+ |
| HTTP framework | Echo v4 |
| Database | MySQL 8 |
| SQL driver/helper | `sqlx` + `go-sql-driver/mysql` |
| Migration | `golang-migrate` |
| Auth | JWT (access + refresh token), `bcrypt` untuk hashing password |
| Observability | Prometheus (metrics) + Grafana (dashboard) |
| Container | Docker + Docker Compose |
| Reverse proxy (opsional, production) | Nginx |

## Scope fase ini

Hanya 2 modul domain yang didokumentasikan sampai tuntas di iterasi ini:

1. **User & Authentication** — register, login, refresh token, profil,
   alamat pengiriman, role `customer`/`admin`.
2. **Product** — kategori, produk, gambar produk, spesifikasi, highlight
   (mengikuti bentuk data `Product` yang sudah dipakai frontend, lihat
   [`src/entities/product/model/product.types.ts`](../pisau-pedia/src/entities/product/model/product.types.ts)).

## Cara membaca dokumentasi ini

Baca berurutan sesuai nomor file di `docs/`:

| # | File | Isi |
|---|---|---|
| 00 | [project-initiation.md](docs/00-project-initiation.md) | Inisiasi repo, tooling, skeleton Docker Compose |
| 01 | [architecture-and-tech-stack.md](docs/01-architecture-and-tech-stack.md) | Clean Architecture, struktur folder |
| 02 | [database-design.md](docs/02-database-design.md) | ERD + DDL MySQL |
| 03 | [api-contract.md](docs/03-api-contract.md) | Kontrak endpoint REST (request/response) |
| 04 | [module-user-auth.md](docs/04-module-user-auth.md) | Urutan implementasi modul User & Auth |
| 05 | [module-product.md](docs/05-module-product.md) | Urutan implementasi modul Product |
| 06 | [observability-prometheus-grafana.md](docs/06-observability-prometheus-grafana.md) | Instrumentasi metrics & dashboard |
| 07 | [docker-and-deployment.md](docs/07-docker-and-deployment.md) | Dockerfile, docker-compose, Makefile |
| 08 | [roadmap-checklist.md](docs/08-roadmap-checklist.md) | Checklist bertahap + apa yang belum di-scope |
| 09 | [integrasi-admin-panel-dan-dev-lokal.md](docs/09-integrasi-admin-panel-dan-dev-lokal.md) | Sambungkan admin panel frontend ke backend, setup dev lokal, bug yang ditemukan |
| 10 | [i18n-dan-currency-storefront.md](docs/10-i18n-dan-currency-storefront.md) | i18n EN/ID & currency storefront |
| 11 | [orders-reports-excel-pdf.md](docs/11-orders-reports-excel-pdf.md) | Modul Order (checkout, admin) + Sales/Inventory Report (export Excel/PDF) |
| 12 | [perbaikan-newsletter-customer-admin.md](docs/12-perbaikan-newsletter-customer-admin.md) | Perbaikan admin panel Newsletter (pagination/filter) & Customer (modal detail), fix N+1 query & bug kolom `role` |
| 13 | [notifikasi-aktivitas-admin.md](docs/13-notifikasi-aktivitas-admin.md) | Sistem notifikasi admin (polling) untuk modul Product, Service, Order, Customer |

## Prinsip yang dipegang

- **Dependency Rule Clean Architecture**: `delivery` → `usecase` →
  `repository`/`entity` ← `infrastructure`. Layer dalam tidak pernah
  bergantung ke layer luar.
- **Migration-first**: skema database didesain dan direview di
  `02-database-design.md` sebelum satu baris kode Go pun ditulis.
- **Contract-first**: bentuk request/response API disepakati di
  `03-api-contract.md` sebelum handler diimplementasikan, supaya frontend
  bisa mulai integrasi tanpa menunggu backend selesai penuh.
- **Observability sejak awal**, bukan ditempel belakangan — middleware
  metrics masuk di fase yang sama dengan modul pertama (`user-auth`).
