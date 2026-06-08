
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN GOOS=linux go build -o payment ./cmd/payment

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/* \
    && wget -qO /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.25/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe

WORKDIR /app

COPY --from=builder /app/payment .

EXPOSE 50052

ENTRYPOINT ["./payment"]
