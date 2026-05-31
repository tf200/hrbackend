-- name: GetPayPeriodByEmployeePeriod :one
SELECT
    pp.id,
    pp.employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    pp.period_start,
    pp.period_end,
    pp.payroll_group,
    pp.cutoff_at,
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
  AND pp.period_end = sqlc.arg('period_end')
  AND pp.payroll_group = sqlc.arg('payroll_group');

-- name: LockPayrollPreviewWorkItems :many
WITH schedule_items AS (
    SELECT
        s.id AS source_id,
        s.employee_id,
        ep.first_name AS employee_first_name,
        ep.last_name AS employee_last_name,
        COALESCE(NULLIF(btrim(s.shift_name_snapshot), ''), 'Scheduled shift') AS label,
        DATE(s.start_datetime) AS work_date,
        s.start_datetime::time AS start_time_val,
        s.end_datetime::time AS end_time_val,
        0 AS break_minutes,
        EXTRACT(EPOCH FROM (s.end_datetime - s.start_datetime)) / 60 AS minutes_worked,
        'schedule'::text AS source_type,
        s.id AS schedule_id,
        NULL::uuid AS overtime_entry_id,
        NULL::uuid AS leave_payout_request_id,
        cc.contract_type,
        css.hourly_rate::double precision AS contract_rate,
        NULL::double precision AS gross_amount_override,
        'none'::text AS irregular_hours_profile
    FROM schedules s
    JOIN employee_profile ep ON ep.id = s.employee_id
    JOIN LATERAL (
        SELECT
            c.id,
            c.contract_type,
            c.contract_end_date
        FROM employee_contracts c
        WHERE c.employee_id = s.employee_id
          AND c.start_date <= DATE(s.start_datetime)
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= DATE(s.start_datetime))
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= DATE(s.start_datetime))
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    JOIN LATERAL (
        SELECT esa.salary_scale_step_id
        FROM employee_salary_assignments esa
        WHERE esa.employee_id = s.employee_id
          AND (esa.contract_id IS NULL OR esa.contract_id = cc.id)
          AND esa.effective_from <= DATE(s.start_datetime)
          AND (esa.effective_to IS NULL OR esa.effective_to > DATE(s.start_datetime))
        ORDER BY
          (esa.contract_id = cc.id) DESC,
          esa.effective_from DESC,
          esa.created_at DESC
        LIMIT 1
    ) latest_salary ON TRUE
    JOIN cao_salary_scale_steps css ON css.id = latest_salary.salary_scale_step_id
    WHERE s.employee_id = sqlc.arg(employee_id)
      AND DATE(s.start_datetime) >= sqlc.arg(period_start)
      AND s.end_datetime <= sqlc.arg(cutoff_at)
      AND s.paid_period_id IS NULL
),
overtime_items AS (
    SELECT
        oe.id AS source_id,
        oe.employee_id,
        ep.first_name AS employee_first_name,
        ep.last_name AS employee_last_name,
        COALESCE(NULLIF(btrim(oe.reason::text), ''), 'Overtime') AS label,
        oe.entry_date AS work_date,
        NULL::time AS start_time_val,
        NULL::time AS end_time_val,
        0 AS break_minutes,
        oe.minutes::double precision AS minutes_worked,
        'overtime'::text AS source_type,
        NULL::uuid AS schedule_id,
        oe.id AS overtime_entry_id,
        NULL::uuid AS leave_payout_request_id,
        cc.contract_type,
        css.hourly_rate::double precision AS contract_rate,
        NULL::double precision AS gross_amount_override,
        'none'::text AS irregular_hours_profile
    FROM overtime_entries oe
    JOIN employee_profile ep ON ep.id = oe.employee_id
    JOIN LATERAL (
        SELECT
            c.id,
            c.contract_type,
            c.contract_end_date
        FROM employee_contracts c
        WHERE c.employee_id = oe.employee_id
          AND c.start_date <= oe.entry_date
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= oe.entry_date)
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= oe.entry_date)
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    JOIN LATERAL (
        SELECT esa.salary_scale_step_id
        FROM employee_salary_assignments esa
        WHERE esa.employee_id = oe.employee_id
          AND (esa.contract_id IS NULL OR esa.contract_id = cc.id)
          AND esa.effective_from <= oe.entry_date
          AND (esa.effective_to IS NULL OR esa.effective_to > oe.entry_date)
        ORDER BY
          (esa.contract_id = cc.id) DESC,
          esa.effective_from DESC,
          esa.created_at DESC
        LIMIT 1
    ) latest_salary ON TRUE
    JOIN cao_salary_scale_steps css ON css.id = latest_salary.salary_scale_step_id
    WHERE oe.employee_id = sqlc.arg(employee_id)
      AND oe.status = 'approved'::overtime_status_enum
      AND oe.paid_period_id IS NULL
      AND oe.entry_date >= sqlc.arg(period_start)
      AND oe.entry_date <= sqlc.arg(cutoff_date)
),
leave_payout_items AS (
    SELECT
        lpr.id AS source_id,
        lpr.employee_id,
        ep.first_name AS employee_first_name,
        ep.last_name AS employee_last_name,
        'Leave payout'::text AS label,
        lpr.salary_month AS work_date,
        NULL::time AS start_time_val,
        NULL::time AS end_time_val,
        0 AS break_minutes,
        (lpr.requested_hours * 60)::double precision AS minutes_worked,
        'leave_payout'::text AS source_type,
        NULL::uuid AS schedule_id,
        NULL::uuid AS overtime_entry_id,
        lpr.id AS leave_payout_request_id,
        cc.contract_type,
        lpr.hourly_rate::double precision AS contract_rate,
        lpr.gross_amount::double precision AS gross_amount_override,
        'none'::text AS irregular_hours_profile
    FROM leave_payout_requests lpr
    JOIN employee_profile ep ON ep.id = lpr.employee_id
    JOIN LATERAL (
        SELECT c.contract_type
        FROM employee_contracts c
        WHERE c.employee_id = lpr.employee_id
          AND c.start_date <= lpr.salary_month
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= lpr.salary_month)
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= lpr.salary_month)
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    WHERE lpr.employee_id = sqlc.arg(employee_id)
      AND lpr.status = 'approved'::payout_request_status_enum
      AND lpr.paid_period_id IS NULL
      AND lpr.salary_month >= sqlc.arg(period_start)
      AND lpr.salary_month <= sqlc.arg(cutoff_date)
)
SELECT * FROM schedule_items
UNION ALL
SELECT * FROM overtime_items
UNION ALL
SELECT * FROM leave_payout_items
ORDER BY work_date ASC, source_type ASC, source_id ASC;

