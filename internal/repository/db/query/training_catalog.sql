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
