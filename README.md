# GigaScale

GigaScale, yüksek ölçekli rezervasyon sistemi projesidir. Bu proje, servisler arası iletişim için Protocol Buffers (Proto) ve gRPC kullanmaktadır.

## Geliştirme Ortamı Kurulumu

Projede proto dosyalarından Go kodlarını üretmek için aşağıdaki adımları takip etmelisiniz.

### 1. Protobuf Derleyicisi (protoc) Kurulumu

1. [Protobuf Releases](https://github.com/protocolbuffers/protobuf/releases) sayfasına gidin.
2. İşletim sisteminize uygun dosyayı indirin (Windows için: `protoc-xx.x-win64.zip`).
3. Zip içerisindeki `bin/protoc.exe` dosyasını sisteminizde PATH'e ekli olan bir klasöre kopyalayın. (Tavsiye: `C:\Users\Oxur\go\bin`)
4. Terminalde `protoc --version` komutu ile kurulumu doğrulayın.

### 2. Go Proto ve gRPC Eklentileri

`protoc` derleyicisinin Go kodu üretebilmesi için şu eklentileri kurmalısınız:

```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. Makefile ve Kod Üretimi

Proje kök dizininde bir `Makefile` bulunmaktadır. Bu dosya, karmaşık terminal komutlarını tek bir komutla çalıştırmanızı sağlar.

#### Makefile İçeriği:
```makefile
PROTO_DIR=api/proto
OUT_DIR=internal/pb

.PHONY: proto clean

proto:
	@if not exist internal\pb mkdir internal\pb
	protoc --proto_path=$(PROTO_DIR) \
	       --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/*.proto

clean:
	@if exist internal\pb rmdir /s /q internal\pb
```

#### Komutların Kullanımı:

- **Proto kodlarını üretmek için:**
  ```bash
  make proto
  ```
  Bu komut `api/proto` altındaki dosyaları okur ve `internal/pb` altına Go modellerini üretir.

- **Üretilen dosyaları temizlemek için:**
  ```bash
  make clean
  ```

## Proje Yapısı

- `api/proto/`: Sistemin arayüz tanımları (.proto dosyaları).
- `internal/pb/`: 自動 üretilen Go kodları.
- `cmd/`: Uygulama giriş noktaları.
- `internal/`: İş mantığı ve HTTP handler'ları.

