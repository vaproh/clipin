CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id TEXT NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT,
    brief_url TEXT,                         -- link to detailed brief
    platform TEXT NOT NULL,                 -- 'youtube' | 'instagram' | 'multi'
    status TEXT NOT NULL DEFAULT 'draft',   -- 'draft' | 'funded' | 'active' | 'paused' | 'completed' | 'cancelled'
    cpm_rate INTEGER NOT NULL,              -- paise per 1000 views (e.g. 5000 = ₹50/1k)
    total_budget INTEGER NOT NULL,          -- paise
    remaining_budget INTEGER NOT NULL,      -- paise
    platform_fee INTEGER NOT NULL DEFAULT 0, -- paise, 10% of deposit
    escrow_id TEXT,                         -- payment reference
    max_clips_per_campaign INTEGER,         -- optional cap
    max_clips_per_clipper INTEGER DEFAULT 3, -- default per clipper
    min_views_per_clip INTEGER DEFAULT 1000,
    auto_approve_hours INTEGER DEFAULT 48,  -- auto-approve if owner doesn't act
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaigns_owner ON campaigns (owner_id);
CREATE INDEX idx_campaigns_status ON campaigns (status);
CREATE INDEX idx_campaigns_platform ON campaigns (platform);
