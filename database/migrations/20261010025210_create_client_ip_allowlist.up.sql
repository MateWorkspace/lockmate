CREATE TABLE client_ip_allowlist (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES clients (id),
    cidr TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_client_ip_allowlist_client_id_cidr ON client_ip_allowlist (client_id, cidr)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_client_ip_allowlist_client_id ON client_ip_allowlist (client_id);

CREATE INDEX idx_client_ip_allowlist_deleted_at ON client_ip_allowlist (deleted_at);
