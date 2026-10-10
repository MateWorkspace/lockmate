CREATE TABLE clients (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    secret_key_hash TEXT NOT NULL,
    secret_key_redacted TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_clients_slug ON clients (slug);

CREATE INDEX idx_clients_space_id ON clients (space_id);

CREATE INDEX idx_clients_name_trgm ON clients USING GIN (name gin_trgm_ops)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_clients_deleted_at ON clients (deleted_at);
