-- name: CreateTrainingCatalogItem :one
INSERT INTO training_catalog_items (
    title,
    description,
    category,
    estimated_duration_minutes,
    created_by_employee_id
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListTrainingCatalogItemsPaginated :many
SELECT tci.*,
       COUNT(*) OVER() AS total_count
FROM training_catalog_items tci
WHERE (
    sqlc.narg('search')::text IS NULL
    OR tci.title ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(tci.description, '') ILIKE '%' || sqlc.narg('search')::text || '%'
    OR COALESCE(tci.category, '') ILIKE '%' || sqlc.narg('search')::text || '%'
)
AND (
    sqlc.narg('is_active')::boolean IS NULL
    OR tci.is_active = sqlc.narg('is_active')::boolean
)
ORDER BY tci.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
