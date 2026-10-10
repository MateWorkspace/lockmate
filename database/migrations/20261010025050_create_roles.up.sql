CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    slug TEXT NOT NULL,
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

CREATE INDEX idx_roles_space_id ON roles (space_id);

CREATE UNIQUE INDEX uq_roles_space_id_slug ON roles (space_id, slug)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_roles_name_trgm ON roles USING GIN (name gin_trgm_ops);

CREATE UNIQUE INDEX uq_roles_space_id_default ON roles (space_id)
WHERE
    is_default = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);
