# GigaScale Production Deployment

---

## 1. Prod Redis ve Genel Güvenlik Yapılandırması

> **ÖNEMLİ:** Redis servisi production ortamında **kesinlikle** dış dünyaya açılmamalıdır.

Redis sadece Docker'ın `internal` ağı üzerinden diğer mikroservislerle iletişim kurmalıdır.
`docker-compose.yml` dosyasında Redis için port yönlendirmesi (`ports: "6379:6379"`) yapılmamış
olduğundan emin olun. Canlı ortamda RedisInsight gibi GUI araçları kullanılmaz.

---

## 2. UFW (Güvenlik Duvarı) Ayarları

Sunucu güvenliğini sağlamak için Ubuntu üzerindeki UFW (Uncomplicated Firewall) yapılandırmasında
yalnızca dışarıdan erişilmesi zorunlu olan portlara izin verilmelidir.

```bash
# Gerekli portları dış erişime açın:
sudo ufw allow 22/tcp    # Sunucu yönetimi için SSH erişimi
sudo ufw allow 8080/tcp  # API Gateway (Kullanıcıların sisteme erişim noktası)

# Güvenlik duvarını aktifleştirin:
sudo ufw enable
```

---

## 3. Servis ve Port Güvenliği Haritası

Sistemdeki container'ların port kullanım durumları ve dışa açıklık seviyeleri aşağıdaki gibidir:

| Servis      | Port  | Dışarıya Açık?    |Açıklama                                       |
|-------------|-------|-------------------|-----------------------------------------------|
| API Gateway | 8080  |  Evet             | HTTP üzerinden gelen dış trafiği karşılar     |
| Backend     | 50051 |  Hayır (Internal) | gRPC ile iç ağda haberleşir                   |
| Payment     | 50052 |  Hayır (Internal) | gRPC ile iç ağda haberleşir                   |
| Redis       | 6379  |  Hayır (Internal) | Önbellek servisi, dışarıdan erişilemez        |
| PostgreSQL  | 5432  | Hayır (Internal)  | Ana veritabanı, dışarıdan erişilemez          |

---

## 4. Çevresel Değişkenler (.env)

Projeyi sunucuda ayağa kaldırmadan önce kök dizinde bir `.env` dosyası oluşturulmalıdır.
Aşağıdaki değişkenlerin projeye uygun, güvenli değerlerle tanımlanması zorunludur:

```env
# PostgreSQL Bağlantı Ayarları
POSTGRES_USER=gigascale_user
POSTGRES_PASSWORD=kendi_guvenli_sifrenizi_yazin
POSTGRES_DB=gigascale
```

---

## 5. Veritabanı Migration İşlemi

Container'lar ayağa kalktıktan sonra, Payment ve Backend servislerinin veri yazabilmesi için
veritabanı tablolarının oluşturulması (migration) gerekir.

`docker compose up -d` ile sistemi başlattıktan sonra şu komutu çalıştırın:

```bash
# Postgres container'ı içinde migration dosyasını çalıştırarak tabloları oluşturur
docker compose exec postgres psql -U $POSTGRES_USER -d $POSTGRES_DB -f /migration.sql
```
