CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key TEXT NOT NULL UNIQUE,
    entry_type TEXT NOT NULL,      -- 'platform_fee' | 'earning' | 'refund' | 'escrow_lock' | 'escrow_release'
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    submission_id UUID,            -- NULL for campaign-level entries (fees, refunds)
    clipper_id TEXT,               -- NULL for non-earning entries
    amount INTEGER NOT NULL,       -- paise, positive = credit, negative = debit
    description TEXT,
    metadata JSONB,                -- flexible data (e.g., { cpm_rate, views, eligible_views })
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_campaign ON ledger_entries (campaign_id);
CREATE INDEX idx_ledger_entries_clipper ON ledger_entries (clipper_id);
CREATE INDEX idx_ledger_entries_idempotency ON ledger_entries (idempotency_key);
CREATE INDEX idx_ledger_entries_type ON ledger_entries (entry_type);
