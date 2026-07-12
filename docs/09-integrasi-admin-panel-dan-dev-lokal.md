# 09 — Integrasi Admin Panel Frontend & Setup Dev Lokal

Dokumen ini merangkum hasil kerja sesi pengembangan yang menyambungkan
backend (dokumen [04](04-module-user-auth.md) dan
[05](05-module-product.md)) dengan admin panel di frontend
`pisau-pedia`, termasuk perbaikan bug yang ditemukan di jalan, setup
environment dev lokal (non-Docker untuk app, Docker untuk service
pendukung), dan rencana lanjutan yang belum dieksekusi. Kode lengkap
sengaja **tidak** disalin ulang di sini — cukup rujuk path file dan
status tiap bagian. Detail implementasi ada di riwayat percakapan sesi
ini kalau dibutuhkan.

## Lokasi kerja aktif

Project sempat dipindah strukturnya oleh developer di tengah sesi.
Lokasi yang **aktif dipakai** sekarang:

```
pisau-pedia-project/pisau-pedia-project/pisau-pedia-backend
pisau-pedia-project/pisau-pedia-project/pisau-pedia
```

Folder `pisau-pedia-project/pisau-pedia-backend` (satu level di atas,
tanpa nested) adalah sisa lokasi lama sebelum reorganisasi — **tidak
dipakai lagi**, sengaja belum dihapus otomatis.

## 1. Environment dev lokal — keputusan yang diambil

Berbeda dari asumsi awal di [07-docker-and-deployment.md](07-docker-and-deployment.md)
(semua service naik lewat `docker compose up`), untuk pengembangan
harian dipilih pendekatan campuran:

| Komponen | Cara jalan | Alasan |
|---|---|---|
| Backend Go (`cmd/api`) | `make run` langsung di host | Iterasi lebih cepat daripada rebuild image tiap ubah kode |
| MySQL | Instalasi native di host (Laragon/MySQL Server), bukan container | Sudah terpasang di mesin dev, tidak perlu duplikasi |
| MinIO | Container Docker standalone (bukan lewat `docker-compose.yml` backend) | Tidak ada instalasi native untuk object storage |
| Frontend Next.js | `npm run dev` langsung | Standar dev Next.js |

Konsekuensi konfigurasi:

- `.env` backend: `DB_HOST=localhost` (bukan `mysql` seperti default
  `.env.example`, yang mengasumsikan nama service Docker Compose).
- `.env` backend: `DB_USER=root`, `DB_PASSWORD` kosong — mengikuti
  akun root MySQL lokal di mesin dev ini (bukan user `pisaupedia`
  dedicated seperti skenario Docker Compose penuh).
- Database `pisau_pedia` dan migration (`000001`, `000002`) sudah
  dijalankan terhadap MySQL lokal ini menggunakan image
  `migrate/migrate:v4.17.1` sekali jalan (bukan `migrate` CLI native,
  karena belum ter-install di mesin).

## 2. Bug backend yang ditemukan & diperbaiki

Ditemukan saat proses testing register/login manual — gejalanya:
register sukses (user kebentuk), tapi response selalu
`"failed to register"` / `"failed to login"` karena proses generate
token gagal diam-diam.

**Root cause:** [internal/infrastructure/mysql/refresh_token_repo.go](../internal/infrastructure/mysql/refresh_token_repo.go)
melakukan query ke tabel `refresh_token` (tanpa akhiran "s"),
sedangkan nama tabel hasil migration adalah `refresh_tokens`. Semua
query (`Create`, `FindByTokenHash`, `Revoke`, `RevokeAllForUser`)
diperbaiki ke nama tabel yang benar.

Sekalian ditemukan & diperbaiki bug kedua di fungsi yang sama:
`RevokeAllForUser` memakai `WHERE id = ?` dengan parameter `userID`
(harusnya `WHERE user_id = ?`) — akibatnya revoke-semua-sesi-user
tidak pernah benar-benar mengenai baris yang dimaksud. Sudah
diperbaiki di file yang sama.

**Status:** ✅ Sudah diperbaiki di kode dan diverifikasi lewat
`curl` — login mengembalikan `access_token` + `refresh_token` dengan
benar setelah fix.

## 3. Router extraction

`main.go` sebelumnya mendefinisikan semua route Echo langsung di
`func main()`. Ini sudah dipindah ke package terpisah:
[internal/delivery/http/router/router.go](../internal/delivery/http/router/router.go),
dengan `router.Dependencies` sebagai struct berisi semua handler +
middleware, dan `router.Register(e, deps)` dipanggil dari `main.go`.
Tujuannya supaya `main.go` murni jadi entry point (wiring dependency +
start server), bukan tempat definisi routing.

