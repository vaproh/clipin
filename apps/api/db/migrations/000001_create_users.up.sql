CREATE TABLE users (
    id TEXT PRIMARY KEY,                    -- Clerk user ID (e.g. "user_2abc123")
    email TEXT NOT NULL,
    display_name TEXT,
    role TEXT NOT NULL DEFAULT 'clipper',   -- 'clipper' | 'owner' | 'admin'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role);
