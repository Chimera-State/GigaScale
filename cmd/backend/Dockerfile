 
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

RUN wget -qO /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.25/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe


WORKDIR /app

# Sadece binary'yi al, binary image boyutunu küçültmek için
COPY --from=builder /app/backend .


EXPOSE 50051

ENTRYPOINT ["./backend"]
