-- name: GetOrganizationByID :one
SELECT
    id,
    name,
    status,
    created_at,
    updated_at
FROM organizations
WHERE id = $1;
