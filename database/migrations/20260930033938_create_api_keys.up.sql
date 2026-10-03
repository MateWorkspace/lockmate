CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    member_id BIGINT NOT NULL REFERENCES space_members (id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    hash TEXT NOT NULL,
    redacted TEXT NOT NULL,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE INDEX idx_api_keys_space_id_member_id ON api_keys (space_id, member_id);

CREATE INDEX idx_api_keys_name_trgm ON api_keys USING GIN (name gin_trgm_ops);

CREATE UNIQUE INDEX uq_api_keys_hash ON api_keys (hash)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_api_keys_redacted_trgm ON api_keys USING GIN (redacted gin_trgm_ops);

CREATE INDEX idx_api_keys_deleted_at ON api_keys (deleted_at);
