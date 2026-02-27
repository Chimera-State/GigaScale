# GigaScale Reservation Engine

## Proje Klasör Yapısı

```text
GIGASCALE/
├── api/
│   └── proto/
│       └── reservation.proto         # gRPC servis & mesaj tanımları
├── cmd/
│   ├── backend/
│   │   ├── main.go                   # Backend gRPC stub giriş noktası
│   │   └── Dockerfile                # Multi-stage backend image
│   └── gateway/
│       ├── main.go                   # Gateway HTTP sunucu giriş noktası
│       └── Dockerfile                # Multi-stage gateway image
├── internal/
│   ├── gateway/
│   │   ├── handler.go                # HTTP → gRPC proxy handler
│   │   └── models.go                 # HTTP request/response modelleri
│   └── pb/
│       └── reservationv1/
│           └── reservation.go        # gRPC client stub (geçici)
├── tests/
│   └── postman/
│       └── gigascale-min.json        # Minimum API test seti
├── docker-compose.yml                # Tüm altyapıyı ayağa kaldıran dosya
├── gigascale.ps1                     # Docker build/run yardımcı script (opsiyonel)
├── go.mod                            # Go modül tanımı
├── go.sum                            # Go bağımlılık checksum'ları
└── README.md                         # Bu dosya
```






## Docker Komutları


### Servisleri Başlatma
```bash
# Image'ları derle
docker compose build

# Servisleri arka planda başlat
docker compose up -d

# Tek komutla derle ve başlat
docker compose up -d --build
```

### Durum Kontrolü
```bash
docker ps
```
`gigascale-gateway`, `gigascale-backend` ve `gigascale-redis` container'larının **Up** durumunda olduğunu görmelisiniz.

### Logları İzleme
```bash
# Tüm servislerin loglarını canlı izle
docker compose logs -f

# Sadece belirli bir servisin logları
docker compose logs -f gateway
docker compose logs -f backend
```

### Servisi Durdurma
```bash
# Durdurmak (container'lar korunur)
docker compose stop


### Yeniden Derleme
```bash

docker compose down
docker compose up -d --build
```

### Testler (Postman)
- `tests/postman/gigascale-min.json` dosyasını Postman'e import edin.
- `/api/v1/reserve` endpoint'i için örnek request body ve header tanımlarını bulabilirsiniz.

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