-- name: LockPayrollOvertimeEntries :many
SELECT id
FROM overtime_entries
WHERE employee_id = sqlc.arg(employee_id)
  AND status = 'approved'::overtime_status_enum
  AND paid_period_id IS NULL
  AND entry_date >= sqlc.arg(period_start)
  AND entry_date <= sqlc.arg(cutoff_date)
FOR UPDATE;

-- name: CreatePayPeriod :one
INSERT INTO pay_periods (
    employee_id,
    period_start,
    period_end,
    payroll_group,
    cutoff_at,
    status,
    base_gross_amount,
    irregular_gross_amount,
    gross_amount,
    created_by_employee_id
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('period_start'),
    sqlc.arg('period_end'),
    sqlc.arg('payroll_group'),
    sqlc.narg('cutoff_at'),
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
    schedule_id,
    overtime_entry_id,
    leave_payout_request_id,
    contract_type,
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
    sqlc.narg('schedule_id'),
    sqlc.narg('overtime_entry_id'),
    sqlc.narg('leave_payout_request_id'),
    sqlc.arg('contract_type'),
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

-- name: AssignOvertimeEntriesToPayPeriod :exec
UPDATE overtime_entries
SET
    paid_period_id = sqlc.arg('pay_period_id'),
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('overtime_entry_ids')::uuid[]);

-- name: AssignSchedulesToPayPeriod :exec
UPDATE schedules
SET
    paid_period_id = sqlc.arg('pay_period_id'),
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('schedule_ids')::uuid[]);

-- name: AssignLeavePayoutRequestsToPayPeriod :exec
UPDATE leave_payout_requests
SET
    paid_period_id = sqlc.arg('pay_period_id'),
    updated_at = NOW()
WHERE id = ANY(sqlc.arg('leave_payout_request_ids')::uuid[]);

-- name: GetPayPeriodByID :one
SELECT
    pp.id,
    pp.employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    pp.period_start,
    pp.period_end,
    pp.payroll_group,
    pp.cutoff_at,
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
    pp.payroll_group,
    pp.cutoff_at,
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
    schedule_id,
    overtime_entry_id,
    leave_payout_request_id,
    contract_type,
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

-- name: MarkLeavePayoutRequestsPaidByPayPeriod :exec
UPDATE leave_payout_requests
SET
    status = 'paid'::payout_request_status_enum,
    paid_by_employee_id = sqlc.arg('paid_by_employee_id'),
    paid_at = NOW(),
    updated_at = NOW()
WHERE paid_period_id = sqlc.arg('pay_period_id')
  AND status = 'approved'::payout_request_status_enum;
