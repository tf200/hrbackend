-- name: CreateEmployeeSalaryAssignment :one
INSERT INTO employee_salary_assignments (
    employee_id,
    contract_id,
    salary_scale_step_id,
    effective_from,
    effective_to,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.narg('contract_id'),
    sqlc.arg('salary_scale_step_id'),
    sqlc.arg('effective_from'),
    sqlc.narg('effective_to'),
    sqlc.narg('created_by_employee_id')
)
RETURNING *;
