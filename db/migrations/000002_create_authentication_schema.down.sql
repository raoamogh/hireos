
-- Roll back HIREOS authentication credentials and session foundation.
-- Migration: 000002

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM sessions
    ) THEN
        RAISE EXCEPTION
            'Cannot roll back migration 000002: active or historical session records exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM user_auth_identities
    ) THEN
        RAISE EXCEPTION
            'Cannot roll back migration 000002: external authentication identities exist';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM users u
        LEFT JOIN user_password_credentials c
            ON c.user_id = u.id
        WHERE c.user_id IS NULL
    ) THEN
        RAISE EXCEPTION
            'Cannot roll back migration 000002: one or more users have no password credential';
    END IF;
END;
$$;

ALTER TABLE users
    ADD COLUMN password_hash TEXT;

UPDATE users u
SET password_hash = c.password_hash
FROM user_password_credentials c
WHERE c.user_id = u.id;

ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL;

DROP TABLE session_token_history;
DROP TABLE sessions;
DROP TABLE user_auth_identities;
DROP TABLE user_password_credentials;
