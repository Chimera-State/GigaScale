#aşama 1 : derleme 
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Bağımlılıkları önce kopyala 
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodu kopyala ve derle
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/gateway

# aşama 2 : çalıştırma 
FROM alpine:3.19

WORKDIR /app

# Sadece binary'yi al, binary image boyutunu küçültmek için
COPY --from=builder /app/gateway .

EXPOSE 8080

# Çalıştırma
ENTRYPOINT ["./gateway"]
