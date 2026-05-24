-- name: AddEmployeeAuthorization :one
INSERT INTO employee_authorizations (
    employee_id,
    authorization_id,
    granted_date,
    expiry_date,
    is_active,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: AddEmployeeAuthorizationsBatch :copyfrom
INSERT INTO employee_authorizations (
    employee_id,
    authorization_id,
    granted_date,
    expiry_date,
    is_active,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: ListEmployeeAuthorizations :many
SELECT * FROM employee_authorizations WHERE employee_id = $1;

-- name: UpdateEmployeeAuthorization :one
UPDATE employee_authorizations
SET
    authorization_id = COALESCE(sqlc.narg('authorization_id'), authorization_id),
    granted_date = COALESCE(sqlc.narg('granted_date'), granted_date),
    expiry_date = COALESCE(sqlc.narg('expiry_date'), expiry_date),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    notes = COALESCE(sqlc.narg('notes'), notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteEmployeeAuthorization :one
DELETE FROM employee_authorizations WHERE id = $1 RETURNING *;
