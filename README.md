# GigaScale 

Şu anki aşama: **Sprint 1 - Foundation (GIGA-17, GIGA-18, GIGA-19)**


### 1. Altyapıyı Başlatın (Docker Compose)
Backend ve diğer servislerin ihtiyaç duyduğu ağ ve veritabanı altyapısını ayağa kaldırmak için:

```bash
docker-compose up -d
```

Bu komut şunları yapacaktır:
- `gigascale-internal`: Servislerin kendi aralarında haberleşeceği kapalı ağ oluşturulur.
- `gigascale-public`: Dış dünyaya açılacak servisler (Örn: API Gateway) için ağ oluşturulur.
- `gigascale-redis`: Veritabanı container'ı başlatılır. **Sadece internal ağa bağlıdır.**

> **Backend Geliştiricileri İçin Not:**
> Go kodunuzu henüz Docker içine almadan kendi bilgisayarınızda çalıştırırken Redis'e erişebilmeniz için, `docker-compose.yml` içinde Redis 6379 portu geçici olarak `localhost`'a açılmıştır. Kodunuzda Redis bağlantı adresini `localhost:6379` (veya `127.0.0.1:6379`) olarak kullanabilirsiniz.

### 2. Durumu Kontrol Edin
Container'ların sorunsuz çalıştığından emin olmak için:
```bash
docker ps
```
`gigascale-redis` container'ının `Up` durumunda olduğunu görmelisiniz.

### 3. Testleri İnceleyin (Postman)
Minimum API standartlarımızı ve payload yapımızı görmek için:
- `tests/postman/gigascale-min.json` dosyasını Postman'e (veya Insomnia'ya) import edin.
- `/api/v1/reserve` endpoint'i için örnek request body ve header tanımlarını bulabilirsiniz.
- Sizin geliştirdiğiniz Go API'si ayağa kalktığında ilk testleri bu collection üzerinden yapacağız.

## Altyapıyı Kapatmak
Çalışmanız bittiğinde veya sistemi sıfırlamak istediğinizde:
```bash
# Sadece durdurmak için:
docker-compose stop

# Durdurup ağları/containerları silmek için:
docker-compose down
```

## Proje Klasör Yapısı (Mevcut Durum)
```text
GİGASCALE/
├── api/
│   └── proto/                  # gRPC proto tanımları
├── cmd/
│   └── gateway/                # API Gateway giriş noktası
├── internal/
│   └── gateway/                # Gateway iç mantığı
├── go.mod                      # Go modül tanımı
├── docker-compose.yml          # Tüm altyapıyı ayağa kaldıran dosya
├── README.md                   # Proje dokümantasyonu
└── tests/
    └── postman/
        └── gigascale-min.json  # Minimum API endpoint test seti
```