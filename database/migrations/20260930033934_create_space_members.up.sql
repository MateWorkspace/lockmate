CREATE TABLE space_members (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    user_id BIGINT NOT NULL REFERENCES users (id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT
);

CREATE UNIQUE INDEX uq_space_members_space_id_user_id ON space_members (space_id, user_id)
WHERE
    deleted_at IS NULL;

CREATE INDEX idx_space_members_user_id ON space_members (user_id);

CREATE INDEX idx_space_members_deleted_at ON space_members (deleted_at);
