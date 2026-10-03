CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_permissions_name ON permissions (name)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_permissions_name_trgm ON permissions USING GIN (name gin_trgm_ops);

CREATE INDEX idx_permissions_deleted_at ON permissions (deleted_at);
