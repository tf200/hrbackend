-- name: AddEmployeeExperience :one
INSERT INTO employee_experience (
    employee_id,
    job_title,
    company_name,
    start_date,
    end_date,
    description
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListEmployeeExperience :many
SELECT * FROM employee_experience WHERE employee_id = $1;

-- name: UpdateEmployeeExperience :one
UPDATE employee_experience
SET
    job_title = COALESCE(sqlc.narg('job_title'), job_title),
    company_name = COALESCE(sqlc.narg('company_name'), company_name),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date),
    description = COALESCE(sqlc.narg('description'), description)
WHERE id = $1
RETURNING *;

-- name: DeleteEmployeeExperience :one
DELETE FROM employee_experience WHERE id = $1 RETURNING *;