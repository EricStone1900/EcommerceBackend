CREATE TABLE IF NOT EXISTS push_tokens (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(512) NOT NULL,
    platform     VARCHAR(20)  NOT NULL DEFAULT 'ios',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_push_tokens_user_device ON push_tokens (user_id, device_token);
CREATE INDEX IF NOT EXISTS idx_push_tokens_deleted_at ON push_tokens (deleted_at);
