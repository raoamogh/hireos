-- HIREOS email verification and password reset token storage.
-- Migration: 000003

CREATE TABLE user_action_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    purpose TEXT NOT NULL
        CHECK (
            purpose IN (
                'email_verification',
                'password_reset'
            )
        ),

    token_hash BYTEA NOT NULL UNIQUE
        CHECK (octet_length(token_hash) = 32),

    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    invalidated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT user_action_tokens_expiry_after_creation
        CHECK (expires_at > created_at),

    CONSTRAINT user_action_tokens_consumed_after_creation
        CHECK (
            consumed_at IS NULL
            OR consumed_at >= created_at
        ),

    CONSTRAINT user_action_tokens_invalidated_after_creation
        CHECK (
            invalidated_at IS NULL
            OR invalidated_at >= created_at
        ),

    CONSTRAINT user_action_tokens_single_terminal_state
        CHECK (
            consumed_at IS NULL
            OR invalidated_at IS NULL
        )
);

CREATE INDEX user_action_tokens_user_id_idx
    ON user_action_tokens(user_id);

CREATE UNIQUE INDEX user_action_tokens_one_outstanding_idx
    ON user_action_tokens(user_id, purpose)
    WHERE consumed_at IS NULL
      AND invalidated_at IS NULL;

CREATE INDEX user_action_tokens_expiry_idx
    ON user_action_tokens(expires_at)
    WHERE consumed_at IS NULL
      AND invalidated_at IS NULL;
