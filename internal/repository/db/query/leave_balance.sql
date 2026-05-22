-- name: EnsureLeaveBalanceForYear :exec
INSERT INTO leave_balances (
    employee_id,
    year,
    legal_total_hours,
    extra_total_hours,
    legal_used_hours,
    extra_used_hours
) SELECT
    ep.id,
    sqlc.arg('year'),
    calculate_legal_leave_hours(ep.id, sqlc.arg('year')::int),
    0,
    0,
    0
FROM employee_profile ep
WHERE ep.id = sqlc.arg(employee_id)
ON CONFLICT (employee_id, year) DO NOTHING;

-- name: ComputeLegalLeaveTotalForYear :one
SELECT calculate_legal_leave_hours(sqlc.arg('employee_id'), sqlc.arg('year'))::int AS legal_total_hours;

-- name: GetEmployeeContractForLeave :one
SELECT
    hours_per_week AS contract_hours,
    contract_type
FROM employee_contracts
WHERE employee_id = sqlc.arg('employee_id')
ORDER BY start_date DESC, created_at DESC
LIMIT 1;

-- name: LockLeaveBalanceByEmployeeYear :one
SELECT *
FROM leave_balances
WHERE employee_id = sqlc.arg(employee_id)
  AND year = sqlc.arg('year')
FOR UPDATE;

-- name: ListLeaveBalancesForEmployeeFromYearForUpdate :many
SELECT *
FROM leave_balances
WHERE employee_id = sqlc.arg('employee_id')
  AND year >= sqlc.arg('year_from')
ORDER BY year
FOR UPDATE;

-- name: ApplyLeaveBalanceDeduction :one
UPDATE leave_balances
SET
    extra_used_hours = extra_used_hours + sqlc.arg(extra_hours),
    legal_used_hours = legal_used_hours + sqlc.arg(legal_hours),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ListLeaveBalancesPaginated :many
SELECT
    lb.id,
    lb.employee_id,
    lb.year,
    lb.legal_total_hours,
    lb.extra_total_hours,
    lb.legal_used_hours,
    lb.extra_used_hours,
    lb.created_at,
    lb.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    COUNT(*) OVER() AS total_count
FROM leave_balances lb
JOIN employee_profile ep ON ep.id = lb.employee_id
LEFT JOIN LATERAL (
    SELECT *
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
WHERE (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
)
  AND (
    sqlc.narg('year')::int IS NULL
    OR lb.year = sqlc.narg('year')::int
)
ORDER BY lb.year DESC, lb.updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListMyLeaveBalancesPaginated :many
SELECT
    lb.id,
    lb.employee_id,
    lb.year,
    lb.legal_total_hours,
    lb.extra_total_hours,
    lb.legal_used_hours,
    lb.extra_used_hours,
    lb.created_at,
    lb.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    COUNT(*) OVER() AS total_count
FROM leave_balances lb
JOIN employee_profile ep ON ep.id = lb.employee_id
LEFT JOIN LATERAL (
    SELECT *
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
WHERE lb.employee_id = sqlc.arg('employee_id')
  AND (
    sqlc.narg('year')::int IS NULL
    OR lb.year = sqlc.narg('year')::int
)
ORDER BY lb.year DESC, lb.updated_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ApplyLeaveBalanceTotalAdjustment :one
UPDATE leave_balances
SET
    legal_total_hours = legal_total_hours + sqlc.arg('legal_hours_delta'),
    extra_total_hours = extra_total_hours + sqlc.arg('extra_hours_delta'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CreateLeaveBalanceAdjustmentAudit :one
INSERT INTO leave_balance_adjustments (
    leave_balance_id,
    employee_id,
    year,
    legal_hours_delta,
    extra_hours_delta,
    reason,
    adjusted_by_employee_id,
    legal_total_hours_before,
    extra_total_hours_before,
    legal_total_hours_after,
    extra_total_hours_after
) VALUES (
    sqlc.arg('leave_balance_id'),
    sqlc.arg('employee_id'),
    sqlc.arg('year'),
    sqlc.arg('legal_hours_delta'),
    sqlc.arg('extra_hours_delta'),
    sqlc.arg('reason'),
    sqlc.arg('adjusted_by_employee_id'),
    sqlc.arg('legal_total_hours_before'),
    sqlc.arg('extra_total_hours_before'),
    sqlc.arg('legal_total_hours_after'),
    sqlc.arg('extra_total_hours_after')
)
RETURNING *;
