CREATE TABLE role_permission (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces (id),
    role_id BIGINT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT,
    CONSTRAINT uq_role_permission_space_role_permission UNIQUE (space_id, role_id, permission_id)
);

CREATE INDEX idx_role_permission_role_id ON role_permission (role_id);

CREATE INDEX idx_role_permission_permission_id ON role_permission (permission_id);
