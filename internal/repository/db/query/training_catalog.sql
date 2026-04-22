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

-- name: AssignTrainingToEmployee :one
WITH target_employee AS (
    SELECT ep.id
    FROM employee_profile ep
    WHERE ep.id = sqlc.arg('employee_id')
    LIMIT 1
), target_training AS (
    SELECT tci.id
    FROM training_catalog_items tci
    WHERE tci.id = sqlc.arg('training_id')
      AND is_active = TRUE
    LIMIT 1
)
INSERT INTO employee_training_assignments (
    employee_id,
    training_id,
    assigned_by_employee_id,
    due_at
)
SELECT
    te.id,
    tt.id,
    sqlc.narg('assigned_by_employee_id'),
    sqlc.arg('due_at')
FROM target_employee te
CROSS JOIN target_training tt
RETURNING *;

-- name: GetTrainingAssignmentByID :one
SELECT *
FROM employee_training_assignments
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: CancelTrainingAssignment :one
UPDATE employee_training_assignments
SET
    status = 'cancelled',
    cancelled_at = NOW(),
    cancellation_reason = sqlc.narg('cancellation_reason'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND status IN ('assigned', 'in_progress')
RETURNING *;

-- name: ListTrainingAssignmentsPaginated :many
SELECT
    eta.id AS assignment_id,
    eta.employee_id,
    ep.employee_number,
    ep.employment_number,
    ep.first_name,
    ep.last_name,
    ep.department_id,
    d.name AS department_name,
    eta.training_id,
    tci.title AS training_title,
    tci.category AS training_category,
    eta.status::text AS status,
    eta.assigned_at,
    eta.due_at,
    eta.started_at,
    eta.completed_at,
    eta.assigned_by_employee_id,
    NULLIF(TRIM(CONCAT_WS(' ', assigner.first_name, assigner.last_name)), '')::text AS assigned_by_name,
    CASE WHEN (
        eta.due_at IS NOT NULL
        AND eta.due_at < NOW()
        AND eta.status NOT IN ('completed', 'cancelled')
    ) THEN TRUE ELSE FALSE END AS is_overdue,
    COUNT(*) OVER() AS total_count
FROM employee_training_assignments eta
JOIN employee_profile ep ON ep.id = eta.employee_id
JOIN training_catalog_items tci ON tci.id = eta.training_id
LEFT JOIN departments d ON d.id = ep.department_id
LEFT JOIN employee_profile assigner ON assigner.id = eta.assigned_by_employee_id
WHERE (
    sqlc.narg('employee_search')::text IS NULL
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR COALESCE(ep.employee_number, '') ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR COALESCE(ep.employment_number, '') ILIKE '%' || sqlc.narg('employee_search')::text || '%'
)
AND (
    sqlc.narg('department_id')::uuid IS NULL
    OR ep.department_id = sqlc.narg('department_id')::uuid
)
AND (
    sqlc.narg('training_id')::uuid IS NULL
    OR eta.training_id = sqlc.narg('training_id')::uuid
)
AND (
    (sqlc.narg('status_filter')::text IS NULL AND eta.status <> 'cancelled')
    OR eta.status::text = sqlc.narg('status_filter')::text
)
ORDER BY eta.assigned_at DESC, eta.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

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
