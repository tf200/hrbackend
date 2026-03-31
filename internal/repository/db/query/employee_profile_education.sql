-- name: AddEducationToEmployeeProfile :one
INSERT INTO employee_education (
    employee_id,
    institution_name,
    degree,
    field_of_study,
    start_date,
    end_date
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListEducations :many
SELECT * FROM employee_education WHERE employee_id = $1;

-- name: UpdateEmployeeEducation :one
UPDATE employee_education
SET
    institution_name = COALESCE(sqlc.narg('institution_name'), institution_name),
    degree = COALESCE(sqlc.narg('degree'), degree),
    field_of_study = COALESCE(sqlc.narg('field_of_study'), field_of_study),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date)
WHERE id = $1
RETURNING *;

-- name: DeleteEmployeeEducation :one
DELETE FROM employee_education WHERE id = $1 RETURNING *;