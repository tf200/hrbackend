-- name: ListPayrollMonthEmployeesPaginated :many
WITH month_employees AS (
    SELECT DISTINCT pp.employee_id
    FROM pay_periods pp
    WHERE pp.period_start = sqlc.arg('month_start')
      AND pp.period_end = sqlc.arg('month_end')

    UNION

    SELECT DISTINCT s.employee_id
    FROM schedules s
    WHERE DATE(s.start_datetime) >= sqlc.arg('month_start')
      AND DATE(s.start_datetime) <= sqlc.arg('month_end')

    UNION

    SELECT DISTINCT oe.employee_id
    FROM overtime_entries oe
    WHERE oe.entry_date >= sqlc.arg('month_start')
      AND oe.entry_date <= sqlc.arg('month_end')
      AND oe.status IN ('approved'::overtime_status_enum, 'submitted'::overtime_status_enum)
)
SELECT
    ep.id AS employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM month_employees me
JOIN employee_profile ep ON ep.id = me.employee_id
WHERE (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
)
ORDER BY ep.first_name ASC, ep.last_name ASC, ep.id ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListPayrollMonthEmployeesAll :many
WITH month_employees AS (
    SELECT DISTINCT pp.employee_id
    FROM pay_periods pp
    WHERE pp.period_start = sqlc.arg('month_start')
      AND pp.period_end = sqlc.arg('month_end')

    UNION

    SELECT DISTINCT s.employee_id
    FROM schedules s
    WHERE DATE(s.start_datetime) >= sqlc.arg('month_start')
      AND DATE(s.start_datetime) <= sqlc.arg('month_end')

    UNION

    SELECT DISTINCT oe.employee_id
    FROM overtime_entries oe
    WHERE oe.entry_date >= sqlc.arg('month_start')
      AND oe.entry_date <= sqlc.arg('month_end')
      AND oe.status IN ('approved'::overtime_status_enum, 'submitted'::overtime_status_enum)
)
SELECT
    ep.id AS employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name
FROM month_employees me
JOIN employee_profile ep ON ep.id = me.employee_id
WHERE (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
)
ORDER BY ep.first_name ASC, ep.last_name ASC, ep.id ASC;

-- name: ListPayPeriodsByEmployeeIDsAndRange :many
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
WHERE pp.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND pp.period_start = sqlc.arg('month_start')
  AND pp.period_end = sqlc.arg('month_end')
ORDER BY pp.employee_id ASC, pp.created_at DESC;

-- name: ListLockedPayPeriodMultiplierSummaries :many
SELECT
    ppl.pay_period_id,
    ppl.applied_rate_percent,
    COALESCE(SUM(ppl.minutes_worked), 0)::double precision AS worked_minutes,
    COALESCE(SUM(ppl.minutes_worked), 0)::double precision AS paid_minutes,
    COALESCE(SUM(ppl.base_amount), 0)::double precision AS base_amount,
    COALESCE(SUM(ppl.premium_amount), 0)::double precision AS premium_amount
FROM pay_period_line_items ppl
WHERE ppl.pay_period_id = ANY(sqlc.arg('pay_period_ids')::uuid[])
GROUP BY ppl.pay_period_id, ppl.applied_rate_percent
ORDER BY ppl.pay_period_id ASC, ppl.applied_rate_percent ASC;

-- name: ListPayrollMonthApprovedWorkItems :many
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
        cc.contract_type,
        css.hourly_rate::double precision AS contract_rate,
        'none'::text AS irregular_hours_profile
    FROM schedules s
    JOIN employee_profile ep ON ep.id = s.employee_id
    JOIN LATERAL (
        SELECT c.contract_type
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
    WHERE s.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
      AND DATE(s.start_datetime) >= sqlc.arg('month_start')
      AND DATE(s.start_datetime) <= sqlc.arg('month_end')
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
        cc.contract_type,
        css.hourly_rate::double precision AS contract_rate,
        'none'::text AS irregular_hours_profile
    FROM overtime_entries oe
    JOIN employee_profile ep ON ep.id = oe.employee_id
    JOIN LATERAL (
        SELECT c.contract_type
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
    WHERE oe.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
      AND oe.status = 'approved'::overtime_status_enum
      AND oe.entry_date >= sqlc.arg('month_start')
      AND oe.entry_date <= sqlc.arg('month_end')
)
SELECT * FROM schedule_items
UNION ALL
SELECT * FROM overtime_items
ORDER BY employee_id ASC, work_date ASC, source_type ASC, source_id ASC;

-- name: ListPayrollMonthPendingOvertimeSummaries :many
SELECT
    oe.employee_id,
    COUNT(*)::INT AS pending_entry_count,
    COALESCE(SUM(oe.minutes), 0)::INT AS pending_worked_minutes
FROM overtime_entries oe
WHERE oe.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND oe.status = 'submitted'::overtime_status_enum
  AND oe.entry_date >= sqlc.arg('month_start')
  AND oe.entry_date <= sqlc.arg('month_end')
GROUP BY oe.employee_id
ORDER BY oe.employee_id ASC;

-- name: ListPayrollMonthPendingOvertimeEntries :many
SELECT
    oe.employee_id,
    oe.minutes AS worked_minutes,
    cc.contract_type
FROM overtime_entries oe
JOIN employee_profile ep ON ep.id = oe.employee_id
LEFT JOIN LATERAL (
    SELECT c.contract_type
    FROM employee_contracts c
    WHERE c.employee_id = oe.employee_id
      AND c.start_date <= oe.entry_date
      AND (c.contract_end_date IS NULL OR c.contract_end_date >= oe.entry_date)
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) cc ON TRUE
WHERE oe.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND oe.status = 'submitted'::overtime_status_enum
  AND oe.entry_date >= sqlc.arg('month_start')
  AND oe.entry_date <= sqlc.arg('month_end')
ORDER BY oe.employee_id ASC, oe.entry_date ASC, oe.created_at ASC;
