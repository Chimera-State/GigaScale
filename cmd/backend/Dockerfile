 
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Bağımlılıkları önce kopyala
COPY go.mod go.sum  ./
RUN  go mod download

# Kaynak kodu kopyala ve derle
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend ./cmd/backend

# AŞaam 2 : çalıştırma  
FROM alpine:3.19

WORKDIR /app

# Sadece binary'yi al, binary image boyutunu küçültmek için
COPY --from=builder /app/backend .


EXPOSE 50051

ENTRYPOINT ["./backend"]
