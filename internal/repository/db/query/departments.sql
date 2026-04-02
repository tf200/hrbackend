-- name: CreateDepartment :one
INSERT INTO departments (
    name,
    description,
    department_head_employee_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetDepartment :one
SELECT *
FROM departments
WHERE id = $1;

-- name: ListDepartmentsPaginated :many
SELECT d.*,
       COUNT(*) OVER() AS total_count
FROM departments d
WHERE ($3::text IS NULL OR d.name ILIKE '%' || $3 || '%')
ORDER BY d.name
LIMIT $1 OFFSET $2;

-- name: UpdateDepartment :one
UPDATE departments
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    department_head_employee_id = COALESCE(sqlc.narg('department_head_employee_id'), department_head_employee_id),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDepartment :one
DELETE FROM departments
WHERE id = $1
RETURNING *;
