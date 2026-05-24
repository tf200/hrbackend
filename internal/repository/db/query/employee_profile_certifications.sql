-- name: AddEmployeeQualification :one
INSERT INTO employee_qualifications (
    employee_id,
    qualification_id,
    achieved_on,
    expiration_date,
    certificate_number
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListEmployeeQualifications :many
SELECT * FROM employee_qualifications WHERE employee_id = $1;

-- name: UpdateEmployeeQualification :one
UPDATE employee_qualifications
SET
    qualification_id = COALESCE(sqlc.narg('qualification_id'), qualification_id),
    achieved_on = COALESCE(sqlc.narg('achieved_on'), achieved_on),
    expiration_date = COALESCE(sqlc.narg('expiration_date'), expiration_date),
    certificate_number = COALESCE(sqlc.narg('certificate_number'), certificate_number),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteEmployeeQualification :one
DELETE FROM employee_qualifications WHERE id = $1 RETURNING *;

-- name: ListQualificationTypes :many
SELECT * FROM qualifications WHERE is_active = TRUE ORDER BY name;

-- name: GetQualificationType :one
SELECT * FROM qualifications WHERE code = $1;

-- name: UpdateQualificationType :one
UPDATE qualifications
SET
    code = $2,
    name = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CreateQualificationType :one
INSERT INTO qualifications (
    code,
    name,
    app_context
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: AddEmployeeQualificationsBatch :copyfrom
INSERT INTO employee_qualifications (
    employee_id,
    qualification_id,
    achieved_on,
    expiration_date,
    certificate_number
) VALUES (
    $1, $2, $3, $4, $5
);

