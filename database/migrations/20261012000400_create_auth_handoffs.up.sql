CREATE TABLE auth_handoffs (
    code_hash TEXT NOT NULL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    space_id BIGINT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    space_member_id BIGINT NOT NULL REFERENCES space_members(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    is_new_user BOOLEAN NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_auth_handoffs_expires_at ON auth_handoffs (expires_at);
