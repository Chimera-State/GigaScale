CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, 
    idempotency_key VARCHAR(255) NOT NULL,

    trip_id UUID NOT NULL,
    seat_id VARCHAR(50) NOT NULL,

    payment_id UUID,
    amount DECIMAL(12, 2) NOT NULL CHECK (amount >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT chk_status CHECK (status IN ('confirmed', 'cancelled', 'pending'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_idempotency 
ON reservations(user_id, idempotency_key);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reservations_trip_seat_confirmed 
ON reservations(trip_id, seat_id) 
WHERE status = 'confirmed';
