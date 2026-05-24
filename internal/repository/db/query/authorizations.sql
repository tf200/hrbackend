-- name: ListAuthorizations :many
SELECT * FROM authorizations WHERE is_active = TRUE ORDER BY name;

-- name: CreateAuthorization :one
INSERT INTO authorizations (
    name,
    description,
    category,
    requires_expiry
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateAuthorization :one
UPDATE authorizations
SET
    name = $2,
    description = $3,
    category = $4,
    requires_expiry = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
