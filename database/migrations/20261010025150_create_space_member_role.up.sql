CREATE TABLE space_member_role (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    space_member_id BIGINT NOT NULL REFERENCES space_members (id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT
);

CREATE UNIQUE INDEX uq_space_member_role_space_member_role ON space_member_role (space_id, space_member_id, role_id);

CREATE INDEX idx_space_member_role_space_id_role_id ON space_member_role (space_id, role_id);
