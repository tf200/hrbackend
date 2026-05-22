-- name: AddEmployeeContractDetails :one
INSERT INTO employee_contracts (
    employee_id,
    job_title,
    department_id,
    location_id,
    organizational_role_id,
    contract_type,
    contract_hours_type,
    start_date,
    contract_end_date,
    hours_per_week,
    min_hours_per_week,
    max_hours_per_week,
    roster_free_day,
    wage_tax_table,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('job_title'),
    sqlc.arg('department_id'),
    sqlc.arg('location_id'),
    sqlc.narg('organizational_role_id'),
    sqlc.arg('contract_type'),
    sqlc.arg('contract_hours_type'),
    sqlc.arg('start_date'),
    sqlc.narg('contract_end_date'),
    sqlc.narg('hours_per_week'),
    sqlc.narg('min_hours_per_week'),
    sqlc.narg('max_hours_per_week'),
    sqlc.narg('roster_free_day'),
    sqlc.narg('wage_tax_table'),
    sqlc.narg('created_by_employee_id')
)
RETURNING *;


