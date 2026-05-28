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
    sqlc.arg('roster_free_day'),
    sqlc.narg('wage_tax_table'),
    sqlc.narg('created_by_employee_id')
)
RETURNING *;

-- name: GetLatestEmployeeContractDetail :one
SELECT
    ec.id,
    ec.employee_id,
    ec.job_title,
    ec.department_id,
    d.name AS department_name,
    ec.location_id,
    concat_ws(' ', l.street, l.house_number, l.house_number_addition, l.postal_code, l.city) AS location_address,
    ec.organizational_role_id,
    org_role.name AS organizational_role_name,
    ec.contract_type,
    ec.contract_hours_type,
    ec.start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    ec.hours_per_week,
    ec.min_hours_per_week,
    ec.max_hours_per_week,
    ec.roster_free_day,
    ec.wage_tax_table,
    ec.created_at,
    ec.updated_at
FROM employee_contracts ec
JOIN departments d ON d.id = ec.department_id
JOIN location l ON l.id = ec.location_id
LEFT JOIN organizational_roles org_role ON org_role.id = ec.organizational_role_id
WHERE ec.employee_id = $1
ORDER BY ec.start_date DESC, ec.created_at DESC
LIMIT 1;

-- name: GetActiveEmployeeContract :one
SELECT *
FROM employee_contracts
WHERE employee_id = $1
  AND start_date <= CURRENT_DATE
  AND (effective_end_date IS NULL OR effective_end_date >= CURRENT_DATE)
  AND (contract_end_date IS NULL OR contract_end_date >= CURRENT_DATE)
ORDER BY start_date DESC, created_at DESC
LIMIT 1;

-- name: GetActiveEmployeeContractDetail :one
SELECT
    ec.id,
    ec.employee_id,
    ec.job_title,
    ec.department_id,
    d.name AS department_name,
    ec.location_id,
    concat_ws(' ', l.street, l.house_number, l.house_number_addition, l.postal_code, l.city) AS location_address,
    ec.organizational_role_id,
    org_role.name AS organizational_role_name,
    ec.contract_type,
    ec.contract_hours_type,
    ec.start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    ec.hours_per_week,
    ec.min_hours_per_week,
    ec.max_hours_per_week,
    ec.roster_free_day,
    ec.wage_tax_table,
    ec.created_at,
    ec.updated_at
FROM employee_contracts ec
JOIN departments d ON d.id = ec.department_id
JOIN location l ON l.id = ec.location_id
LEFT JOIN organizational_roles org_role ON org_role.id = ec.organizational_role_id
WHERE ec.employee_id = $1
  AND ec.start_date <= CURRENT_DATE
  AND (ec.effective_end_date IS NULL OR ec.effective_end_date >= CURRENT_DATE)
  AND (ec.contract_end_date IS NULL OR ec.contract_end_date >= CURRENT_DATE)
ORDER BY ec.start_date DESC, ec.created_at DESC
LIMIT 1;

-- name: GetEmployeeContractAtDate :one
SELECT *
FROM employee_contracts
WHERE employee_id = $1
  AND start_date <= sqlc.arg('target_date')::date
  AND (effective_end_date IS NULL OR effective_end_date >= sqlc.arg('target_date')::date)
  AND (contract_end_date IS NULL OR contract_end_date >= sqlc.arg('target_date')::date)
ORDER BY start_date DESC, created_at DESC
LIMIT 1;

-- name: ListEmployeeContracts :many
SELECT *
FROM employee_contracts
WHERE employee_id = $1
ORDER BY start_date ASC, created_at ASC;

-- name: ListEmployeeContractDetails :many
SELECT
    ec.id,
    ec.employee_id,
    ec.job_title,
    ec.department_id,
    d.name AS department_name,
    ec.location_id,
    concat_ws(' ', l.street, l.house_number, l.house_number_addition, l.postal_code, l.city) AS location_address,
    ec.organizational_role_id,
    org_role.name AS organizational_role_name,
    ec.contract_type,
    ec.contract_hours_type,
    ec.start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    ec.previous_contract_id,
    ec.contract_event_type,
    ec.hours_per_week,
    ec.min_hours_per_week,
    ec.max_hours_per_week,
    ec.roster_free_day,
    ec.wage_tax_table,
    ec.created_at,
    ec.updated_at