**Status:** ✅ Selesai.

## 4. Akun admin untuk testing lokal

Karena endpoint `POST /auth/register` **selalu** membuat user dengan
role `customer` (ini disengaja — publik tidak boleh bisa daftar jadi
admin sendiri), akun admin untuk testing dibuat manual:

1. Register lewat API seperti biasa.
2. Role di-promote manual lewat SQL langsung ke tabel `users`
   (`UPDATE users SET role='admin' WHERE email=...`).

Kredensial testing yang sudah aktif di database lokal:
`admin@pisaupedia.test` / `Admin12345`.

**Rencana perbaikan (belum dikerjakan):** `Makefile` sudah punya
target `make seed` yang memanggil `go run ./cmd/seed`, tapi folder
`cmd/seed` **belum pernah dibuat**. Rencananya: seed script yang
baca `SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` / `SEED_ADMIN_NAME`
dari `.env` (sudah ditambahkan ke `.env`, lihat tabel env di bawah)
dan membuat/mempromosikan akun admin secara idempotent, supaya tidak
perlu SQL manual lagi. Juga perlu perbaikan kecil di
`userRepository.Update()` ([internal/infrastructure/mysql/user_repo.go](../internal/infrastructure/mysql/user_repo.go))
yang saat ini tidak menyertakan kolom `role` — jadi tidak bisa dipakai
untuk promote user existing sampai ditambahkan.

**Status:** ⏳ Direncanakan, belum diimplementasi.

## 5. Integrasi autentikasi di admin panel frontend

Sebelum sesi ini, `/admin` di frontend bisa diakses siapa saja tanpa
login (tidak ada guard sama sekali). Sekarang sudah ada alur login
penuh, mengikuti arsitektur FSD (`entities` → `features` → `widgets` →
`app`) yang sudah dipakai di codebase frontend:

| File | Peran |
|---|---|
| `src/shared/config/env.ts` | Baca `NEXT_PUBLIC_API_BASE_URL` |
| `src/shared/api/client.ts` + `http-error.ts` | Wrapper fetch: attach `Authorization: Bearer`, auto-refresh token sekali kalau kena 401 |
| `src/entities/session/model/session.types.ts` | Tipe `SessionUser`, `TokenPair` |
| `src/entities/session/model/token-storage.ts` | Simpan/ambil token + data user dari `localStorage` |
| `src/features/auth/api/auth.api.ts` | Panggil `/auth/login`, `/auth/logout` |
| `src/features/auth/model/AuthProvider.tsx` | React context status login (`loading`/`authenticated`/`unauthenticated`) |
| `src/features/auth/ui/LoginForm.tsx` | Form login |
| `src/app/admin/login/page.tsx` | Halaman login, di luar shell sidebar/header |
| `src/app/admin/layout.tsx` | `AdminGuard` — redirect ke login kalau belum auth, redirect keluar (`/`) kalau role bukan `admin` |
| `src/widgets/admin/admin-header/AdminHeader.tsx` | Nampilkan nama & role user asli dari sesi + dropdown berisi tombol Logout |

