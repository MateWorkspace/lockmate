CREATE TABLE classic_auths (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id),
    username TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    password_hash TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_classic_auths_user_id ON classic_auths (user_id)
WHERE
    deleted_at IS NULL;

CREATE UNIQUE INDEX uq_classic_auths_username ON classic_auths (username)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_classic_auths_username_trgm ON classic_auths USING GIN (username gin_trgm_ops);

CREATE UNIQUE INDEX uq_classic_auths_email ON classic_auths (email)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_classic_auths_email_trgm ON classic_auths USING GIN (email gin_trgm_ops);

CREATE UNIQUE INDEX uq_classic_auths_phone ON classic_auths (phone)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_classic_auths_phone_trgm ON classic_auths USING GIN (phone gin_trgm_ops);

CREATE INDEX idx_classic_auths_is_email_verified ON classic_auths (is_email_verified);

CREATE INDEX idx_classic_auths_is_phone_verified ON classic_auths (is_phone_verified);

CREATE INDEX idx_classic_auths_deleted_at ON classic_auths (deleted_at);
