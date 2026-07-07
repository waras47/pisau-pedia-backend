# 00 — Inisiasi Project

Tujuan dokumen ini: langkah konkret dari "folder kosong" sampai skeleton
project siap diisi kode modul User/Auth dan Product. Tidak ada kode bisnis
di tahap ini — hanya scaffolding, tooling, dan konfigurasi.

## Prasyarat di mesin developer

- [Go 1.22+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) 24+
  (menyediakan Docker Engine + Docker Compose v2)
- [Make](https://gnuwin32.sourceforge.net/packages/make.htm) (opsional tapi
  disarankan — dipakai di semua contoh perintah dokumentasi lain)
- `golang-migrate` CLI (opsional untuk lokal; kalau tidak diinstal, migration
  tetap bisa dijalankan lewat service `migrate` di Docker Compose)
- Git

## Langkah 1 — Buat struktur folder repo

```
pisau-pedia-backend/
├── cmd/
│   └── api/                  # entrypoint (main.go)
├── internal/
│   ├── entity/                # struct domain (User, Product, ...)
│   ├── repository/            # interface repository (kontrak)
│   ├── usecase/                # business logic
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/       # HTTP handler (Echo)
│   │       ├── middleware/    # auth, cors, logger, metrics
│   │       └── dto/           # request/response struct
│   └── infrastructure/
│       └── mysql/             # implementasi repository untuk MySQL
├── pkg/
│   ├── config/                 # load .env / viper
│   ├── database/                # koneksi MySQL (sqlx)
│   ├── logger/                  # structured logger
│   ├── response/                 # helper JSON response konsisten
│   └── validator/                # wrapper validasi input
├── migrations/                  # file .up.sql / .down.sql (golang-migrate)
├── monitoring/
│   ├── prometheus/
│   │   └── prometheus.yml
│   └── grafana/
│       ├── provisioning/
│       └── dashboards/
├── docs/                         # dokumentasi ini
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
├── .env.example
└── .gitignore
```

Alasan strukturnya sama persis pola `kissaki-backend`: supaya siapa pun yang
sudah paham project referensi bisa langsung navigasi tanpa belajar ulang
konvensi baru. Detail tanggung jawab tiap layer ada di
[01-architecture-and-tech-stack.md](01-architecture-and-tech-stack.md).

## Langkah 2 — Inisialisasi Go module

```bash
mkdir pisau-pedia-backend && cd pisau-pedia-backend
go mod init github.com/<username>/pisau-pedia-backend
```

Dependency inti yang akan ditambahkan bertahap (jangan `go get` semua
sekaligus di awal — tambahkan saat benar-benar dipakai di modul terkait):

| Dependency | Dipakai di modul | Kapan ditambahkan |
|---|---|---|
| `github.com/labstack/echo/v4` | semua handler | Fase User/Auth |
| `github.com/go-sql-driver/mysql` | infrastructure/mysql | Fase User/Auth |
| `github.com/jmoiron/sqlx` | infrastructure/mysql | Fase User/Auth |
| `github.com/golang-jwt/jwt/v5` | auth middleware & usecase | Fase User/Auth |
| `golang.org/x/crypto` (bcrypt) | usecase auth | Fase User/Auth |
| `github.com/go-playground/validator/v10` | dto validation | Fase User/Auth |
| `github.com/rs/zerolog` | pkg/logger | Fase User/Auth |
| `github.com/spf13/viper` | pkg/config | Fase User/Auth |
| `github.com/google/uuid` | entity ID | Fase User/Auth |
| `github.com/prometheus/client_golang` | metrics middleware | Fase User/Auth (barengan modul pertama) |
| `github.com/golang-migrate/migrate/v4` | CLI migration (opsional lokal) | Sebelum modul User/Auth |

## Langkah 3 — Siapkan file konfigurasi environment

Buat `.env.example` (nilai dummy, aman untuk di-commit):

```env
APP_NAME=pisau-pedia-backend
APP_PORT=8080
APP_ENV=development
APP_DEBUG=true

DB_HOST=mysql
DB_PORT=3306
DB_USER=pisaupedia
DB_PASSWORD=pisaupedia_secret_change_me
DB_NAME=pisau_pedia
DB_PARAMS=parseTime=true&charset=utf8mb4&loc=Local

JWT_SECRET=change-this-to-a-random-string
JWT_ACCESS_EXPIRY_MINUTES=60
JWT_REFRESH_EXPIRY_HOURS=168

FRONTEND_URL=http://localhost:3000

GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin_change_me
```

`.env` (hasil `cp .env.example .env`) masuk `.gitignore`. Detail tiap
variabel dan artinya dijabarkan lengkap di
[07-docker-and-deployment.md](07-docker-and-deployment.md).

## Langkah 4 — `.gitignore` awal

Minimal exclude: `.env`, `/tmp`, binary hasil `go build` (mis. `/bin`,
`pisau-pedia-backend`), `*.log`, folder `uploads/` kalau nanti dipakai untuk
gambar produk lokal.

## Langkah 5 — Skeleton `docker-compose.yml` (tanpa app dulu)

Di tahap inisiasi, cukup pastikan **MySQL bisa menyala dan bisa diakses**
sebelum menulis satu baris kode Go. Service `app`, `prometheus`, dan
`grafana` baru dilengkapi di
[07-docker-and-deployment.md](07-docker-and-deployment.md) — di sini hanya
service `mysql` untuk validasi environment:

```yaml
services:
  mysql:
    image: mysql:8.0
    restart: unless-stopped
    environment:
      MYSQL_DATABASE: pisau_pedia
      MYSQL_USER: pisaupedia
      MYSQL_PASSWORD: pisaupedia_secret_change_me
      MYSQL_ROOT_PASSWORD: root_secret_change_me
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

Validasi: `docker compose up -d mysql` lalu
`docker exec -it <container> mysql -u pisaupedia -p pisau_pedia` untuk
memastikan bisa login sebelum lanjut.

## Langkah 6 — `Makefile` awal (shortcut, opsional)

Target minimal yang dibutuhkan dari awal (isi lengkap tiap target dijelaskan
di dokumen modul terkait, bukan di sini):

```
run              # go run ./cmd/api
build            # go build -o bin/api ./cmd/api
migrate-create   # buat file migration baru (name=xxx)
migrate-up       # jalankan migration
migrate-down     # rollback 1 migration
docker-up        # docker compose up -d --build
docker-down      # docker compose down
lint             # golangci-lint run
test             # go test ./...
```

## Checklist keluar dari tahap inisiasi

- [ ] Struktur folder dibuat sesuai Langkah 1
- [ ] `go.mod` ada, module path benar
- [ ] `.env.example` + `.gitignore` ada dan `.env` **tidak** ter-commit
- [ ] `docker compose up -d mysql` berhasil, bisa login ke database
- [ ] Repo di-push ke Git remote (kosongan tapi terstruktur)

Setelah checklist ini selesai, lanjut ke
[01-architecture-and-tech-stack.md](01-architecture-and-tech-stack.md).
