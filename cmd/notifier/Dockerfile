FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o notifier ./cmd/notifier

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/notifier .

ENTRYPOINT ["./notifier"]
