CREATE TABLE social_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform TEXT NOT NULL,                 -- 'youtube' | 'instagram' | 'tiktok'
    platform_user_id TEXT NOT NULL,         -- channel ID, username etc.
    platform_username TEXT,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_social_accounts_platform_user ON social_accounts (platform, platform_user_id);
CREATE INDEX idx_social_accounts_user_id ON social_accounts (user_id);
