CREATE TABLE IF NOT EXISTS files (
    id             BIGSERIAL    PRIMARY KEY,
    owner_id       BIGINT       NOT NULL REFERENCES users(id),
    type           VARCHAR(20)  NOT NULL,
    original_name  VARCHAR(255) NOT NULL,
    url            VARCHAR(512) NOT NULL,
    size           BIGINT       NOT NULL DEFAULT 0,
    status         VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files (deleted_at);
CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files (owner_id);
