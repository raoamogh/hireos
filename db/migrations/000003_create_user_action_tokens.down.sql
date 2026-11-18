-- Roll back HIREOS email verification and password reset token storage.
-- Migration: 000003

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM user_action_tokens
    ) THEN
        RAISE EXCEPTION
            'Cannot roll back migration 000003: user action token records exist';
    END IF;
END;
$$;

DROP TABLE user_action_tokens;
