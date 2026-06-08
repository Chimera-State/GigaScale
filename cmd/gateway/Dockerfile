#aşama 1 : derleme 
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Bağımlılıkları önce kopyala 
COPY go.mod go.sum  ./
RUN go mod download



# Kaynak kodu kopyala ve derle
COPY . .
RUN GOOS=linux go build -o gateway ./cmd/gateway

# aşama 2 : çalıştırma 
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Sadece binary'yi al, binary image boyutunu küçültmek için
COPY --from=builder /app/gateway .

EXPOSE 8080

# Çalıştırma
ENTRYPOINT ["./gateway"]
