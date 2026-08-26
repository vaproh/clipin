-- UPI details for clippers
ALTER TABLE users ADD COLUMN upi_id TEXT;

-- Payout requests
CREATE TABLE payout_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clipper_id TEXT NOT NULL REFERENCES users(id),
    amount INTEGER NOT NULL,           -- paise, must be >= 50000 (₹500)
    upi_id TEXT NOT NULL,              -- UPI ID for this payout
    status TEXT NOT NULL DEFAULT 'pending',  -- 'pending'|'processing'|'completed'|'failed'
    provider_ref TEXT,                 -- Razorpay transfer ID
    failure_reason TEXT,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payout_requests_clipper ON payout_requests (clipper_id);
CREATE INDEX idx_payout_requests_status ON payout_requests (status);
