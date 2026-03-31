-- name: AddEmployeeCertification :one
INSERT INTO certification (
    employee_id,
    name,
    issued_by,
    date_issued
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListEmployeeCertifications :many
SELECT * FROM certification WHERE employee_id = $1;

-- name: UpdateEmployeeCertification :one
UPDATE certification
SET
    name = COALESCE(sqlc.narg('name'), name),
    issued_by = COALESCE(sqlc.narg('issued_by'), issued_by),
    date_issued = COALESCE(sqlc.narg('date_issued'), date_issued)
WHERE id = $1
RETURNING *;

-- name: DeleteEmployeeCertification :one
DELETE FROM certification WHERE id = $1 RETURNING *;