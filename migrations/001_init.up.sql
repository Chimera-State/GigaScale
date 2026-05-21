-- 1. Trips (Seferler) Tablosu
-- Rezervasyon yapılabilmesi için önce bir seferin var olması gerekir.
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(12, 2) DEFAULT 100.00,
    departure_time TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + INTERVAL '1 day')
);

-- 2. Reservations (Rezervasyonlar) Tablosu
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, 
    idempotency_key VARCHAR(255) NOT NULL,
    trip_id UUID NOT NULL, -- Trips tablosuna bağlı
    seat_id VARCHAR(50) NOT NULL,
    payment_id UUID,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_status CHECK (status IN ('confirmed', 'cancelled', 'pending'))
);

-- Idempotency Index 
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_idempotency 
ON reservations(user_id, idempotency_key);

-- 4. doublebooking
CREATE UNIQUE INDEX IF NOT EXISTS idx_reservations_trip_seat_confirmed 
ON reservations(trip_id, seat_id) 
WHERE status = 'confirmed';

-- 5. K6 Test Seed data
INSERT INTO trips (id, name, price) 
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'GigaScale Remote Express', 100.00)
ON CONFLICT (id) DO NOTHING;
