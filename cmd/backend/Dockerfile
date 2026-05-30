
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Bağımlılıkları önce kopyala
COPY go.mod go.sum  ./
RUN go mod download


COPY . .
RUN GOOS=linux go build -o backend ./cmd/backend

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/* && wget -qO /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.25/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe


WORKDIR /app

COPY --from=builder /app/backend .

EXPOSE 50051

ENTRYPOINT ["./backend"]
