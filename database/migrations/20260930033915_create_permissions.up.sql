CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    slug TEXT NOT NULL,
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

CREATE INDEX idx_permissions_space_id ON permissions (space_id);

CREATE UNIQUE INDEX uq_permissions_space_id_slug ON permissions (space_id, slug)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_permissions_name_trgm ON permissions USING GIN (name gin_trgm_ops);

CREATE INDEX idx_permissions_deleted_at ON permissions (deleted_at);
