# GigaScale Reservation Engine

GigaScale is a highly scalable, distributed reservation and payment orchestration engine. Built with a microservices architecture, it demonstrates advanced patterns such as Saga orchestration, distributed locking, idempotency, and end-to-end distributed tracing.

## Architecture & Services

The system is composed of several specialized microservices and infrastructure components:

| Service | Container | Port | Description |
|---------|-----------|------|-------------|
| **Gateway** | `gigascale-gateway` | `8080` | HTTP to gRPC proxy. Handles rate limiting, validation, and Saga orchestration. |
| **Backend** | `gigascale-backend` | `50051` | gRPC server. Manages core reservation logic, distributed locks, and database transactions. |
| **Payment** | `gigascale-payment` | `50052` | gRPC server. Simulates payment processing with artificial latency and failure rates. |
| **Notifier** | `gigascale-notifier` | — | Kafka consumer. Processes successful reservation events and triggers external webhooks. |
| **PostgreSQL** | `gigascale-postgres` | `5432` | Relational database for persistent storage (reservations and notification audit logs). |
| **Redis Cluster** | 6 nodes | `6379` | Distributed caching and locking mechanism. |
| **Kafka** | `gigascale-kafka` | `9092` | Message broker for asynchronous event processing. |
| **Zookeeper** | `gigascale-zookeeper` | `2181` | Coordination service for Kafka. |

## Project Structure

```text
GIGASCALE/
├── api/proto/                # gRPC service definitions and messages
├── cmd/                      # Application entrypoints
│   ├── backend/              # Backend service main & Dockerfile
│   ├── gateway/              # Gateway service main & Dockerfile
│   ├── notifier/             # Notifier service main & Dockerfile
│   └── payment/              # Payment service main & Dockerfile
├── internal/                 # Private application and library code
│   ├── backend/              # Backend business logic, repositories, and Redis integration
│   ├── gateway/              # HTTP handlers, orchestrator (Saga), rate limiters, MQ publisher
│   ├── notifier/             # Webhook client and notification repository
│   └── payment/              # Payment processing logic
├── migrations/               # PostgreSQL schema migrations
├── k6/                       # Load testing scripts
├── tests/postman/            # API test collections
├── docker-compose.yml        # Main infrastructure deployment configuration
└── docker-compose-monitoring.yml # Monitoring and observability stack
```

## Getting Started

### Prerequisites

- Docker and Docker Compose installed on your system.

### Configuration

Before starting the services, configure the environment variables:

```bash
cp .env.example .env
```

If you wish to test the webhook notification flow, update the `WEBHOOK_URL` in your `.env` file (e.g., using a service like webhook.site). If left empty, the notifier will operate in log-only mode.

### Running the Services

Deploy the main application stack:

```bash
docker compose up -d --build
```

To verify that all containers are running successfully:

```bash
docker ps
```

To view the application logs:

```bash
docker compose logs -f
```

## API Reference

### Create Reservation

`POST /api/v1/reserve`

Initiates a seat reservation process.

**Request Body:**
```json
{
  "user_id": "u-001",
  "trip_id": "550e8400-e29b-41d4-a716-446655440000",
  "seat_id": "A1",
  "idempotency_key": "key-abc-123",
  "amount": 100.00
}
```

**Response (Success):**
```json
{
  "success": true,
  "payment_id": "uuid-string",
  "message": "Reservation and payment completed."
}
```

## Advanced Workflows

### Saga Orchestration & Idempotency

The Gateway service acts as the orchestrator. When a reservation request is received:
1. A Redis lock is acquired to prevent concurrent modifications for the same seat.
2. The Backend service is called to persist the reservation.
3. The Payment service is called to process the transaction.
4. If payment fails, a compensating transaction (CancelReservation) is triggered on the Backend.
5. All operations utilize an `idempotency_key` to ensure that duplicate requests do not result in duplicate reservations or payments.

### Asynchronous Notifications

Upon a successful transaction, the Gateway publishes a `reservations.created` event to Kafka. The Notifier service consumes this event, logs the attempt in PostgreSQL for audit purposes (ensuring at-most-once delivery via idempotency checks), and sends an HTTP POST request to the configured `WEBHOOK_URL`.

## Observability and Monitoring

The project includes a comprehensive observability stack based on OpenTelemetry, Prometheus, and Grafana.

### Note on OpenTelemetry Integration

All OpenTelemetry instrumentation in this project (including context propagation across HTTP, gRPC, and Kafka boundaries) is powered by the **[go-otel-kit](https://github.com/Chimera-State/go-otel-kit)** library. If you are interested in how the telemetry infrastructure is abstracted and implemented, please check out the `go-otel-kit` repository.

### Starting the Monitoring Stack

Ensure the main stack is running, then start the monitoring components:

```bash
docker compose -f docker-compose-monitoring.yml up -d
```

### Accessing the Dashboards

- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000` (Anonymous access enabled)
- **RedisInsight**: `http://localhost:5540` (Visual interface for Redis)

In Grafana, navigate to **Dashboards -> Browse -> GigaScale** to view predefined metrics including RPS, Latency, Error Rates, and Go Runtime statistics.

All services are instrumented with OpenTelemetry, propagating trace context across HTTP, gRPC, and Kafka boundaries. Traces can be exported to Jaeger or any OTLP-compatible backend (requires additional configuration in the collector).

## Load Testing

The repository includes k6 scripts for performance validation.

```bash
docker compose run --rm k6 run /scripts/basic-get.js
```

This will execute a predefined load test against the API gateway and output the performance metrics directly to your terminal.
