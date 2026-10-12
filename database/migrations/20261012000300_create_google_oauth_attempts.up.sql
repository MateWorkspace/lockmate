CREATE TABLE google_oauth_attempts (
    state_hash TEXT NOT NULL PRIMARY KEY,
    binding_hash TEXT NOT NULL,
    nonce_hash TEXT NOT NULL,
    verifier TEXT NOT NULL,
    client_id BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    space_id BIGINT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    client_state TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_google_oauth_attempts_expires_at ON google_oauth_attempts (expires_at);
