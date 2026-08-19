CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,

    access_token_expired_at TIMESTAMPTZ NOT NULL,
    refresh_token_expired_at TIMESTAMPTZ NOT NULL,

    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMPTZ,

    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,

    ip_address INET,
    user_agent TEXT,

    device_id VARCHAR(255),
    device_name VARCHAR(255),
    platform VARCHAR(50),

    last_used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);