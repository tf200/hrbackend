-- name: AddEmployeeContractDetails :one
UPDATE employee_profile
SET
    contract_hours = COALESCE(sqlc.narg('contract_hours'), contract_hours),
    contract_start_date = COALESCE(sqlc.narg('contract_start_date'), contract_start_date),
    contract_end_date = COALESCE(sqlc.narg('contract_end_date'), contract_end_date),
    contract_type = COALESCE(sqlc.narg('contract_type'), contract_type),
    contract_rate = COALESCE(sqlc.narg('contract_rate'), contract_rate),
    irregular_hours_profile = COALESCE(sqlc.narg('irregular_hours_profile'), irregular_hours_profile)
WHERE id = $1
RETURNING *;

-- name: GetEmployeeContractDetails :one
SELECT
    contract_hours,
    contract_start_date,
    contract_end_date,
    contract_type,
    contract_rate,
    irregular_hours_profile
FROM employee_profile
WHERE id = $1;

-- name: UpdateEmployeeIsSubcontractor :one
UPDATE employee_profile
SET
    contract_type = $2
WHERE id = $1
RETURNING *;

-- name: CountEmployeeContractChanges :one
SELECT COUNT(*)::bigint
FROM employee_contract_changes
WHERE employee_id = $1;

-- name: GetEmployeeContractSnapshotForContractChange :one
SELECT
    contract_hours,
    contract_start_date,
    contract_end_date,
    contract_type,
    contract_rate,
    irregular_hours_profile
FROM employee_profile
WHERE id = $1;

-- name: CreateEmployeeContractChange :one
INSERT INTO employee_contract_changes (
    employee_id,
    effective_from,
    contract_hours,
    contract_type,
    contract_rate,
    irregular_hours_profile,
    contract_end_date,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('effective_from'),
    sqlc.arg('contract_hours'),
    sqlc.arg('contract_type'),
    sqlc.narg('contract_rate'),
    sqlc.arg('irregular_hours_profile'),
    sqlc.narg('contract_end_date'),
    sqlc.arg('created_by_employee_id')
)
RETURNING *;

-- name: ListEmployeeContractChanges :many
SELECT
    c.id,
    c.employee_id,
    c.effective_from,
    (
        LEAD(c.effective_from) OVER (
            PARTITION BY c.employee_id
            ORDER BY c.effective_from
        ) - INTERVAL '1 day'
    )::date AS effective_to,
    c.contract_hours,
    c.contract_type,
    c.contract_rate,
    c.irregular_hours_profile,
    c.contract_end_date,
    c.created_by_employee_id,
    c.created_at,
    c.updated_at
FROM employee_contract_changes c
WHERE c.employee_id = $1
ORDER BY c.effective_from DESC, c.created_at DESC;

-- name: SyncEmployeeProfileContractFromLatestChange :one
WITH latest AS (
    SELECT
        contract_hours,
        contract_type,
        contract_rate,
        irregular_hours_profile,
        contract_end_date,
        effective_from
    FROM employee_contract_changes
    WHERE employee_id = $1
    ORDER BY effective_from DESC
    LIMIT 1
)
UPDATE employee_profile ep
SET
    contract_hours = latest.contract_hours,
    contract_type = latest.contract_type,
    contract_rate = latest.contract_rate,
    irregular_hours_profile = latest.irregular_hours_profile,
    contract_end_date = latest.contract_end_date,
    contract_start_date = latest.effective_from
FROM latest
WHERE ep.id = $1
RETURNING ep.*;
