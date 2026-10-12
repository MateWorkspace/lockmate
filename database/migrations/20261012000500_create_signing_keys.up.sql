CREATE TABLE signing_keys (
    id BIGSERIAL PRIMARY KEY,
    kid TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    private_key_cipher BYTEA NOT NULL,
    private_key_nonce BYTEA NOT NULL,
    key_fingerprint TEXT NOT NULL,
    public_jwk JSONB NOT NULL,
    status TEXT NOT NULL,
    not_before TIMESTAMPTZ NOT NULL,
    retiring_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT,
    CONSTRAINT ck_signing_keys_status CHECK (status IN ('active', 'retiring'))
);

CREATE UNIQUE INDEX uq_signing_keys_kid ON signing_keys (kid);

CREATE UNIQUE INDEX uq_signing_keys_active ON signing_keys (status)
WHERE
    status = 'active'
    AND deleted_at IS NULL;

CREATE INDEX idx_signing_keys_deleted_at ON signing_keys (deleted_at);
