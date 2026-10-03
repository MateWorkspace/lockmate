CREATE TABLE member_role (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    member_id BIGINT NOT NULL REFERENCES space_members (id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT,
    CONSTRAINT uq_member_role_space_member_role UNIQUE (space_id, member_id, role_id)
);

CREATE INDEX idx_member_role_space_id_role_id ON member_role (space_id, role_id);
