
-- HIREOS authentication credentials and session foundation.
-- Migration: 000002

-- Store password credentials separately from the global user record.
CREATE TABLE user_password_credentials (
    user_id UUID PRIMARY KEY
        REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL
        CHECK (length(btrim(password_hash)) > 0),
    password_changed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Preserve existing password hashes before removing the legacy column.
-- Historical password-change timestamps are unknown, so leave them NULL.
INSERT INTO user_password_credentials (
    user_id,
    password_hash
)
SELECT
    id,
    password_hash
FROM users;

ALTER TABLE users
    DROP COLUMN password_hash;

-- Link external OIDC identities to canonical HIREOS users.
CREATE TABLE user_auth_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL
        CHECK (provider IN ('google', 'linkedin')),
    issuer TEXT NOT NULL
        CHECK (length(btrim(issuer)) > 0),
    subject TEXT NOT NULL
        CHECK (length(btrim(subject)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_authenticated_at TIMESTAMPTZ,

    CONSTRAINT user_auth_identities_issuer_subject_unique
        UNIQUE (issuer, subject)
);

CREATE INDEX user_auth_identities_user_id_idx
    ON user_auth_identities(user_id);

-- Store only the hash of the current random session secret.
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE
        CHECK (octet_length(token_hash) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,

    CONSTRAINT sessions_expiry_after_creation
        CHECK (expires_at > created_at),

    CONSTRAINT sessions_revocation_after_creation
        CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX sessions_user_id_idx
    ON sessions(user_id);

CREATE INDEX sessions_active_expiry_idx
    ON sessions(expires_at)
    WHERE revoked_at IS NULL;

-- Retain hashes of previously rotated session secrets.
-- Reuse of a historical token can then be detected.
CREATE TABLE session_token_history (
    token_hash BYTEA PRIMARY KEY
        CHECK (octet_length(token_hash) = 32),
    session_id UUID NOT NULL
        REFERENCES sessions(id) ON DELETE CASCADE,
    rotated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX session_token_history_session_id_idx
    ON session_token_history(session_id);
