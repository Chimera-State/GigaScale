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
│   │   ├── service/                  # gRPC servis implementasyonu
│   │   └── redisclient/
│   │       └── redis.go              # Redis bağlantısı, Connection Pool & Health Check
│   └── pb/
│       └── reservationv1/
│           └── reservation.go        # gRPC client stub (geçici)
├── k6/
│   └── basic-get.js                  # k6 yük testi scripti (10 VU, 30s)
├── tests/
│   └── postman/
│       └── gigascale-min.json        # Minimum API test seti
├── docker-compose.yml                # Tüm altyapıyı ayağa kaldıran dosya
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
