-- name: AddEmployeeContractDetails :one
UPDATE employee_profile
SET
    contract_hours = COALESCE(sqlc.narg('contract_hours'), contract_hours),
    contract_start_date = COALESCE(sqlc.narg('contract_start_date'), contract_start_date),
    contract_end_date = COALESCE(sqlc.narg('contract_end_date'), contract_end_date),
    contract_type = COALESCE(sqlc.narg('contract_type'), contract_type),
    contract_rate = COALESCE(sqlc.narg('contract_rate'), contract_rate)
WHERE id = $1
RETURNING *;

-- name: GetEmployeeContractDetails :one
SELECT
    contract_hours,
    contract_start_date,
    contract_end_date,
    contract_type,
    contract_rate
FROM employee_profile
WHERE id = $1;

-- name: UpdateEmployeeIsSubcontractor :one
UPDATE employee_profile
SET
    contract_type = $2
WHERE id = $1
RETURNING *;
