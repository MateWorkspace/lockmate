CREATE TABLE google_auths (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id),
    subject TEXT NOT NULL,
    email TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_google_auths_user_id ON google_auths (user_id)
WHERE
    deleted_at IS NULL;

CREATE UNIQUE INDEX uq_google_auths_subject ON google_auths (subject)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_google_auths_email_trgm ON google_auths USING GIN (email gin_trgm_ops);

CREATE INDEX idx_google_auths_deleted_at ON google_auths (deleted_at);