**Keputusan desain:** token disimpan di `localStorage`, bukan
`httpOnly` cookie — konsisten dengan catatan scope di
[08-roadmap-checklist.md](08-roadmap-checklist.md) ("Refresh token via
cookie... tidak" ada di fase ini) dan karena backend memang
mengembalikan token lewat JSON body, bukan `Set-Cookie`. CORS backend
([internal/delivery/http/middleware/cors.go](../internal/delivery/http/middleware/cors.go))
juga cuma mengizinkan satu origin (`FRONTEND_URL`) tanpa
`AllowCredentials`, jadi pendekatan cookie lintas-origin tidak akan
langsung jalan tanpa perubahan tambahan.

Sudah diverifikasi end-to-end lewat browser otomatis: guard redirect
ke `/admin/login` saat belum login, login sukses membawa masuk ke
dashboard, dropdown user menampilkan nama/role asli, tombol Logout
berfungsi (clear token + redirect balik ke halaman login).

**Status:** ✅ Selesai & terverifikasi.

## 6. CRUD Product di admin panel

Halaman `src/app/admin/products/page.tsx` sudah tersambung ke API
backend asli (sebelumnya cuma UI di atas array mock
`entities/product/model/product.data.ts`).

**Backend — `PATCH /admin/products/:id` sekarang bisa ubah
`images`/`specs`/`highlights`** (sebelumnya cuma kolom dasar):

| File | Perubahan |
|---|---|
| [internal/delivery/http/dto/product_dto.go](../internal/delivery/http/dto/product_dto.go) | `UpdateProductRequest` tambah field `Images`, `Specs`, `Highlights`; `ToInput()` di-mapping ke semuanya |
| [internal/usecase/product_usecase.go](../internal/usecase/product_usecase.go) | `UpdateProduct` replace `product.Images/Specs/Highlights` kalau field terkait dikirim (non-nil) |
| [internal/infrastructure/mysql/product_repo.go](../internal/infrastructure/mysql/product_repo.go) | `Update()` jadi transactional: update kolom dasar + delete-and-reinsert untuk 3 tabel anak, pola sama seperti `Create` |

**Frontend:**

| File | Peran |
|---|---|
| `src/entities/product/api/product.api.ts` | `listProducts`, `getProductBySlug`, `createProduct`, `updateProduct`, `deleteProduct` |
| `src/entities/category/api/category.api.ts` | `listCategories`, dipakai buat dropdown kategori di form |
| `src/app/admin/products/page.tsx` | Fetch data asli saat mount, `handleSave`/`handleDelete` manggil API, kategori jadi `<select>` dari data asli |

Sudah diverifikasi lewat browser: halaman render tanpa error, request
ke `/products` & `/categories` terkirim dengan benar (gagal karena
CORS origin beda saat dites di port preview 3010, bukan bug kode —
di `localhost:3000` seharusnya jalan normal).

**Yang masih jadi blocker buat testing end-to-end:**

- Tabel `categories` di database lokal masih **kosong** — perlu diisi
  minimal satu baris sebelum form Create Product bisa submit
  (`category_id` wajib UUID valid, dan dropdown kategori juga masih
  kosong tanpa data ini).
- Slug produk **sengaja tidak bisa diedit** lewat update (keputusan
  desain, supaya URL produk stabil) — kalau perlu, ini perubahan
  terpisah.

**Status:** ✅ Kode backend & frontend sudah diimplementasi dan
build/compile sukses; ⏳ belum ada testing end-to-end dengan data
sungguhan karena tabel `categories` masih kosong.

## 7. Upload gambar produk via MinIO

**Infra lokal sudah berjalan** (dijalankan manual lewat `docker run`,
bukan lewat `docker-compose.yml`):

- Container `pisaupedia-minio` (image `minio/minio`), network Docker
  `pisaupedia-net`, volume persisten `pisaupedia_minio_data`.
- API MinIO di `http://localhost:9000`, Console admin di
  `http://localhost:9001`.
- Bucket `pisaupedia` sudah dibuat dan diset akses baca publik
  (`mc anonymous set download`), supaya URL gambar bisa langsung
  diakses browser tanpa signed URL.

Variabel environment baru yang sudah ditambahkan ke `.env`:

| Variabel | Nilai dev lokal |
|---|---|
| `MINIO_ROOT_USER` | `pisaupedia_admin` |
| `MINIO_ROOT_PASSWORD` | `pisaupedia_minio_secret` |
| `MINIO_ENDPOINT` | `localhost:9000` |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | sama dengan root user/password |
| `MINIO_BUCKET` | `pisaupedia` |
| `MINIO_USE_SSL` | `false` |
| `MINIO_PUBLIC_URL` | `http://localhost:9000/pisaupedia` |

Backend integrasinya sudah ditulis:

| File | Peran |
|---|---|
| `pkg/config/config.go` | `MinioConfig` dibaca dari env `MINIO_*` |
| `pkg/storage/minio.go` | Wrapper client `minio-go/v7`, method `UploadImage` |
| `internal/delivery/http/handler/upload_handler.go` | Endpoint upload, validasi tipe (`jpeg`/`png`/`webp`) & ukuran (maks 5MB) |
| `internal/delivery/http/router/router.go` | Route `POST /api/v1/admin/uploads/image` (admin-only) |
| `cmd/api/main.go` | Wiring `storage.New` → `handler.NewUploadHandler` → `router.Dependencies` |

Yang **belum** dikerjakan (masih rencana):

- Frontend: `src/shared/api/upload.api.ts` (upload via `FormData`),
  komponen `ImagesEditor` di form Add/Edit Product (pola serupa
  `SpecsEditor`/`HighlightsEditor` yang sudah ada), field `images`
  ditambahkan ke tipe `Product`.
- Belum ada testing end-to-end upload file sungguhan (baru sebatas
  compile sukses + infra MinIO sehat).
- Entri service `minio` + `minio-init` di `docker-compose.yml` kalau
  nanti mau dipindah dari `docker run` manual ke Compose (relevan
  untuk deployment, bukan cuma dev lokal).

**Status:** ✅ Backend selesai ditulis & compile sukses, infra MinIO
jalan & sehat; ⏳ frontend (`ImagesEditor` + upload API) belum
dikerjakan, belum ada testing end-to-end.

## 8. Bug tambahan yang ditemukan & diperbaiki (saat wiring product update & upload)

Ditemukan saat `go build` gagal setelah kode Bagian 6 & 7 di atas
ditulis:

| File | Bug | Perbaikan |
|---|---|---|
| `internal/infrastructure/mysql/product_repo.go` | `Update()` pakai `r.db.BeginTx(...)` (method bawaan `*sql.DB`, hasilnya `*sql.Tx` — tidak punya `NamedExecContext`) | Ganti ke `r.db.BeginTxx(...)` (method `sqlx`, hasilnya `*sqlx.Tx`), sama seperti pola di `Create` |
| `internal/infrastructure/mysql/product_repo.go` | Error dari query `UPDATE products` utama di-`return nil` (ditelan), bukan `return err` — kalau update gagal, tetap dianggap sukses | Diperbaiki jadi `return err` |
| `internal/delivery/http/router/router.go` | Struct `Dependencies` dipakai field `UploadHandler` tapi belum didefinisikan di struct | Field `UploadHandler *handler.UploadHandler` ditambahkan |
| `cmd/api/main.go` | `uploadHandler` dibuat **setelah** `router.Register(...)` dipanggil, jadi tidak pernah dimasukkan ke `router.Dependencies{}` | Urutan dipindah: `storage.New` + `handler.NewUploadHandler` sekarang sebelum `router.Register`, dan `UploadHandler: uploadHandler` ditambahkan ke struct literal |
| `pkg/config/config.go` | `MinioConfig` sudah didefinisikan tapi tidak dipakai di `Config` struct maupun di-baca di `Load()` | Field `Minio MinioConfig` ditambahkan ke `Config`, dan `Load()` diisi baca `MINIO_*` dari env |
| `internal/delivery/http/dto/product_dto.go` | Struct tag `UpdateProductRequest.Specs` salah format: `` `json:"specs", validate:"omitempty,dive"` `` (ada koma nyasar antar tag, seharusnya dipisah spasi) — bikin parsing tag `validate` gagal diam-diam | Koma dihapus |
| `internal/delivery/http/dto/product_dto.go` | `UpdateProductRequest.ToInput()` tidak pernah memetakan `Images`/`Specs`/`Highlights` ke `usecase.ProductInput` — walau field-nya sudah ada di DTO dan usecase sudah siap menerimanya, request dari frontend tidak akan pernah sampai ke usecase | Mapping ditambahkan, sama seperti pola di `CreateProductRequest.ToInput()` |
| `src/app/admin/products/page.tsx` (frontend) | Fungsi `listProducts`, `listCategories`, `createProduct`, `updateProduct`, `deleteProduct`, `getProductBySlug`, tipe `CategoryApiItem`, dan `HttpError` dipakai di kode tapi tidak di-*import* — `ReferenceError: listProducts is not defined` saat runtime | Import ditambahkan dari `@/entities/product/api/product.api`, `@/entities/category/api/category.api`, dan `@/shared/api/http-error` |

Semua sudah diverifikasi: `go build ./...` sukses tanpa error, dan
halaman `/admin/products` di frontend sudah dicek lewat browser
otomatis — render normal tanpa `ReferenceError`.

## Ringkasan status

| Bagian | Status |
|---|---|
| Router extraction (`main.go` → `router.go`) | ✅ Selesai |
| Bug `refresh_token` table name + `RevokeAllForUser` | ✅ Diperbaiki |
| Env & migration dev lokal (MySQL native, host) | ✅ Jalan |
| Akun admin testing | ✅ Ada (manual SQL); seed script otomatis ⏳ direncanakan |
| Auth admin panel frontend (login/guard/logout) | ✅ Selesai & terverifikasi |
| CRUD Product tersambung ke API asli | ✅ Kode selesai; ⏳ testing end-to-end tertahan tabel `categories` kosong |
| Upload gambar via MinIO (backend) | ✅ Kode selesai, compile sukses |
| Upload gambar via MinIO (frontend `ImagesEditor`) | ⏳ Belum dikerjakan |
| Bug wiring (BeginTx, Dependencies, config Minio, DTO mapping, import) | ✅ Diperbaiki (lihat Bagian 8) |

**Langkah selanjutnya yang paling blocking:** isi tabel `categories`
(minimal 1 baris) supaya form Create/Edit Product bisa benar-benar
dites end-to-end, lalu lanjut ke `ImagesEditor` di frontend supaya
upload gambar juga bisa dites.
