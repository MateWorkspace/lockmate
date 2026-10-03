CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    password_hash TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    avatar_path TEXT,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE INDEX idx_users_name_trgm ON users USING GIN (name gin_trgm_ops);

CREATE UNIQUE INDEX uq_users_username ON users (username)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_users_username_trgm ON users USING GIN (username gin_trgm_ops);

CREATE UNIQUE INDEX uq_users_email ON users (email)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_users_email_trgm ON users USING GIN (email gin_trgm_ops);

CREATE UNIQUE INDEX uq_users_phone ON users (phone)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_users_phone_trgm ON users USING GIN (phone gin_trgm_ops);

CREATE INDEX idx_users_is_email_verified ON users (is_email_verified);

CREATE INDEX idx_users_is_phone_verified ON users (is_phone_verified);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);
