CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_roles_name ON roles (name)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_roles_name_trgm ON roles USING GIN (name gin_trgm_ops);

CREATE UNIQUE INDEX uq_roles_is_default_true ON roles (is_default)
WHERE
    is_default = TRUE;

CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);
