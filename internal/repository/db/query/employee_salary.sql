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

-- name: GetActiveEmployeeSalaryAssignment :one
SELECT
    id,
    contract_id,
    salary_scale_step_id,
    effective_from,
    effective_to
FROM employee_salary_assignments
WHERE employee_id = sqlc.arg('employee_id')
  AND (
      sqlc.narg('contract_id')::uuid IS NULL
      OR contract_id IS NULL
      OR contract_id = sqlc.narg('contract_id')::uuid
  )
  AND effective_from <= sqlc.arg('target_date')::date
  AND (effective_to IS NULL OR effective_to > sqlc.arg('target_date')::date)
ORDER BY
    (contract_id = sqlc.narg('contract_id')::uuid) DESC,
    effective_from DESC,
    created_at DESC
LIMIT 1;

-- name: EndEmployeeSalaryAssignment :exec
UPDATE employee_salary_assignments
SET
    effective_to = sqlc.arg('effective_to')::date,
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND (effective_to IS NULL OR effective_to > sqlc.arg('effective_to')::date);

-- name: GetEmployeeSalaryAssignmentByContract :one
SELECT *
FROM employee_salary_assignments
WHERE employee_id = sqlc.arg('employee_id')
  AND contract_id = sqlc.arg('contract_id')
ORDER BY effective_from DESC, created_at DESC
LIMIT 1;

-- name: UpdateEmployeeSalaryAssignmentScaleStep :one
UPDATE employee_salary_assignments
SET
    salary_scale_step_id = sqlc.arg('salary_scale_step_id'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: GetEmployeeSalaryAssignmentDetailByID :one
SELECT
    esa.id,
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
WHERE esa.id = sqlc.arg('id');
