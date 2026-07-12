# 08 — Roadmap & Checklist Gabungan

Ringkasan semua fase yang dibahas dokumen 00-07, plus daftar eksplisit apa
yang **belum** di-scope supaya tidak ada asumsi keliru soal cakupan
dokumentasi ini.

## Fase 0 — Inisiasi Project

- [ ] Struktur folder dibuat ([00](00-project-initiation.md))
- [ ] `go.mod` + `.env.example` + `.gitignore` siap
- [ ] MySQL bisa menyala via Docker Compose dan bisa dikoneksi
- [ ] Repo di-push ke Git remote

## Fase 1 — User & Authentication

- [ ] Migration `000001_create_users_and_auth` ([02](02-database-design.md))
- [ ] Entity, repository interface, usecase, infrastructure MySQL, DTO,
      middleware, handler, routing ([04](04-module-user-auth.md))
- [ ] Prometheus metrics middleware terpasang ([06](06-observability-prometheus-grafana.md))
- [ ] 7 skenario testing manual modul auth lulus
- [ ] Unit test usecase auth lulus

## Fase 2 — Product

- [ ] Migration `000002_create_products_and_categories` ([02](02-database-design.md))
- [ ] Entity, repository interface, usecase, infrastructure MySQL, DTO,
      handler, routing ([05](05-module-product.md))
- [ ] Seed data produk untuk development
- [ ] 8 skenario testing manual modul product lulus
- [ ] Response API dicocokkan manual dengan `Product` type frontend

## Fase 3 — Observability & Containerization (menyatukan Fase 1 & 2)

- [ ] `prometheus.yml` + service Prometheus jalan, target `UP` ([06](06-observability-prometheus-grafana.md))
- [ ] Grafana datasource + dashboard "API Overview" ter-provision otomatis
- [ ] `Dockerfile` multi-stage + `docker-compose.yml` lengkap ([07](07-docker-and-deployment.md))
- [ ] Seluruh stack (`app`, `mysql`, `migrate`, `prometheus`, `grafana`)
      naik bersamaan lewat `docker compose up -d --build`
- [ ] Verifikasi akhir: `/api/v1/categories`, Prometheus `/-/healthy`,
      Grafana `/api/health` semua merespons sukses

## Definition of Done fase ini (Fase 0-3)

Backend dianggap selesai untuk iterasi dokumentasi ini kalau:

1. Frontend `pisau-pedia` **bisa** memanggil `GET /api/v1/products`,
   `GET /api/v1/products/:slug`, `GET /api/v1/categories` dan mendapat data
   nyata dari MySQL (menggantikan `entities/product/model/product.data.ts`
   yang saat ini statis).
2. User bisa register, login, refresh token, kelola profil & alamat lewat
   API — walaupun frontend belum punya halaman login/register (ikon akun
   masih dekoratif sesuai catatan di README frontend), **API-nya sudah
   siap dipakai** saat halaman tersebut dibangun.
3. Traffic ke API kelihatan real-time di Grafana dashboard.

## Eksplisit di luar scope dokumentasi ini (jangan diasumsikan sudah termasuk)

Modul-modul berikut **ada** di frontend (`src/app/(store)/cart`,
`checkout`, `entities/review`, `entities/configurator`,
`admin/coupons`, `admin/orders`, `admin/engravings`, `admin/newsletter`)
tapi **sengaja tidak dibahas** di dokumentasi fase ini karena instruksi
awal membatasi ke User/Auth + Product saja:

- Cart & checkout (termasuk integrasi payment gateway apa pun)
- Order & order management
- Review produk (kolom `rating_avg`/`review_count` sudah disiapkan di
  skema, tapi **tidak** ada tabel `reviews` atau endpoint review di fase
  ini)
- Knife configurator (blade/handle/accessory)
- Coupon/diskon
- Newsletter subscriber
- Custom engraving
- Redis / caching layer
- Refresh token via cookie (fase ini pakai token di body/header, bukan
  `httpOnly` cookie)
- Rate limiting endpoint publik
- Email verification flow (kolom `email_verified_at` disiapkan, alurnya
  belum)

Modul-modul ini bisa didokumentasikan dengan pola yang sama (entity →
repository → usecase → infrastructure → delivery) sebagai dokumen
`09-module-*.md`, `10-module-*.md`, dst. di iterasi berikutnya, mengikuti
skema database yang **sudah disiapkan pondasinya** di fase ini (mis. FK ke
`users` dan `products` sudah ada, jadi modul order/review tinggal
nambah tabel yang mereferensikan keduanya).

> **Update:** Order & Review sudah didokumentasikan di
> [11-orders-reports-excel-pdf.md](11-orders-reports-excel-pdf.md).
> Newsletter subscriber dan sebagian perbaikan admin panel Customer
> sudah didokumentasikan di
> [12-perbaikan-newsletter-customer-admin.md](12-perbaikan-newsletter-customer-admin.md)
> — Coupon, Configurator, dan Custom engraving masih belum ada dokumen
> tersendiri. Integrasi payment gateway (RajaOngkir ongkir + Komerce
> Payment VA/QRIS) sudah didokumentasikan di
> [14-integrasi-rajaongkir-payment-qris.md](14-integrasi-rajaongkir-payment-qris.md).
> Perbaikan keandalan pembayaran (idempotency key, race condition webhook,
> cek status manual) ada di
> [15-keandalan-pembayaran-idempotency.md](15-keandalan-pembayaran-idempotency.md).
