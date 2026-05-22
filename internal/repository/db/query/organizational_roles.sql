-- name: ListOrganizationalRoles :many
SELECT id, name, description, is_active
FROM organizational_roles
WHERE (
    sqlc.narg('active_only')::bool IS NULL
    OR sqlc.narg('active_only')::bool = false
    OR is_active = true
)
AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR description ILIKE '%' || sqlc.narg('search')::text || '%'
)
ORDER BY name;
