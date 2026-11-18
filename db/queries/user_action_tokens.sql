
-- Lock the canonical user row while issuing an action token.
-- The caller must execute this inside a transaction.
-- name: LockUserForActionToken :one
SELECT id
FROM users
WHERE id = $1
FOR UPDATE;


-- Invalidate all previously outstanding tokens for this user and purpose.
-- This also invalidates expired tokens that have not been consumed.
-- The caller must execute this inside the same transaction as token creation.
-- name: InvalidateOutstandingUserActionTokens :exec
UPDATE user_action_tokens
SET invalidated_at = statement_timestamp()
WHERE user_id = $1
  AND purpose = $2
  AND consumed_at IS NULL
  AND invalidated_at IS NULL;


-- Persist only the token hash, never the raw token.
-- name: CreateUserActionToken :one
INSERT INTO user_action_tokens (
    user_id,
    purpose,
    token_hash,
    expires_at
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING
    id,
    user_id,
    purpose,
    token_hash,
    expires_at,
    consumed_at,
    invalidated_at,
    created_at;


-- Consume a token only if it is outstanding and unexpired.
-- The caller must execute this in the same transaction as the
-- business operation authorized by the token.
-- name: ConsumeUserActionToken :one
UPDATE user_action_tokens
SET consumed_at = statement_timestamp()
WHERE token_hash = $1
  AND purpose = $2
  AND consumed_at IS NULL
  AND invalidated_at IS NULL
  AND expires_at > statement_timestamp()
RETURNING
    id,
    user_id,
    purpose;
