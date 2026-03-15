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
yalnızca dışarıdan erişilmesi zorunlu olan portlara izin verilmelidir. Dış dünyaya sadece **22, 80 ve 443** açılır;
8080, 50051, 50052, 6379, 5432 dışarıya kapalı kalır (sadece iç ağda kullanılır).

```bash
# Varsayılan: gelen tüm trafik reddedilir
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Gerekli portları dış erişime açın:
sudo ufw allow 22/tcp   # Sunucu yönetimi için SSH erişimi
sudo ufw allow 80/tcp   # HTTP (Caddy reverse proxy)
sudo ufw allow 443/tcp  # HTTPS (Caddy, Let's Encrypt)

# Güvenlik duvarını aktifleştirin (önce 22'yi açtığınızdan emin olun):
sudo ufw enable
sudo ufw status numbered
```

---

## 3. Servis ve Port Güvenliği Haritası

Sistemdeki container'ların port kullanım durumları ve dışa açıklık seviyeleri aşağıdaki gibidir.
Dış erişim Caddy üzerinden 80/443 ile yapılır; API Gateway (8080) sadece iç ağda Caddy tarafından kullanılır.

| Servis      | Port  | Dışarıya Açık?    | Açıklama                                              |
|-------------|-------|-------------------|--------------------------------------------------------|
| Caddy       | 80    | Evet              | HTTP, reverse proxy (Gateway'e yönlendirir)            |
| Caddy       | 443   | Evet              | HTTPS, TLS sonlandırma (Let's Encrypt)                 |
| API Gateway | 8080  | Hayır (Internal)  | Sadece iç ağda; Caddy bu porta proxy yapar             |
| Backend     | 50051 | Hayır (Internal)  | gRPC ile iç ağda haberleşir                            |
| Payment     | 50052 | Hayır (Internal)  | gRPC ile iç ağda haberleşir                           |
| Redis       | 6379  | Hayır (Internal)  | Önbellek servisi, dışarıdan erişilemez                |
| PostgreSQL  | 5432  | Hayır (Internal)  | Ana veritabanı, dışarıdan erişilemez                  |

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
