CREATE TABLE client_origins (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES clients (id),
    origin TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_client_origins_client_id_origin ON client_origins (client_id, origin)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_client_origins_client_id ON client_origins (client_id);

CREATE INDEX idx_client_origins_origin ON client_origins (origin)
WHERE
    deleted_at IS NULL AND is_active = TRUE;

CREATE INDEX idx_client_origins_deleted_at ON client_origins (deleted_at);
