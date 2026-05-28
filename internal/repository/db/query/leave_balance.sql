-- name: EnsureLeaveBalanceForYear :exec
INSERT INTO leave_balances (
    employee_id,
    year,
    legal_adjustment_minutes,
    extra_total_minutes,
    legal_used_minutes,
    extra_used_minutes
) SELECT
    ep.id,
    sqlc.arg('year'),
    0,
    0,
    0,
    0
FROM employee_profile ep
WHERE ep.id = sqlc.arg(employee_id)
ON CONFLICT (employee_id, year) DO NOTHING;

-- name: ComputeLegalLeaveTotalForYear :one
SELECT calculate_legal_leave_minutes(
    sqlc.arg('employee_id'),
    sqlc.arg('year')::int,
    sqlc.arg('as_of')::timestamptz
)::int AS legal_total_minutes;

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
    extra_used_minutes = extra_used_minutes + sqlc.arg(extra_minutes),
    legal_used_minutes = legal_used_minutes + sqlc.arg(legal_minutes),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ListLeaveBalancesPaginated :many
SELECT
    lb.id,
    lb.employee_id,
    lb.year,
    (ent.legal_calculated_minutes + lb.legal_adjustment_minutes)::int AS legal_total_minutes,
    lb.legal_adjustment_minutes,
    lb.extra_total_minutes,
    lb.legal_used_minutes,
    lb.extra_used_minutes,
    lb.created_at,
    lb.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    COUNT(*) OVER() AS total_count
FROM leave_balances lb
JOIN employee_profile ep ON ep.id = lb.employee_id
JOIN LATERAL (
    SELECT calculate_legal_leave_minutes(
        lb.employee_id,
        lb.year,
        CASE
            WHEN lb.year < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(lb.year, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN lb.year = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(lb.year, 1, 1)::timestamptz
        END
    )::int AS legal_calculated_minutes
) ent ON true
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
    (ent.legal_calculated_minutes + lb.legal_adjustment_minutes)::int AS legal_total_minutes,
    lb.legal_adjustment_minutes,
    lb.extra_total_minutes,
    lb.legal_used_minutes,
    lb.extra_used_minutes,
    lb.created_at,
    lb.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    COUNT(*) OVER() AS total_count
FROM leave_balances lb
JOIN employee_profile ep ON ep.id = lb.employee_id
JOIN LATERAL (
    SELECT calculate_legal_leave_minutes(
        lb.employee_id,
        lb.year,
        CASE
            WHEN lb.year < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(lb.year, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN lb.year = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(lb.year, 1, 1)::timestamptz
        END
    )::int AS legal_calculated_minutes
) ent ON true
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
    legal_adjustment_minutes = legal_adjustment_minutes + sqlc.arg('legal_adjustment_minutes_delta'),
    extra_total_minutes = extra_total_minutes + sqlc.arg('extra_total_minutes_delta'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CreateLeaveBalanceAdjustmentAudit :one
INSERT INTO leave_balance_adjustments (
    leave_balance_id,
    employee_id,
    year,
    legal_adjustment_minutes_delta,
    extra_total_minutes_delta,
    reason,
    adjusted_by_employee_id,
    legal_adjustment_minutes_before,
    extra_total_minutes_before,
    legal_adjustment_minutes_after,
    extra_total_minutes_after
) VALUES (
    sqlc.arg('leave_balance_id'),
    sqlc.arg('employee_id'),
    sqlc.arg('year'),
    sqlc.arg('legal_adjustment_minutes_delta'),
    sqlc.arg('extra_total_minutes_delta'),
    sqlc.arg('reason'),
    sqlc.arg('adjusted_by_employee_id'),
    sqlc.arg('legal_adjustment_minutes_before'),
    sqlc.arg('extra_total_minutes_before'),
    sqlc.arg('legal_adjustment_minutes_after'),
    sqlc.arg('extra_total_minutes_after')
)
RETURNING *;
