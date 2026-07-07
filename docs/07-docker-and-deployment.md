# 07 — Docker & Deployment

Dokumen ini merangkai semua service (`app`, `mysql`, `prometheus`,
`grafana`, opsional `nginx`) jadi satu `docker-compose.yml`, plus
`Dockerfile` multi-stage untuk aplikasi Go. Dikerjakan **setelah** modul
User/Auth dan Product selesai secara lokal (`go run ./cmd/api` langsung,
MySQL via Docker saja seperti di
[00-project-initiation.md](00-project-initiation.md)) — supaya saat
containerize aplikasi, bug yang muncul jelas berasal dari
Dockerfile/networking, bukan tercampur dengan bug logic yang belum
ketemu.

## Langkah 1 — `Dockerfile` (multi-stage build)

```dockerfile
# ---- Build stage ----
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# ---- Run stage ----
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/api .
COPY migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./api"]
```

Alasan multi-stage: image akhir tidak membawa toolchain Go (ukuran image
jauh lebih kecil, permukaan serangan lebih kecil).

## Langkah 2 — `docker-compose.yml` lengkap

```yaml
services:
  app:
    build: .
    restart: unless-stopped
    env_file: .env
    depends_on:
      mysql:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully
    ports:
      - "8080:8080"
    networks: [backend]

  migrate:
    image: migrate/migrate:v4.17.1
    volumes:
      - ./migrations:/migrations
    env_file: .env
    depends_on:
      mysql:
        condition: service_healthy
    entrypoint:
      [
        "migrate", "-path", "/migrations",
        "-database", "mysql://${DB_USER}:${DB_PASSWORD}@tcp(mysql:3306)/${DB_NAME}",
        "up",
      ]
    networks: [backend]

  mysql:
    image: mysql:8.0
    restart: unless-stopped
    environment:
      MYSQL_DATABASE: ${DB_NAME}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASSWORD}
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p${DB_ROOT_PASSWORD}"]
      interval: 5s
      timeout: 5s
      retries: 10
    networks: [backend]

  prometheus:
    image: prom/prometheus:v2.53.0
    restart: unless-stopped
    volumes:
      - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    depends_on: [app]
    networks: [backend]

  grafana:
    image: grafana/grafana:11.1.0
    restart: unless-stopped
    environment:
      GF_SECURITY_ADMIN_USER: ${GRAFANA_ADMIN_USER}
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_ADMIN_PASSWORD}
    volumes:
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
      - grafana_data:/var/lib/grafana
    ports:
      - "3001:3000"
    depends_on: [prometheus]
    networks: [backend]

networks:
  backend:

volumes:
  mysql_data:
  prometheus_data:
  grafana_data:
```

Catatan keputusan:

- **`migrate` sebagai service terpisah** dengan `depends_on:
  service_completed_successfully` pada `app` — migration selalu jalan dulu
  sebelum aplikasi start, tidak ada race condition "app start sebelum
  tabel ada".
- **Healthcheck MySQL** dibutuhkan karena `depends_on` biasa (tanpa
  `condition`) hanya menunggu container *start*, bukan menunggu MySQL
  *siap menerima koneksi* — tanpa healthcheck ini, `migrate` sering gagal
  di run pertama karena race condition klasik Docker Compose + database.
- **Grafana dipetakan ke port `3001`** (bukan `3000`) karena `3000` sudah
  dipakai oleh dev server Next.js frontend (`pisau-pedia`) — supaya
  keduanya bisa jalan bersamaan di mesin developer yang sama.
- **`/metrics` (port 8080) tidak diexpose ke luar** selain untuk keperluan
  dev lokal — di production, batasi lewat Nginx atau security group cloud
  provider (lihat Langkah 5).

## Langkah 3 — `Makefile`

```makefile
run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp(localhost:3306)/$(DB_NAME)" up

migrate-down:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp(localhost:3306)/$(DB_NAME)" down 1

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

lint:
	golangci-lint run

test:
	go test ./... -cover
```

## Langkah 4 — Environment variables (referensi lengkap)

| Variabel | Default (dev) | Keterangan |
|---|---|---|
| `APP_PORT` | 8080 | Port HTTP server |
| `APP_ENV` | development | `development` / `production` |
| `DB_HOST` | mysql | Host MySQL (nama service Compose) |
| `DB_PORT` | 3306 | Port MySQL |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | — | Kredensial database |
| `DB_ROOT_PASSWORD` | — | Untuk healthcheck & admin MySQL, **jangan** dipakai aplikasi |
| `JWT_SECRET` | — | **Wajib diganti** string random panjang di production |
| `JWT_ACCESS_EXPIRY_MINUTES` | 60 | TTL access token |
| `JWT_REFRESH_EXPIRY_HOURS` | 168 | TTL refresh token (7 hari) |
| `FRONTEND_URL` | http://localhost:3000 | Untuk CORS allow-origin |
| `GRAFANA_ADMIN_USER` / `GRAFANA_ADMIN_PASSWORD` | admin / (ganti) | Login awal Grafana |

## Langkah 5 — Production checklist (dicatat sekarang, dieksekusi saat deploy)

- [ ] `JWT_SECRET` diganti string random ≥32 byte, disimpan di secret
      manager (bukan `.env` di-commit)
- [ ] `DB_PASSWORD` & `DB_ROOT_PASSWORD` kuat, unik per environment
- [ ] `APP_ENV=production` (mematikan debug log verbose)
- [ ] Port `9090` (Prometheus) dan `3001` (Grafana) **tidak** diexpose
      langsung ke internet publik — akses lewat VPN/SSH tunnel, atau Nginx
      dengan basic auth
- [ ] `GRAFANA_ADMIN_PASSWORD` diganti dari default
- [ ] Backup terjadwal untuk volume `mysql_data`
- [ ] `FRONTEND_URL` diarahkan ke domain production frontend
- [ ] TLS/HTTPS di depan `app` (lewat Nginx atau load balancer cloud)

## Verifikasi akhir (menjalankan seluruh stack)

```bash
cp .env.example .env   # lalu isi nilai yang perlu diganti
docker compose up -d --build

curl http://localhost:8080/api/v1/categories
curl http://localhost:9090/-/healthy      # Prometheus
curl http://localhost:3001/api/health     # Grafana
```

Semua 3 harus merespons sukses sebelum stack dianggap siap. Lanjut ke
[08-roadmap-checklist.md](08-roadmap-checklist.md) untuk melihat checklist
gabungan seluruh fase dan apa yang sengaja belum dikerjakan.
