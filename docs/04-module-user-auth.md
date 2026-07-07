# 04 — Modul User & Authentication: Urutan Implementasi

Urutan di bawah **wajib diikuti sesuai nomor** — tiap langkah bergantung pada
langkah sebelumnya karena mengikuti arah dependency Clean Architecture
(dari dalam ke luar: entity → repository interface → usecase →
infrastructure → delivery).

## Langkah 1 — Migration database

Jalankan migration `000001_create_users_and_auth` dari
[02-database-design.md](02-database-design.md). Verifikasi 3 tabel
(`users`, `addresses`, `refresh_tokens`) terbentuk sebelum lanjut.

## Langkah 2 — Entity (`internal/entity/`)

Buat `user.go`, `address.go`, `refresh_token.go`. Struct murni, method kecil
kalau perlu (mis. `User.IsAdmin() bool`). **Tidak ada** import selain
`time`, `github.com/google/uuid` — tidak boleh import `sqlx`, `echo`, dsb di
sini.

Field mengikuti kolom migration 1:1 (lihat
[02-database-design.md](02-database-design.md)). `PasswordHash` diberi tag
`json:"-"` supaya tidak pernah ter-serialize ke response walaupun struct
`entity.User` sempat lolos ke layer delivery.

## Langkah 3 — Repository interface (`internal/repository/`)

Buat `user_repository.go` dengan interface, contoh method yang dibutuhkan
usecase di langkah berikutnya:

- `Create(ctx, *entity.User) error`
- `FindByEmail(ctx, email string) (*entity.User, error)`
- `FindByID(ctx, id string) (*entity.User, error)`
- `Update(ctx, *entity.User) error`

Buat juga `address_repository.go` (`Create`, `FindByUserID`, `FindByID`,
`Update`, `Delete`, `UnsetDefaultForUser`) dan `refresh_token_repository.go`
(`Create`, `FindByTokenHash`, `Revoke`, `RevokeAllForUser`).

**Kenapa dipisah dari usecase**: supaya usecase di langkah 4 bisa ditulis
dan di-unit-test terhadap interface ini sebelum implementasi MySQL-nya ada.

## Langkah 4 — Usecase (`internal/usecase/auth_usecase.go`, `user_usecase.go`)

Ini jantung modul. Tulis usecase berikut, masing-masing sebagai method pada
struct yang menerima repository interface via constructor (dependency
injection manual):

1. **`Register(ctx, email, password, fullName, phone) (*entity.User, tokens, error)`**
   - Normalisasi email ke lowercase.
   - Cek `FindByEmail` — kalau sudah ada, return error "email already
     registered".
   - Hash password dengan `bcrypt.GenerateFromPassword` (cost default 10-12).
   - Generate `uuid.New()` untuk ID user, `Create` ke repository.
   - Panggil helper `generateTokenPair` (lihat poin 5) untuk access +
     refresh token.
2. **`Login(ctx, email, password) (*entity.User, tokens, error)`**
   - `FindByEmail`, kalau tidak ada → error generik "invalid credentials"
     (jangan bocorkan apakah email terdaftar atau tidak).
   - `bcrypt.CompareHashAndPassword`, gagal → error generik yang sama.
   - Cek `IsActive`, kalau `false` → error "account disabled".
   - Generate token pair.
3. **`RefreshToken(ctx, refreshTokenRaw string) (tokens, error)`**
   - Hash token yang diterima (SHA-256), cari via
     `refreshTokenRepo.FindByTokenHash`.
   - Validasi: ada, belum revoked, belum expired.
   - Revoke token lama, generate token pair baru (**rotation** — mencegah
     replay kalau refresh token lama bocor).
4. **`Logout(ctx, refreshTokenRaw string) error`** — hash token, revoke.
5. **Helper `generateTokenPair(user)`**:
   - Access token: JWT (`golang-jwt/jwt/v5`), claim `sub` (user id),
     `role`, `exp` (`JWT_ACCESS_EXPIRY_MINUTES` dari config), signing
     method HS256 dengan `JWT_SECRET`.
   - Refresh token: random opaque string (`crypto/rand`, 32 byte,
     base64url-encoded) — **bukan** JWT, supaya bisa di-revoke langsung
     lewat lookup database tanpa perlu blocklist.
   - Simpan **hash** refresh token (SHA-256) ke `refresh_tokens` table via
     repository, dengan `expires_at` dari `JWT_REFRESH_EXPIRY_HOURS`.
   - Return token mentah (bukan hash) ke caller — hash hanya untuk
     penyimpanan.

Lanjut `user_usecase.go` untuk profil & address:

6. **`GetProfile(ctx, userID)`**, **`UpdateProfile(ctx, userID, patch)`**,
   **`ChangePassword(ctx, userID, currentPassword, newPassword)`** (verifikasi
   `currentPassword` dulu sebelum hash & simpan yang baru).
7. **`ListAddresses`, `CreateAddress`, `UpdateAddress`, `DeleteAddress`** —
   setiap operasi **wajib** cek `address.UserID == userID` yang sedang
   login sebelum update/delete (jangan hanya mengandalkan `WHERE id = ?`
   di query — cek eksplisit di usecase supaya intent keamanannya jelas
   dibaca ulang nanti). Kalau `CreateAddress`/`UpdateAddress` mengirim
   `is_default: true`, panggil `UnsetDefaultForUser` dulu sebelum insert.

