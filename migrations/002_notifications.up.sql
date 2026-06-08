-- Notifications audit log table.
-- Tracks every outbound notification attempt for idempotency and observability.
CREATE TABLE IF NOT EXISTS notifications (
    id              BIGSERIAL PRIMARY KEY,
    idempotency_key VARCHAR(255)              NOT NULL,
    user_id         UUID                      NOT NULL,
    trip_id         UUID                      NOT NULL,
    seat_id         VARCHAR(50)               NOT NULL,
    payment_id      VARCHAR(255)              NOT NULL,
    channel         VARCHAR(50)               NOT NULL, -- 'webhook' | 'email' | 'sms'
    status          VARCHAR(20)               NOT NULL, -- 'sent' | 'failed'
    sent_at         TIMESTAMP WITH TIME ZONE  NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- One successful delivery per idempotency_key+channel is enough.
    CONSTRAINT uq_notifications_key_channel UNIQUE (idempotency_key, channel),
    CONSTRAINT chk_notifications_status CHECK (status IN ('sent', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_sent_at  ON notifications (sent_at DESC);
