-- name: GetPayPeriodByEmployeePeriod :one
SELECT
    pp.id,
    pp.employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    pp.period_start,
    pp.period_end,
    pp.status,
    pp.base_gross_amount,
    pp.irregular_gross_amount,
    pp.gross_amount,
    pp.paid_at,
    pp.created_by_employee_id,
    pp.created_at,
    pp.updated_at
FROM pay_periods pp
JOIN employee_profile ep ON ep.id = pp.employee_id
WHERE pp.employee_id = sqlc.arg('employee_id')
  AND pp.period_start = sqlc.arg('period_start')
  AND pp.period_end = sqlc.arg('period_end');

-- name: LockPayrollPreviewTimeEntries :many
SELECT
    te.id,
    te.employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    te.entry_date,
    te.start_time,
    te.end_time,
    te.break_minutes,
    te.hour_type,
    COALESCE(cc.contract_type, ep.contract_type) AS contract_type,
    COALESCE(cc.contract_rate, ep.contract_rate) AS contract_rate,
    COALESCE(cc.irregular_hours_profile, ep.irregular_hours_profile) AS irregular_hours_profile
FROM time_entries te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN LATERAL (
    SELECT
        c.contract_type,
        c.contract_rate,
        c.irregular_hours_profile
    FROM employee_contract_changes c
    WHERE c.employee_id = te.employee_id
      AND c.effective_from <= te.entry_date
    ORDER BY c.effective_from DESC, c.created_at DESC
    LIMIT 1
) cc ON TRUE
WHERE te.employee_id = sqlc.arg(employee_id)
  AND te.status = 'approved'::time_entry_status_enum
  AND te.paid_period_id IS NULL
  AND te.hour_type IN (
      'normal'::time_entry_hour_type_enum,
      'overtime'::time_entry_hour_type_enum,
      'travel'::time_entry_hour_type_enum,
      'training'::time_entry_hour_type_enum
  )
  AND te.entry_date >= sqlc.arg(period_start)
  AND te.entry_date <= sqlc.arg(period_end)
ORDER BY te.entry_date ASC, te.start_time ASC, te.created_at ASC
FOR UPDATE OF te;

-- name: CreatePayPeriod :one
INSERT INTO pay_periods (
    employee_id,
    period_start,
    period_end,
    status,
    base_gross_amount,
    irregular_gross_amount,
    gross_amount,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('period_start'),
    sqlc.arg('period_end'),
    'draft'::pay_period_status_enum,
    sqlc.arg('base_gross_amount'),
    sqlc.arg('irregular_gross_amount'),
    sqlc.arg('gross_amount'),
    sqlc.arg('created_by_employee_id')
)
RETURNING *;

-- name: CreatePayPeriodLineItem :one
INSERT INTO pay_period_line_items (
    pay_period_id,
    time_entry_id,
    work_date,
    line_type,
    irregular_hours_profile,
    applied_rate_percent,
    minutes_worked,
    base_amount,
    premium_amount,
    metadata
) VALUES (
    sqlc.arg('pay_period_id'),
    sqlc.narg('time_entry_id'),
    sqlc.arg('work_date'),
    sqlc.arg('line_type'),
    sqlc.arg('irregular_hours_profile'),
    sqlc.arg('applied_rate_percent'),
    sqlc.arg('minutes_worked'),
    sqlc.arg('base_amount'),
    sqlc.arg('premium_amount'),
    COALESCE(sqlc.narg('metadata'), '{}'::jsonb)
)
RETURNING *;

-- name: AssignTimeEntriesToPayPeriod :exec
UPDATE time_entries
SET
    paid_period_id = sqlc.arg('pay_period_id'),
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('time_entry_ids')::uuid[]);

-- name: GetPayPeriodByID :one
SELECT
    pp.id,
    pp.employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    pp.period_start,
    pp.period_end,
    pp.status,
    pp.base_gross_amount,
    pp.irregular_gross_amount,
    pp.gross_amount,
    pp.paid_at,
    pp.created_by_employee_id,
    pp.created_at,
    pp.updated_at
FROM pay_periods pp
JOIN employee_profile ep ON ep.id = pp.employee_id
WHERE pp.id = sqlc.arg('id');

-- name: ListPayPeriodsPaginated :many
SELECT
    pp.id,
    pp.employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    pp.period_start,
    pp.period_end,
    pp.status,
    pp.base_gross_amount,
    pp.irregular_gross_amount,
    pp.gross_amount,
    pp.paid_at,
    pp.created_by_employee_id,
    pp.created_at,
    pp.updated_at,
    COUNT(*) OVER() AS total_count
FROM pay_periods pp
JOIN employee_profile ep ON ep.id = pp.employee_id
WHERE (
    sqlc.narg('status')::pay_period_status_enum IS NULL
    OR pp.status = sqlc.narg('status')::pay_period_status_enum
)
  AND (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
  )
ORDER BY pp.period_start DESC, pp.period_end DESC, pp.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListPayPeriodLineItemsByPayPeriodID :many
SELECT
    id,
    pay_period_id,
    time_entry_id,
    work_date,
    line_type,
    irregular_hours_profile,
    applied_rate_percent,
    minutes_worked,
    base_amount,
    premium_amount,
    metadata,
    created_at,
    updated_at
FROM pay_period_line_items
WHERE pay_period_id = sqlc.arg('pay_period_id')
ORDER BY work_date ASC, created_at ASC;

-- name: LockPayPeriodByID :one
SELECT *
FROM pay_periods
WHERE id = sqlc.arg('id')
FOR UPDATE;

-- name: MarkPayPeriodPaid :one
UPDATE pay_periods
SET
    status = 'paid'::pay_period_status_enum,
    paid_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
