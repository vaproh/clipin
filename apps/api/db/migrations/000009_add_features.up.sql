-- Users: add avatar_url and bio
ALTER TABLE users ADD COLUMN avatar_url TEXT;
ALTER TABLE users ADD COLUMN bio TEXT;

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,           -- 'submission_approved' | 'submission_rejected' | 'campaign_update' | 'payout_completed' | 'system'
    title TEXT NOT NULL,
    body TEXT,
    link TEXT,                    -- optional deep link
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications (user_id, is_read) WHERE is_read = false;

-- Campaign templates
CREATE TABLE campaign_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    platform TEXT NOT NULL DEFAULT 'youtube',
    cpm_rate INTEGER NOT NULL,
    total_budget INTEGER NOT NULL,
    max_clips_per_clipper INTEGER DEFAULT 3,
    min_views_per_clip INTEGER DEFAULT 1000,
    auto_approve_hours INTEGER DEFAULT 48,
    description_template TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
