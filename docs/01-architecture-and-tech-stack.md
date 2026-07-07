# 01 — Arsitektur & Tech Stack

## Clean Architecture — 4 layer

Sama seperti `kissaki-backend`, dependency hanya boleh mengarah **ke dalam**:

```
┌─────────────────────────────────────────────┐
│  DELIVERY (HTTP handler, middleware, DTO)    │  ← tahu tentang HTTP/Echo
├─────────────────────────────────────────────┤
│  USECASE (business logic)                    │  ← tahu tentang entity + repo interface
├─────────────────────────────────────────────┤
│  ENTITY + REPOSITORY interface (domain)       │  ← tidak tahu apa-apa soal luar
├─────────────────────────────────────────────┤
│  INFRASTRUCTURE (MySQL, JWT signing, dll)      │  ← implementasi kontrak repository
└─────────────────────────────────────────────┘
```

Aturan:

- `entity` tidak boleh mengimpor package lain di `internal/` (murni struct +
  method domain kecil, tanpa dependency ke Echo/sqlx/dsb).
- `repository` hanya berisi **interface**, bukan implementasi. Contoh:
  `UserRepository` interface didefinisikan di `internal/repository/`, lalu
  diimplementasikan oleh `internal/infrastructure/mysql/user_repo.go`.
- `usecase` hanya bergantung pada interface `repository`, tidak pernah
  mengimpor `infrastructure/mysql` langsung. Ini yang membuat usecase bisa
  di-unit-test dengan mock repository tanpa database sungguhan.
- `delivery/http/handler` hanya memanggil `usecase`, tidak pernah memanggil
  `repository` atau `infrastructure` langsung.
- Wiring (menyambungkan implementasi konkret ke interface) terjadi satu kali
  di `cmd/api/main.go`.

## Kenapa MySQL, bukan PostgreSQL (beda dari referensi)

Keputusan ini datang dari kebutuhan project (stack yang diminta), bukan dari
keterbatasan Postgres. Konsekuensi teknis yang perlu diperhatikan saat
menulis migration & repository:

| Hal | Perilaku Postgres (referensi) | Perilaku MySQL (project ini) |
|---|---|---|
| UUID generation | `gen_random_uuid()` bawaan | Tidak ada fungsi native — UUID **digenerate di sisi Go** (`uuid.New()`) sebelum `INSERT`, disimpan sebagai `CHAR(36)` |
| `RETURNING` clause | Ada, bisa `INSERT ... RETURNING id` | **Tidak ada** — harus `SELECT` ulang setelah `INSERT`, atau generate ID di Go dulu lalu pakai ID itu langsung (pendekatan yang dipakai di sini) |
| Partial index | Didukung (`WHERE is_active`) | Tidak didukung — gunakan index biasa + filter di query |
| Auto-update `updated_at` | Trigger PL/pgSQL | `ON UPDATE CURRENT_TIMESTAMP` langsung di kolom (lebih sederhana) |
| Case sensitivity kolom string unik (email) | Case-sensitive by default | Tergantung collation; gunakan `utf8mb4_unicode_ci` (default case-insensitive) — **penting**: putuskan sejak awal apakah email disimpan lowercase-normalized di usecase sebelum query, supaya perilaku unik konsisten |

Keputusan yang diambil: **generate UUID di layer usecase/entity (Go),
bukan di database**, dan **normalisasi email ke lowercase sebelum
disimpan/dicari**. Ini dicatat di sini karena mempengaruhi cara
`user_repository.go` ditulis nanti.

## Struktur folder (ringkasan, detail di 00-project-initiation.md)

```
internal/
├── entity/            user.go, address.go, product.go, category.go, ...
├── repository/         user_repository.go, product_repository.go, ...
├── usecase/            auth_usecase.go, user_usecase.go, product_usecase.go, ...
├── delivery/http/
│   ├── handler/        auth_handler.go, user_handler.go, product_handler.go, ...
│   ├── middleware/      auth.go, cors.go, logger.go, metrics.go
│   └── dto/             request.go, response.go
└── infrastructure/
    └── mysql/            user_repo.go, product_repo.go, ...
```

## Tech Stack (lengkap dengan alasan pilihan)

| Komponen | Pilihan | Alasan |
|---|---|---|
| Bahasa | Go 1.22+ | Konsisten dengan project referensi, performa baik untuk REST API |
| HTTP framework | Echo v4 | Middleware chain sederhana, dipakai di referensi, dokumentasi bagus |
| Database | MySQL 8 | Sesuai kebutuhan stack yang diminta |
| DB driver | `go-sql-driver/mysql` + `sqlx` | `sqlx` memberi `StructScan` tanpa perlu full ORM |
| Migration | `golang-migrate` | File SQL murni, mudah diaudit, sama seperti referensi |
| Auth | JWT (access + refresh) | Stateless, tidak butuh Redis untuk session |
| Password hashing | `bcrypt` | Standar industri untuk hashing password |
| Validasi input | `go-playground/validator` | Validasi via struct tag di DTO |
| Logger | `zerolog` | Structured logging (JSON), murah secara performa |
| Config | `viper` | Load `.env` + override via environment variable saat production |
| Metrics | `prometheus/client_golang` | Expose `/metrics`, standar de facto untuk Prometheus scraping |
| Dashboard | Grafana | Visualisasi metrics dari Prometheus |
| Container | Docker + Docker Compose | Konsistensi environment dev/staging |

## Kenapa tidak pakai Redis (beda dari referensi)

Stack yang diminta hanya menyebut MySQL, Docker, Grafana, Prometheus — tidak
ada Redis. Konsekuensinya:

- **Refresh token** tidak disimpan di Redis, melainkan di tabel MySQL
  `refresh_tokens` (lihat [02-database-design.md](02-database-design.md)),
  supaya tetap bisa di-revoke (logout / rotate token) tanpa cache layer.
- **Tidak ada cache produk** di fase ini. Kalau nanti traffic jadi masalah,
  ini bisa ditambahkan sebagai peningkatan terpisah — tidak diblokir oleh
  keputusan arsitektur sekarang karena `usecase` hanya bergantung pada
  interface `repository`, jadi menambah cache-decorator di kemudian hari
  tidak mengubah kontrak.

## Alur request tipikal (contoh: GET /api/v1/products)

```
Client → Echo router → middleware (CORS, logger, metrics)
       → ProductHandler.List()
       → ProductUsecase.List(filter)
       → ProductRepository.FindAll(filter)   [interface]
       → mysql.productRepo.FindAll(filter)   [implementasi]
       → MySQL
       ← rows
       ← []entity.Product
       ← []dto.ProductResponse (mapping di usecase atau handler, konsisten pilih salah satu — usecase)
       ← JSON response (pkg/response)
```

Lanjut ke [02-database-design.md](02-database-design.md) untuk skema
database.
