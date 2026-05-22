-- name: AddEmployeeContractDetails :one
INSERT INTO employee_contracts (
    employee_id,
    job_title,
    department_id,
    location_id,
    contract_type,
    contract_hours_type,
    start_date,
    contract_end_date,
    hours_per_week,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('job_title'),
    sqlc.arg('department_id'),
    sqlc.arg('location_id'),
    sqlc.arg('contract_type'),
    sqlc.arg('contract_hours_type'),
    sqlc.arg('start_date'),
    sqlc.narg('contract_end_date'),
    sqlc.narg('hours_per_week'),
    sqlc.narg('created_by_employee_id')
)
RETURNING *;

-- name: GetEmployeeContractDetails :one
SELECT
    hours_per_week AS contract_hours,
    start_date AS contract_start_date,
    contract_end_date,
    contract_type,
    NULL::numeric AS contract_rate,
    NULL::text AS irregular_hours_profile
FROM employee_contracts
WHERE employee_id = $1
ORDER BY start_date DESC, created_at DESC
LIMIT 1;

-- name: UpdateEmployeeIsSubcontractor :one
UPDATE employee_contracts ec
SET contract_type = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = (
    SELECT c.id
    FROM employee_contracts c
    WHERE c.employee_id = $1
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
)
RETURNING *;

-- name: CountEmployeeContractChanges :one
SELECT COUNT(*)::bigint
FROM employee_contracts
WHERE employee_id = $1;

-- name: GetEmployeeContractSnapshotForContractChange :one
SELECT
    hours_per_week AS contract_hours,
    start_date AS contract_start_date,
    contract_end_date,
    contract_type,
    NULL::numeric AS contract_rate,
    NULL::text AS irregular_hours_profile
FROM employee_contracts
WHERE employee_id = $1
ORDER BY start_date DESC, created_at DESC
LIMIT 1;

-- name: CreateEmployeeContractChange :one
INSERT INTO employee_contracts (
    employee_id,
    job_title,
    department_id,
    location_id,
    contract_type,
    contract_hours_type,
    start_date,
    contract_end_date,
    hours_per_week,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('job_title'),
    sqlc.arg('department_id'),
    sqlc.arg('location_id'),
    sqlc.arg('contract_type'),
    sqlc.arg('contract_hours_type'),
    sqlc.arg('effective_from'),
    sqlc.narg('contract_end_date'),
    sqlc.narg('contract_hours'),
    sqlc.narg('created_by_employee_id')
)
RETURNING *;

-- name: ListEmployeeContractChanges :many
SELECT
    c.id,
    c.employee_id,
    c.start_date AS effective_from,
    (
        LEAD(c.start_date) OVER (
            PARTITION BY c.employee_id
            ORDER BY c.start_date
        ) - INTERVAL '1 day'
    )::date AS effective_to,
    c.hours_per_week AS contract_hours,
    c.contract_type,
    NULL::numeric AS contract_rate,
    NULL::text AS irregular_hours_profile,
    c.contract_end_date,
    c.created_by_employee_id,
    c.created_at,
    c.updated_at
FROM employee_contracts c
WHERE c.employee_id = $1
ORDER BY c.start_date DESC, c.created_at DESC;

-- name: SyncEmployeeProfileContractFromLatestChange :one
SELECT *
FROM employee_profile
WHERE id = $1;
