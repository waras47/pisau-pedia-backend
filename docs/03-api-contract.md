# 03 — Kontrak API (REST)

Kontrak ini disepakati **sebelum** implementasi handler, supaya frontend
`pisau-pedia` bisa mulai integrasi (atau setidaknya menulis mock berbasis
kontrak ini) tanpa menunggu backend selesai 100%.

Base path: `/api/v1`. Semua response memakai format standar yang sama
seperti `kissaki-backend` (konsisten, tidak perlu didesain ulang):

**Sukses:**
```json
{ "success": true, "message": "Operation successful", "data": { } }
```

**Sukses dengan pagination:**
```json
{
  "success": true,
  "data": [ ],
  "meta": { "page": 1, "per_page": 12, "total": 48, "total_pages": 4 }
}
```

**Error:**
```json
{
  "success": false,
  "message": "Error description",
  "errors": { "email": "email is required" }
}
```

---

## Modul Auth

### `POST /api/v1/auth/register`

Request:
```json
{
  "email": "user@example.com",
  "password": "minimal8karakter",
  "full_name": "Nama Lengkap",
  "phone": "081234567890"
}
```

Response `201`:
```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "user": { "id": "uuid", "email": "user@example.com", "full_name": "...", "role": "customer" },
    "access_token": "jwt...",
    "refresh_token": "opaque-token...",
    "expires_in": 3600
  }
}
```

Error yang mungkin: `409` email sudah terdaftar, `422` validasi gagal.

### `POST /api/v1/auth/login`

Request: `{ "email": "...", "password": "..." }`
Response `200`: sama shape dengan register (`user` + `access_token` +
`refresh_token` + `expires_in`).
Error: `401` kredensial salah, `403` akun `is_active = 0`.

### `POST /api/v1/auth/refresh`

Request: `{ "refresh_token": "..." }`
Response `200`: `{ "access_token": "...", "refresh_token": "...", "expires_in": 3600 }`
(refresh token lama di-revoke, yang baru dikembalikan — token rotation).
Error: `401` token invalid/expired/revoked.

### `POST /api/v1/auth/logout` *(butuh access token)*

Request: `{ "refresh_token": "..." }` → revoke refresh token tsb di database.
Response `200`: `{ "success": true, "message": "Logged out" }`

---

## Modul User (butuh access token, kecuali disebutkan lain)

### `GET /api/v1/users/me`

Response `200`: profil user yang sedang login (tanpa `password_hash`).

### `PATCH /api/v1/users/me`

Request (semua field opsional): `{ "full_name": "...", "phone": "...", "avatar_url": "..." }`
Response `200`: profil yang sudah diupdate.

### `POST /api/v1/users/me/change-password`

Request: `{ "current_password": "...", "new_password": "..." }`
Response `200`: sukses. Error `401` kalau `current_password` salah.

### `GET /api/v1/users/me/addresses`

Response `200`: array `Address` milik user.

### `POST /api/v1/users/me/addresses`

Request: `{ "label", "full_name", "phone", "address_line", "city", "province", "postal_code", "is_default" }`
Response `201`: address yang dibuat. Kalau `is_default: true`, address lain
milik user otomatis di-set `is_default = false` (logic di usecase).

### `PATCH /api/v1/users/me/addresses/:id` & `DELETE /api/v1/users/me/addresses/:id`

Standar update/delete, scoped ke `user_id` milik token — user A tidak boleh
edit/hapus address milik user B (dicek di usecase, bukan hanya di query).

---

## Modul Product (publik, tidak butuh token — kecuali endpoint `/admin/*`)

### `GET /api/v1/categories`

Response `200`: array kategori (`id`, `name`, `slug`, `image_url`).

### `GET /api/v1/products`

Query params: `page`, `per_page` (default 12), `category` (slug),
`search` (cari di `name`), `sort` (`price_asc`|`price_desc`|`newest`).

Response `200`: array `Product` (shape mengikuti tabel mapping di
[02-database-design.md](02-database-design.md)) + `meta` pagination.

### `GET /api/v1/products/:slug`

Response `200`: detail produk lengkap (termasuk `specs[]`, `highlights[]`,
`images[]`). `404` kalau slug tidak ada atau `is_active = 0`.

---

## Modul Product — Admin (butuh access token + `role = admin`)

### `POST /api/v1/admin/products`

Request: seluruh field produk + array `specs`, `highlights`, `image_urls`.
Response `201`: produk yang dibuat.

### `PATCH /api/v1/admin/products/:id`

Update sebagian field. Response `200`: produk terupdate.

### `DELETE /api/v1/admin/products/:id`

Hard delete (konsisten dengan keputusan di `kissaki-backend`). Response
`200`: `{ "success": true, "message": "Product deleted" }`.

### `POST /api/v1/admin/categories`, `PATCH .../:id`, `DELETE .../:id`

CRUD kategori standar, sama pola dengan produk.

---

## Header & konvensi umum

- `Authorization: Bearer <access_token>` untuk semua endpoint yang butuh
  login.
- Semua request/response body: `application/json`.
- Endpoint admin memvalidasi role di **middleware**
  (`RequireRole("admin")`), bukan di masing-masing handler, supaya tidak ada
  endpoint admin yang lupa dicek.
- Semua endpoint (termasuk yang gagal) dicatat oleh Prometheus middleware —
  detail di [06-observability-prometheus-grafana.md](06-observability-prometheus-grafana.md).

Lanjut ke [04-module-user-auth.md](04-module-user-auth.md) untuk urutan
implementasi.
