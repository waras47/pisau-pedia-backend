# 06 — Observability: Prometheus & Grafana

Instrumentasi dipasang **barengan modul User & Auth** (langkah 7 di
[04-module-user-auth.md](04-module-user-auth.md)), bukan ditempel di akhir
project, supaya sejak endpoint pertama ada, sudah ada data untuk
divisualisasikan.

## Bagian 1 — Instrumentasi metrics di aplikasi Go

### Langkah 1 — Tambahkan dependency

`github.com/prometheus/client_golang` (`prometheus` + `promhttp` package).

### Langkah 2 — Definisikan metrics (`internal/delivery/http/middleware/metrics.go`)

Minimal 3 metric untuk fase ini:

```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    httpRequestsInFlight = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_requests_in_flight",
            Help: "Current number of in-flight HTTP requests",
        },
    )
)
```

Register ketiganya ke `prometheus.DefaultRegisterer` saat init.

### Langkah 3 — Middleware Echo

Middleware `Metrics()` membungkus tiap request: increment
`httpRequestsInFlight`, catat waktu mulai, panggil `next(c)`, lalu setelah
selesai: decrement in-flight, `Observe` durasi ke histogram, `Inc` counter
dengan label `status` dari `c.Response().Status`.

**Penting soal cardinality**: label `path` **wajib** memakai route pattern
Echo (`c.Path()`, hasilnya `/api/v1/products/:slug`), **bukan** `c.Request().URL.Path`
(`/api/v1/products/wusthof-santoku-18cm`) — kalau pakai path mentah,
setiap slug produk jadi label unik baru dan jumlah time series meledak
tak terbatas. Ini kesalahan umum saat instrumentasi Prometheus pertama
kali, dicatat eksplisit di sini supaya tidak terulang.

### Langkah 4 — Expose endpoint `/metrics`

```go
e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
```

Endpoint ini **tidak** dipasang di belakang `JWTAuth()` (Prometheus scraper
tidak punya token) — tapi juga tidak boleh publik ke internet di
production. Solusinya: expose port app hanya ke jaringan internal Docker
Compose (`prometheus` container bisa akses `app:8080/metrics` via internal
network), dan kalau perlu akses eksternal, batasi lewat Nginx (allow hanya
dari IP Prometheus) — detail konfigurasi ada di
[07-docker-and-deployment.md](07-docker-and-deployment.md).

### Langkah 5 (opsional, disiapkan tapi tidak wajib fase ini) — Metrics khusus domain

Kalau nanti mau lebih dalam dari sekadar HTTP metrics: counter
`auth_login_attempts_total{result="success|failure"}` dan
`auth_register_total`. Disebut di sini sebagai catatan untuk fase
selanjutnya, **tidak wajib** diimplementasikan bareng modul User/Auth
pertama kali — jangan blokir progres modul demi metric tambahan ini.

## Bagian 2 — Prometheus (scraper)

### Langkah 6 — `monitoring/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "pisau-pedia-backend"
    static_configs:
      - targets: ["app:8080"]
    metrics_path: /metrics

  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
```

`app` di sini adalah nama service Docker Compose (DNS internal Compose
network), bukan `localhost` — karena Prometheus berjalan di container
terpisah.

### Langkah 7 — Service Prometheus di Docker Compose

Ditambahkan penuh di [07-docker-and-deployment.md](07-docker-and-deployment.md)
(image `prom/prometheus`, mount `prometheus.yml`, port `9090`).

### Langkah 8 — Verifikasi

Buka `http://localhost:9090/targets` → target `pisau-pedia-backend` harus
`UP`. Kalau `DOWN`, cek: apakah container `app` sudah expose port 8080 di
network Compose yang sama dengan Prometheus, dan apakah `/metrics` bisa
diakses manual dari dalam container Prometheus
(`docker exec -it <prometheus-container> wget -qO- http://app:8080/metrics`).

## Bagian 3 — Grafana (dashboard)

### Langkah 9 — Provisioning datasource otomatis

`monitoring/grafana/provisioning/datasources/prometheus.yml`:

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

File ini membuat Grafana otomatis punya datasource Prometheus tanpa
klik manual setiap kali container di-recreate.

### Langkah 10 — Dashboard awal (provisioning otomatis)

`monitoring/grafana/provisioning/dashboards/dashboard.yml` (menunjuk ke
folder JSON dashboard), lalu satu file JSON dashboard minimal
(`monitoring/grafana/dashboards/api-overview.json`) berisi panel:

1. **Request rate** — `sum(rate(http_requests_total[1m])) by (path)`
2. **Error rate** — `sum(rate(http_requests_total{status=~"5.."}[1m]))`
3. **p95 latency** — `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))`
4. **In-flight requests** — `http_requests_in_flight` (gauge, langsung)

Panel-panel ini dipilih karena langsung berguna untuk memantau 2 modul
fase ini (traffic auth & product) tanpa perlu domain metric khusus dulu.

### Langkah 11 — Verifikasi

Login ke `http://localhost:3000` (port Grafana default sebelum dipetakan di
Compose — lihat mapping port aktual di
[07-docker-and-deployment.md](07-docker-and-deployment.md)), buka dashboard
"API Overview", generate traffic (`curl` beberapa endpoint dari
[03-api-contract.md](03-api-contract.md)) berulang, pastikan grafik
bergerak.

## Checklist keluar dari dokumen ini

- [ ] `/metrics` mengembalikan data Prometheus text format dari aplikasi Go
- [ ] Prometheus target `pisau-pedia-backend` berstatus `UP`
- [ ] Grafana datasource + dashboard "API Overview" ter-provision otomatis
      (tidak perlu setup manual setelah `docker compose up`)
- [ ] Label `path` di metrics memakai route pattern, bukan path mentah
      (cek cardinality tidak meledak setelah request ke beberapa slug
      produk berbeda)

Lanjut ke [07-docker-and-deployment.md](07-docker-and-deployment.md) untuk
merangkai semua service (app, mysql, prometheus, grafana) jadi satu
`docker-compose.yml`.