FROM employee_contracts ec
JOIN departments d ON d.id = ec.department_id
JOIN location l ON l.id = ec.location_id
LEFT JOIN organizational_roles org_role ON org_role.id = ec.organizational_role_id
WHERE ec.employee_id = $1
ORDER BY ec.start_date DESC, ec.created_at DESC;

-- name: UpdateEmployeeContract :one
UPDATE employee_contracts
SET
    job_title = COALESCE(sqlc.narg('job_title')::employee_job_title_enum, job_title),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    location_id = COALESCE(sqlc.narg('location_id'), location_id),
    organizational_role_id = COALESCE(sqlc.narg('organizational_role_id'), organizational_role_id),
    contract_type = COALESCE(sqlc.narg('contract_type')::employee_contract_type_enum, contract_type),
    contract_hours_type = COALESCE(sqlc.narg('contract_hours_type')::contract_hours_type_enum, contract_hours_type),
    start_date = COALESCE(sqlc.narg('start_date')::date, start_date),
    contract_end_date = COALESCE(sqlc.narg('contract_end_date')::date, contract_end_date),
    hours_per_week = COALESCE(sqlc.narg('hours_per_week'), hours_per_week),
    min_hours_per_week = COALESCE(sqlc.narg('min_hours_per_week'), min_hours_per_week),
    max_hours_per_week = COALESCE(sqlc.narg('max_hours_per_week'), max_hours_per_week),
    roster_free_day = COALESCE(sqlc.narg('roster_free_day'), roster_free_day),
    wage_tax_table = COALESCE(sqlc.narg('wage_tax_table')::wage_tax_table_enum, wage_tax_table),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND employee_id = sqlc.arg('employee_id')
RETURNING *;

-- name: GetEmployeeContractByID :one
SELECT *
FROM employee_contracts
WHERE id = $1;

-- name: EndEmployeeContractSegment :one
UPDATE employee_contracts
SET
    effective_end_date = sqlc.arg('effective_end_date')::date,
    updated_by_employee_id = sqlc.narg('updated_by_employee_id'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: AddEmployeeContractAmendment :one
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
    previous_contract_id,
    contract_event_type,
    change_reason,
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
    sqlc.arg('roster_free_day'),
    sqlc.narg('wage_tax_table'),
    sqlc.arg('previous_contract_id'),
    'amendment',
    sqlc.narg('change_reason'),
    sqlc.narg('created_by_employee_id')
)
RETURNING *;

-- name: AddEmployeeNewContract :one
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
    previous_contract_id,
    contract_event_type,
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
    sqlc.arg('roster_free_day'),
    sqlc.narg('wage_tax_table'),
    sqlc.narg('previous_contract_id'),
    'new_contract',
    sqlc.narg('created_by_employee_id')
)
RETURNING *;

-- name: GetLatestEmployeeSalaryAssignmentDetail :one
SELECT
    esa.id,
    esa.employee_id,
    esa.contract_id,
    esa.salary_scale_step_id,
    cst.cao_code,
    cst.name AS salary_table_name,
    css.scale,
    css.step,
    css.ip_number,
    css.monthly_salary,
    css.hourly_rate,
    esa.effective_from,
    esa.effective_to,
    esa.created_at,
    esa.updated_at
FROM employee_salary_assignments esa
JOIN cao_salary_scale_steps css ON css.id = esa.salary_scale_step_id
JOIN cao_salary_tables cst ON cst.id = css.salary_table_id
WHERE esa.employee_id = $1
ORDER BY esa.effective_from DESC, esa.created_at DESC
LIMIT 1;