## Langkah 5 — Infrastructure MySQL (`internal/infrastructure/mysql/`)

Implementasikan interface dari Langkah 3: `user_repo.go`,
`address_repo.go`, `refresh_token_repo.go`. Pakai `sqlx` — `NamedExec` untuk
insert/update, `Get`/`Select` untuk read. Setiap error `sql.ErrNoRows`
diterjemahkan ke error domain (mis. `ErrUserNotFound`) di layer ini, supaya
usecase tidak pernah bergantung pada tipe error spesifik `database/sql`.

## Langkah 6 — DTO (`internal/delivery/http/dto/`)

Request struct dengan tag `validate` (`go-playground/validator`):

```go
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    FullName string `json:"full_name" validate:"required,min=2"`
    Phone    string `json:"phone" validate:"omitempty,min=8"`
}
```

Response struct: `UserResponse` (tanpa password), `AuthResponse`
(`user` + `access_token` + `refresh_token` + `expires_in`), `AddressResponse`.
Mapping `entity.User → dto.UserResponse` sebagai fungsi kecil di file yang
sama, dipanggil dari handler.

## Langkah 7 — Middleware (`internal/delivery/http/middleware/`)

1. **`auth.go`** — `JWTAuth()`: parse header `Authorization: Bearer`,
   verifikasi signature + `exp`, taruh `user_id` dan `role` ke Echo context
   (`c.Set("user_id", ...)`). Kalau invalid/absent → `401`.
2. **`RequireRole(role string)`** — dipasang setelah `JWTAuth()` di route
   group admin, cek `c.Get("role")`.
3. **`cors.go`** — allow origin dari `FRONTEND_URL` (config), credentials
   true kalau nanti butuh cookie (untuk sekarang token via header, jadi
   cukup allow origin + header `Authorization`).
4. **`logger.go`** — log tiap request (method, path, status, latency)
   pakai `zerolog`.

Middleware metrics (`metrics.go`) dibuat di modul ini juga karena harus aktif
sejak endpoint pertama ada — detail lengkap di
[06-observability-prometheus-grafana.md](06-observability-prometheus-grafana.md).

## Langkah 8 — Handler (`internal/delivery/http/handler/auth_handler.go`, `user_handler.go`)

Handler **tipis** — hanya: bind request → validate → panggil usecase →
mapping ke DTO response → `pkg/response` helper untuk JSON. Tidak ada
business logic di handler.

## Langkah 9 — Routing & wiring (`cmd/api/main.go`)

Urutan wiring per fitur (bottom-up, sesuai arah dependency):

```
mysqlUserRepo := mysql.NewUserRepository(db)
mysqlAddressRepo := mysql.NewAddressRepository(db)
mysqlRefreshTokenRepo := mysql.NewRefreshTokenRepository(db)

authUsecase := usecase.NewAuthUsecase(mysqlUserRepo, mysqlRefreshTokenRepo, cfg.JWT)
userUsecase := usecase.NewUserUsecase(mysqlUserRepo, mysqlAddressRepo)

authHandler := handler.NewAuthHandler(authUsecase)
userHandler := handler.NewUserHandler(userUsecase)

v1 := e.Group("/api/v1")
v1.POST("/auth/register", authHandler.Register)
v1.POST("/auth/login", authHandler.Login)
v1.POST("/auth/refresh", authHandler.Refresh)
v1.POST("/auth/logout", authHandler.Logout, middleware.JWTAuth())

users := v1.Group("/users/me", middleware.JWTAuth())
users.GET("", userHandler.GetProfile)
users.PATCH("", userHandler.UpdateProfile)
users.POST("/change-password", userHandler.ChangePassword)
users.GET("/addresses", userHandler.ListAddresses)
users.POST("/addresses", userHandler.CreateAddress)
users.PATCH("/addresses/:id", userHandler.UpdateAddress)
users.DELETE("/addresses/:id", userHandler.DeleteAddress)
```

## Langkah 10 — Testing manual

Urutan verifikasi manual pakai `curl` (atau Postman/Insomnia):

1. `POST /auth/register` → dapat `access_token` + `refresh_token`.
2. `GET /users/me` dengan `Authorization: Bearer <access_token>` → profil
   sesuai.
3. `GET /users/me` **tanpa** header → harus `401`.
4. `POST /auth/login` dengan password salah → harus `401`, pesan generik.
5. `POST /auth/refresh` dengan `refresh_token` dari langkah 1 → dapat token
   baru; ulangi request dengan `refresh_token` **lama** → harus gagal
   (sudah di-revoke oleh rotation).
6. `POST /users/me/addresses` (buat 2 address, salah satu `is_default:
   true`) → cek `is_default` yang lama otomatis `false`.
7. Buat user kedua, coba `PATCH /users/me/addresses/:id` pakai `:id` milik
   user pertama → harus `403`/`404` (bukan berhasil edit data orang lain).

## Checklist keluar dari modul ini

- [ ] Semua 10 langkah di atas selesai
- [ ] Unit test usecase (`auth_usecase_test.go`) pakai mock repository —
      minimal test `Register` (email duplikat), `Login` (password salah),
      `RefreshToken` (token expired)
- [ ] 7 skenario manual di Langkah 10 lulus semua
- [ ] Endpoint tercatat di Prometheus (`/metrics` menunjukkan hit count
      untuk `/auth/*` dan `/users/*`)

Setelah ini selesai, lanjut ke
[05-module-product.md](05-module-product.md).
