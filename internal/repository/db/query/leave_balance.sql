-- name: EnsureLeaveBalanceForYear :exec
INSERT INTO leave_balances (
    employee_id,
    year
) VALUES (
    sqlc.arg(employee_id),
    sqlc.arg('year')
)
ON CONFLICT (employee_id, year) DO NOTHING;

-- name: LockLeaveBalanceByEmployeeYear :one
SELECT *
FROM leave_balances
WHERE employee_id = sqlc.arg(employee_id)
  AND year = sqlc.arg('year')
FOR UPDATE;

-- name: ApplyLeaveBalanceDeduction :one
UPDATE leave_balances
SET
    extra_used_days = extra_used_days + sqlc.arg(extra_days),
    legal_used_days = legal_used_days + sqlc.arg(legal_days),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ListLeaveBalancesPaginated :many
SELECT
    lb.id,
    lb.employee_id,
    lb.year,
    lb.legal_total_days,
    lb.extra_total_days,
    lb.legal_used_days,
    lb.extra_used_days,
    lb.created_at,
    lb.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM leave_balances lb
JOIN employee_profile ep ON ep.id = lb.employee_id
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
    lb.legal_total_days,
    lb.extra_total_days,
    lb.legal_used_days,
    lb.extra_used_days,
    lb.created_at,
    lb.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM leave_balances lb
JOIN employee_profile ep ON ep.id = lb.employee_id
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
    legal_total_days = legal_total_days + sqlc.arg('legal_days_delta'),
    extra_total_days = extra_total_days + sqlc.arg('extra_days_delta'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CreateLeaveBalanceAdjustmentAudit :one
INSERT INTO leave_balance_adjustments (
    leave_balance_id,
    employee_id,
    year,
    legal_days_delta,
    extra_days_delta,
    reason,
    adjusted_by_employee_id,
    legal_total_days_before,
    extra_total_days_before,
    legal_total_days_after,
    extra_total_days_after
) VALUES (
    sqlc.arg('leave_balance_id'),
    sqlc.arg('employee_id'),
    sqlc.arg('year'),
    sqlc.arg('legal_days_delta'),
    sqlc.arg('extra_days_delta'),
    sqlc.arg('reason'),
    sqlc.arg('adjusted_by_employee_id'),
    sqlc.arg('legal_total_days_before'),
    sqlc.arg('extra_total_days_before'),
    sqlc.arg('legal_total_days_after'),
    sqlc.arg('extra_total_days_after')
)
RETURNING *;
