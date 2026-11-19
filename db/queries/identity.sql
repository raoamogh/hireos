
-- Update the password credential after successful password-reset token consumption.
-- The caller must execute this inside the same transaction as token consumption.
-- name: UpdateUserPasswordCredential :execrows
UPDATE user_password_credentials
SET
    password_hash = $2,
    password_changed_at = statement_timestamp(),
    updated_at = statement_timestamp()
WHERE user_id = $1;


-- Mark a user's email as verified.
-- The caller must execute this inside the same transaction as token consumption.
-- name: MarkUserEmailVerified :execrows
UPDATE users
SET
    email_verified_at = COALESCE(
        email_verified_at,
        statement_timestamp()
    ),
    updated_at = statement_timestamp()
WHERE id = $1;
