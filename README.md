# GigaScale Reservation Engine

## Proje Klasör Yapısı

```text
GIGASCALE/
├── api/
│   └── proto/
│       └── reservation.proto         # gRPC servis & mesaj tanımları
├── cmd/
│   ├── backend/
│   │   ├── main.go                   # Backend gRPC sunucu giriş noktası
│   │   └── Dockerfile                # Multi-stage backend image
│   └── gateway/
│       ├── main.go                   # Gateway HTTP sunucu giriş noktası
│       └── Dockerfile                # Multi-stage gateway image
├── internal/
│   ├── gateway/
│   │   ├── handler.go                # HTTP → gRPC proxy handler
│   │   └── models.go                 # HTTP request/response modelleri
│   ├── backend/
│       ├── service/                  # gRPC servis implementasyonu
│       └── redisclient/
│           └── redis.go              # Redis bağlantısı, Connection Pool & Health Check
│  
├── k6/
│   └── basic-get.js                  # k6 yük testi scripti (10 VU, 30s)
├── tests/
│   └── postman/
│       └── gigascale-min.json        # Minimum API test seti
├── docker-compose.yml                # Uygulama altyapısını ayağa kaldıran dosya
├── docker-compose-monitoring.yml     # Prometheus + Grafana + node-exporter monitoring stack'i
├── prometheus.yml                    # Prometheus scrape configuration (gateway, backend, node-exporter)
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── prometheus.yml        # Prometheus veri kaynağını otomatik sağlayan config
│       └── dashboards/
│           ├── dashboard.yml         # GigaScale dashboard provider tanımı
│           └── gigascale.json        # GigaScale monitoring dashboard'u
├── go.mod                            # Go modül tanımı
├── go.sum                            # Go bağımlılık checksum'ları
└── README.md                         # Bu dosya
```

---

## Servisler

| Servis       | Container               | Port   | Açıklama |
| Backend      | `gigascale-backend`     | `50051` (internal) | gRPC sunucu, dışarıya kapalı|
| Gateway      | `gigascale-gateway`     | `8080` | HTTP → gRPC proxy, dışarıya açık |
| Redis        | `gigascale-redis`       | `6379` (internal) | Kalıcı veri deposu |
| RedisInsight | `gigascale-redisinsight`| `5540` | Redis görsel izleme arayüzü |
| k6           | `gigascale-k6`          | —      | Yük testi aracı (komut bazlı çalışır) |

---

## Monitoring & Observability (Prometheus + Grafana)

Bu projede altyapı ve Go runtime metriklerini izlemek için ayrı bir monitoring stack tanımlıdır:

- **Prometheus** (`prom/prometheus:latest`) – metrikleri toplar
- **Grafana** (`grafana/grafana:latest`) – dashboard ve görselleştirme
- **Node Exporter** (`prom/node-exporter:latest`) – host OS metrikleri (CPU, bellek vb.)
- **InfluxDB** – k6 yük testleri için zaman serisi depolama

### Monitoring stack'i başlatma

Önce ana uygulama stack'ini ayağa kaldır:

```bash
docker compose up -d --build
```

Ardından monitoring stack'i başlat:

```bash
docker compose -f docker-compose-monitoring.yml up -d
```

> Not: `docker-compose-monitoring.yml` dosyası, ana stack'in oluşturduğu `gigascale-internal` network'ünü kullanır; bu yüzden önce ana `docker-compose.yml` çalıştırılmalıdır.

### Uygulamalara erişim

- **Gateway API**: `http://localhost:8080`
- **Prometheus UI**: `http://localhost:9090`
- **Grafana UI**: `http://localhost:3000`
  - Anonymous auth açıktır; doğrudan erişebilirsin.
  - Sol menüden **Dashboards → Browse** altından **GigaScale** klasörünü seçerek `GigaScale - Gateway & Backend` dashboard'unu aç.

### GigaScale dashboard içeriği

`grafana/provisioning/dashboards/gigascale.json` dosyası ile aşağıdaki paneller otomatik olarak sağlanır:

- **Prometheus HTTP RPS**  
  - Sorgu: `sum(rate(prometheus_http_requests_total[5m]))`
- **Prometheus HTTP Latency p50/p95/p99**  
  - Sorgular: `histogram_quantile(0.50|0.95|0.99, sum(rate(prometheus_http_request_duration_seconds_bucket[5m])) by (le))`
- **Prometheus HTTP Error Rate (5xx oranı)**  
  - Sorgu: `sum(rate(prometheus_http_requests_total{code=~"5.."}[5m])) / sum(rate(prometheus_http_requests_total[5m]))`
- **HTTP 429 Rate Limit Oranı (örnek)**  
  - Sorgu: `sum(rate(prometheus_http_requests_total{code="429"}[5m])) / sum(rate(prometheus_http_requests_total[5m]))`
- **Go Runtime**  
  - `go_goroutines` (goroutine sayısı)
  - `go_memstats_heap_alloc_bytes` (heap kullanımı)

> Not: Şu anda gateway/backend uygulamaları kendi iş metriklerini (`/metrics` endpoint'i, RPS, latency, hata oranı, rate limit metrikleri vb.) expose etmemektedir.  
> Backend/gateway ekipleri Prometheus client entegrasyonunu eklediğinde, bu dashboard üzerindeki PromQL sorguları ilgili metrik isimleriyle güncellenerek uygulama seviyesinde detaylı gözlemlenebilirlik sağlanabilir.

---

## Docker Komutları

### Servisleri Başlatma
```bash
# Image'ları derle ve servisleri arka planda başlat
docker compose up -d --build
```

### Durum Kontrolü
```bash
docker ps
```


### Logları İzleme
```bash
# Tüm servislerin loglarını canlı izle
docker compose logs -f

# Sadece belirli bir servisin logları
docker compose logs -f backend
docker compose logs -f gateway
```

### Servisi Durdurma & Yeniden Derleme
```bash
# Durdur ve yeniden başlat
docker compose down
docker compose up -d --build
```

---

## k6 Yük Testi (GIGA-42)

k6, Docker üzerinden çalışır — ayrıca kurulum gerekmez.

```bash
# Temel yük testini çalıştır 
docker compose run --rm k6 run /scripts/basic-get.js
```

Test tamamlandığında terminalde istek sayısı, başarı oranı ve yanıt süreleri görüntülenir.


---

## Redis Bağlantısı & Health Check 

Backend başladığında otomatik olarak:
1. Redis'e bağlanır (`REDIS_ADDR` ortam değişkeninden adresi alır, yoksa `localhost:6379` kullanır)
2. `PING` komutu ile bağlantıyı doğrular



## RedisInsight — Görsel İzleme 

Tarayıcıda **`http://localhost:5540`** adresini aç.

İlk bağlantı için:
1. "Add Redis Database" butonuna tıkla
2. **Host:** `redis` | **Port:** `6379`
3. "Add Redis Database" ile kaydet

Redis içindeki anahtarları, belllek kullanımını ve komut geçmişini buradan izleyebilirsin.

---

## API Referansı

### `POST /api/v1/reserve`

Koltuk rezervasyonu oluşturur.

**Request Body:**
```json
{
  "user_id": "u-001",
  "trip_id": "t-100",
  "seat_id": "A1",
  "Idempotency_key": "key-abc-123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Koltuk başarıyla rezerve edildi."
}
```

### Testler (Postman)
`tests/postman/gigascale-min.json` dosyasını Postman'e import edin.
