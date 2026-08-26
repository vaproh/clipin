CREATE TABLE submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    clipper_id TEXT NOT NULL REFERENCES users(id),
    post_url TEXT NOT NULL,                 -- URL of the published clip
    platform TEXT NOT NULL,
    platform_post_id TEXT,                  -- video/post ID extracted from URL
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending' | 'approved' | 'rejected' | 'auto_approved' | 'disputed'
    rejection_reason TEXT,
    approved_at TIMESTAMPTZ,
    auto_approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_submissions_campaign_clipper_url ON submissions (campaign_id, clipper_id, post_url);
CREATE INDEX idx_submissions_clipper ON submissions (clipper_id);
CREATE INDEX idx_submissions_status ON submissions (status);
CREATE INDEX idx_submissions_campaign ON submissions (campaign_id);
