#aşama 1 : derleme 
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Bağımlılıkları önce kopyala 
COPY go.mod go.sum  ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download



# Kaynak kodu kopyala ve derle
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/gateway

# aşama 2 : çalıştırma 
FROM alpine:3.19
RUN apk add --no-cache wget

WORKDIR /app

# Sadece binary'yi al, binary image boyutunu küçültmek için
COPY --from=builder /app/gateway .

EXPOSE 8080

# Çalıştırma
ENTRYPOINT ["./gateway"]
