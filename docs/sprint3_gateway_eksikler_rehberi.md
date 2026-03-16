# Sprint 3 Gateway Eksikleri — Detaylı Uygulama Rehberi

Bu rehber, Sprint 3 Gateway görevlerinden **henüz yapılmamış** kısımların nasıl ekleneceğini adım adım, öğretici şekilde anlatır.

---

## İçindekiler

1. [GIGA-72.1: Log Format Standardizasyonu](#1-giga-721-log-format-standardizasyonu)
2. [GIGA-72.2: Request Duration Logging](#2-giga-722-request-duration-logging)
3. [GIGA-75.1: Orchestrator Interface](#3-giga-751-orchestrator-interface)
4. [GIGA-75.2: Payment Client Hazırlık](#4-giga-752-payment-client-hazırlık)
5. [GIGA-75.3: Server Struct Güncellemesi](#5-giga-753-server-struct-güncellemesi)
6. [Küçük Düzeltmeler](#6-küçük-düzeltmeler)

---

## 1. GIGA-72.1: Log Format Standardizasyonu

**Dosya:** `internal/gateway/middleware.go`

**Hedef:** Log prefix'leri standart olsun; rate limit blokta ayrı log yazılsın.

---

### 1.1 Neden?

- `[REQUEST]` → Tüm istek logları aynı prefix ile aranabilir
- `[RATE_LIMIT_BLOCK]` → Rate limit olayları ayrı filtrelenebilir
- Ops/debug için tutarlı format

---

### 1.2 Değişiklik 1: İstek başlangıç logu

**Mevcut (satır 41):**
```go
log.Printf("[REQUESTED] ip=%s method=%s path=%s time=%s", ip, r.Method, r.URL.Path, start.Format(time.RFC3339))
```

**Yeni:**
```go
log.Printf("[REQUEST] ip=%s method=%s path=%s", ip, r.Method, r.URL.Path)
```

**Açıklama:** `[REQUESTED]` → `[REQUEST]`, `time=%s` kaldırıldı (duration sonra loglanacak).

---

### 1.3 Değişiklik 2: Rate limit blokta log ekleme

**Mevcut (satır 47-50):**
```go
if !isAllowed {
    http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
    return
}
```

**Yeni:**
```go
if !isAllowed {
    log.Printf("[RATE_LIMIT_BLOCK] ip=%s path=%s", ip, r.URL.Path)
    http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
    return
}
```

**Açıklama:** Limit aşıldığında hangi IP ve path'in bloklandığı loglanır.

---

### 1.4 Sonuç (middleware.go — ilgili bölüm)

```go
func (h *rateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ip := clientIP(r)

    log.Printf("[REQUEST] ip=%s method=%s path=%s", ip, r.Method, r.URL.Path)

    ctx := r.Context()
    isAllowed := h.server.limiter.Allow(ctx, ip)

    if !isAllowed {
        log.Printf("[RATE_LIMIT_BLOCK] ip=%s path=%s", ip, r.URL.Path)
        http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
        return
    }

    h.next.ServeHTTP(w, r)
}
```

---

## 2. GIGA-72.2: Request Duration Logging

**Dosya:** `internal/gateway/middleware.go`

**Hedef:** Her tamamlanan istek için süre (duration) ve HTTP status code loglansın.

---

### 2.1 Sorun: Status code nasıl yakalanır?

`http.ResponseWriter` üzerinden yazıldıktan sonra status code'u doğrudan alamayız. Handler `w.WriteHeader(200)` veya `http.Error(w, ..., 429)` çağırır; biz bu değeri göremeyiz.

**Çözüm:** Response'u saran bir wrapper (`responseRecorder`) kullan. Bu wrapper `WriteHeader` çağrıldığında status code'u saklar.

---

### 2.2 responseRecorder struct

**Nereye:** `middleware.go` — `clientIP` fonksiyonundan önce (veya dosya başında)

```go
// responseRecorder: HTTP response'u sarmalayıp status code'u yakalar
type responseRecorder struct {
    http.ResponseWriter
    statusCode int
    written    bool
}

func (rr *responseRecorder) WriteHeader(code int) {
    if !rr.written {
        rr.statusCode = code
        rr.written = true
    }
    rr.ResponseWriter.WriteHeader(code)
}
```

**Açıklama:**
- `ResponseWriter` embed edilir → tüm `Write`, `WriteHeader` çağrıları geçer
- `WriteHeader` ilk kez çağrıldığında `statusCode` saklanır
- `written` ile aynı response'a birden fazla `WriteHeader` çağrısında sadece ilki sayılır

---

### 2.3 ServeHTTP'ta kullanım

**Mevcut:**
```go
h.next.ServeHTTP(w, r)
```

**Yeni:**
```go
rr := &responseRecorder{ResponseWriter: w, statusCode: 200}
h.next.ServeHTTP(rr, r)

defer func() {
    duration := time.Since(start)
    log.Printf("[REQUEST] path=%s duration=%s status=%d", r.URL.Path, duration, rr.statusCode)
}()
```

**Açıklama:**
1. `rr` → `w`'yi saran wrapper, varsayılan status 200
2. Handler `rr`'ye yazar; `WriteHeader(400)` vb. çağrılırsa `rr.statusCode` güncellenir
3. `defer` → fonksiyon bitmeden hemen önce çalışır; handler tamamlandıktan sonra duration ve status loglanır

---

### 2.4 Önemli: Rate limit blokta defer çalışmaz

Rate limit blokta `return` ile çıkıyoruz; `defer` o blokta tanımlı değil, bu yüzden orada duration logu yazılmaz. Bu istenen davranış: bloklanan istekler için `[RATE_LIMIT_BLOCK]` yeterli.

---

### 2.5 Sonuç (middleware.go — tam ServeHTTP)

```go
func (h *rateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ip := clientIP(r)

    log.Printf("[REQUEST] ip=%s method=%s path=%s", ip, r.Method, r.URL.Path)

    ctx := r.Context()
    isAllowed := h.server.limiter.Allow(ctx, ip)

    if !isAllowed {
        log.Printf("[RATE_LIMIT_BLOCK] ip=%s path=%s", ip, r.URL.Path)
        http.Error(w, "GigaScale: Too many requests! Please wait.", http.StatusTooManyRequests)
        return
    }

    rr := &responseRecorder{ResponseWriter: w, statusCode: 200}
    h.next.ServeHTTP(rr, r)

    defer func() {
        duration := time.Since(start)
        log.Printf("[REQUEST] path=%s duration=%s status=%d", r.URL.Path, duration, rr.statusCode)
    }()
}
```

---

## 3. GIGA-75.1: Orchestrator Interface

**Dosya:** `internal/gateway/orchestrator/interface.go` (yeni dosya)

**Hedef:** Sprint 4'te Saga yazılacak; önce interface ve boş yapı hazırlansın.

---

### 3.1 Klasör oluşturma

```
internal/gateway/
├── orchestrator/
│   └── interface.go   ← YENİ
├── middleware.go
├── handler.go
└── ...
```

---

### 3.2 interface.go içeriği

```go
package orchestrator

import (
    "context"
)

// ReserveRequest: HTTP'den gelen rezervasyon isteği (handler'daki ReserveHTTPRequest ile uyumlu)
type ReserveRequest struct {
    UserID         string `json:"user_id" validate:"required"`
    TripID         string `json:"trip_id" validate:"required"`
    SeatID         string `json:"seat_id" validate:"required"`
    IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

// ReserveResponse: Rezervasyon sonucu
type ReserveResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

// Orchestrator: Reserve + Payment orkestrasyonu (Sprint 4'te implement edilecek)
type Orchestrator interface {
    ReserveWithPayment(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error)
}
```

**Açıklama:**
- `ReserveWithPayment` → Rezervasyon + ödeme adımlarını orkestre edecek
- Şimdilik sadece interface; implementasyon Sprint 4'te

---

### 3.3 Stub implementasyon (opsiyonel)

Handler'ın `Orchestrator` kullanabilmesi için boş bir implementasyon:

**Dosya:** `internal/gateway/orchestrator/stub.go`

```go
package orchestrator

import "context"

// StubOrchestrator: Henüz gerçek orkestrasyon yok; sadece placeholder
type StubOrchestrator struct{}

func NewStubOrchestrator() *StubOrchestrator {
    return &StubOrchestrator{}
}

func (s *StubOrchestrator) ReserveWithPayment(ctx context.Context, req *ReserveRequest) (*ReserveResponse, error) {
    return nil, nil // Sprint 4'te gerçek implementasyon
}
```

---

## 4. GIGA-75.2: Payment Client Hazırlık

**Dosya:** `cmd/gateway/main.go`

**Hedef:** PAYMENT_ADDR env, PaymentService gRPC client. PaymentService henüz yoksa bağlantı hatasında uygulama çökmesin.

---

### 4.1 Payment proto durumu

PaymentService proto (GIGA-74/92) Backend tarafında. Henüz yoksa:
- Gateway sadece `PAYMENT_ADDR` ve bağlantı yapısını hazırlar
- Bağlantı başarısız olursa log + devam (graceful degradation)
- Proto hazır olduğunda client kodu generate edilip kullanılır

---

### 4.2 Basit yaklaşım (proto yokken)

Payment proto yoksa sadece **bağlantı hazırlığı** yapılır:

```go
// main.go içinde, conn'dan sonra

paymentAddr := getEnv("PAYMENT_ADDR", "localhost:50052")
paymentConn, err := grpc.NewClient(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Printf("Payment servisi bağlantı hatası (opsiyonel): %v", err)
    paymentConn = nil
} else {
    defer paymentConn.Close()
}
```

**Not:** `paymentConn` nil ise Server'a `nil` geçilir; orchestrator/payment client kullanımı nil kontrolü ile yapılır.

---

### 4.3 Proto varsa (payment.proto generate edildiyse)

```go
paymentpb "github.com/Chimera-State/GigaScale/api/proto/payment/v1"

paymentClient := paymentpb.NewPaymentServiceClient(paymentConn)
```

---

### 4.4 .env ve .env.example güncellemesi

**Eklenmesi gereken:**
```
PAYMENT_ADDR=localhost:50052
```

---

## 5. GIGA-75.3: Server Struct Güncellemesi

**Dosya:** `internal/gateway/server.go`

**Hedef:** Server'a `paymentClient` ve `orchestrator` alanları eklenip constructor güncellensin.

---

### 5.1 Mevcut Server

```go
type Server struct {
    reserveClient pb.ReservationServiceClient
    validator     *validator.Validate
    limiter       RateLimiter
}
```

---

### 5.2 Yeni Server (interface kullanımı)

Payment client tipi proto'ya bağlı. Proto yoksa `interface{}` veya ayrı bir interface kullanılabilir.

**Seçenek A — Proto yokken (sadece orchestrator):**

```go
import "github.com/Chimera-State/GigaScale/internal/gateway/orchestrator"

type Server struct {
    reserveClient pb.ReservationServiceClient
    orchestrator  orchestrator.Orchestrator
    validator     *validator.Validate
    limiter       RateLimiter
}

func NewServer(client pb.ReservationServiceClient, orch orchestrator.Orchestrator, limiter RateLimiter, v *validator.Validate) *Server {
    return &Server{
        reserveClient: client,
        orchestrator:  orch,
        validator:     v,
        limiter:       limiter,
    }
}
```

**Seçenek B — paymentClient da eklenecekse (proto hazır olduğunda):**

```go
type Server struct {
    reserveClient pb.ReservationServiceClient
    paymentClient paymentpb.PaymentServiceClient // proto'dan
    orchestrator  orchestrator.Orchestrator
    validator     *validator.Validate
    limiter       RateLimiter
}
```

---

### 5.3 main.go güncellemesi

```go
// Orchestrator oluştur (stub)
orch := orchestrator.NewStubOrchestrator()

srv := gateway.NewServer(client, orch, limiter, v)
```

---

### 5.4 Handler'da orchestrator kullanımı (Sprint 4)

Şimdilik handler `reserveClient` kullanmaya devam eder. Sprint 4'te `orchestrator.ReserveWithPayment` çağrılacak.

---

## 6. Küçük Düzeltmeler

### 6.1 main.go — reddisAddr typo

**Mevcut:**
```go
reddisAddr := getEnv("REDIS_ADDR", "localhost:6379")
rdb := redis.NewClient(&redis.Options{
    Addr: reddisAddr,
})
```

**Düzeltme:**
```go
redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
rdb := redis.NewClient(&redis.Options{
    Addr: redisAddr,
})
```

---

## Özet Checklist

| # | Görev | Dosya | Durum |
|---|-------|-------|-------|
| 1 | [REQUEST] + [RATE_LIMIT_BLOCK] log | middleware.go | ☐ |
| 2 | responseRecorder + duration/status log | middleware.go | ☐ |
| 3 | orchestrator/interface.go | Yeni dosya | ☐ |
| 4 | orchestrator/stub.go (opsiyonel) | Yeni dosya | ☐ |
| 5 | PAYMENT_ADDR + paymentConn | main.go | ☐ |
| 6 | Server struct + orchestrator | server.go | ☐ |
| 7 | main.go NewServer çağrısı | main.go | ☐ |
| 8 | reddisAddr → redisAddr | main.go | ☐ |

---

## Sıra Önerisi

1. middleware.go (72.1 + 72.2)
2. orchestrator/interface.go
3. orchestrator/stub.go
4. server.go
5. main.go (orchestrator, PAYMENT_ADDR, typo)
