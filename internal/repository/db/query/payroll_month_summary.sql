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

    UNION

    SELECT DISTINCT lpr.employee_id
    FROM leave_payout_requests lpr
    WHERE lpr.pay_period_start >= sqlc.arg('month_start')
      AND lpr.pay_period_start <= sqlc.arg('month_end')
      AND lpr.status = 'approved'::payout_request_status_enum
      AND lpr.paid_period_id IS NULL
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

    UNION

    SELECT DISTINCT lpr.employee_id
    FROM leave_payout_requests lpr
    WHERE lpr.pay_period_start >= sqlc.arg('month_start')
      AND lpr.pay_period_start <= sqlc.arg('month_end')
      AND lpr.status = 'approved'::payout_request_status_enum
      AND lpr.paid_period_id IS NULL
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

-- name: ListFixedPayrollMonthEmployeesPaginated :many
WITH fixed_month_employees AS (
    SELECT DISTINCT c.employee_id
    FROM employee_contracts c
    JOIN employee_salary_assignments esa ON esa.employee_id = c.employee_id
      AND (esa.contract_id IS NULL OR esa.contract_id = c.id)
    WHERE c.contract_type IN ('permanent'::employee_contract_type_enum, 'temporary'::employee_contract_type_enum)
      AND c.start_date <= sqlc.arg('month_end')
      AND (c.effective_end_date IS NULL OR c.effective_end_date >= sqlc.arg('month_start'))
      AND (c.contract_end_date IS NULL OR c.contract_end_date >= sqlc.arg('month_start'))
      AND esa.effective_from <= LEAST(
          sqlc.arg('month_end')::date,
          COALESCE(c.effective_end_date, sqlc.arg('month_end')::date),
          COALESCE(c.contract_end_date, sqlc.arg('month_end')::date)
      )
      AND (esa.effective_to IS NULL OR esa.effective_to > GREATEST(sqlc.arg('month_start')::date, c.start_date))
)
SELECT
    ep.id AS employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM fixed_month_employees fme
JOIN employee_profile ep ON ep.id = fme.employee_id
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

-- name: ListFixedPayrollContractSegments :many
WITH contract_windows AS (
    SELECT
        c.id AS contract_id,
        c.employee_id,
        c.contract_type,
        GREATEST(c.start_date, sqlc.arg('month_start')::date)::date AS active_from,
        LEAST(
            sqlc.arg('month_end')::date,
            COALESCE(c.effective_end_date, sqlc.arg('month_end')::date),
            COALESCE(c.contract_end_date, sqlc.arg('month_end')::date)
        )::date AS active_until,
        c.hours_per_week::double precision AS hours_per_week
    FROM employee_contracts c
    WHERE c.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
      AND c.contract_type IN ('permanent'::employee_contract_type_enum, 'temporary'::employee_contract_type_enum)
      AND c.start_date <= sqlc.arg('month_end')
      AND (c.effective_end_date IS NULL OR c.effective_end_date >= sqlc.arg('month_start'))
      AND (c.contract_end_date IS NULL OR c.contract_end_date >= sqlc.arg('month_start'))
)
SELECT
    cw.employee_id,
    cw.contract_id,
    cw.contract_type,
    cw.active_from,
    cw.active_until,
    cw.hours_per_week,
    cst.full_time_hours_per_week::double precision AS full_time_hours_per_week,
    css.monthly_salary::double precision AS monthly_salary,
    css.hourly_rate::double precision AS hourly_rate
FROM contract_windows cw
JOIN LATERAL (
    SELECT esa.salary_scale_step_id
    FROM employee_salary_assignments esa
    WHERE esa.employee_id = cw.employee_id
      AND (esa.contract_id IS NULL OR esa.contract_id = cw.contract_id)
      AND esa.effective_from <= cw.active_until
      AND (esa.effective_to IS NULL OR esa.effective_to > cw.active_from)
    ORDER BY
      (esa.contract_id = cw.contract_id) DESC,
      esa.effective_from DESC,
      esa.created_at DESC
    LIMIT 1
) latest_salary ON TRUE
JOIN cao_salary_scale_steps css ON css.id = latest_salary.salary_scale_step_id
JOIN cao_salary_tables cst ON cst.id = css.salary_table_id
WHERE cw.active_until >= cw.active_from
ORDER BY cw.employee_id ASC, cw.active_from ASC, cw.contract_id ASC;

-- name: ListOnCallPayrollMonthEmployeesPaginated :many
WITH on_call_month_employees AS (
    SELECT DISTINCT pp.employee_id
    FROM pay_periods pp
    JOIN pay_period_line_items ppl ON ppl.pay_period_id = pp.id
    WHERE pp.period_start = sqlc.arg('month_start')
      AND pp.period_end = sqlc.arg('month_end')
      AND ppl.contract_type = 'on_call'::employee_contract_type_enum

    UNION

    SELECT DISTINCT s.employee_id
    FROM schedules s
    JOIN LATERAL (
        SELECT c.id, c.contract_type
        FROM employee_contracts c
        WHERE c.employee_id = s.employee_id
          AND c.start_date <= DATE(s.start_datetime)
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= DATE(s.start_datetime))
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= DATE(s.start_datetime))
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    WHERE DATE(s.start_datetime) >= sqlc.arg('month_start')
      AND DATE(s.start_datetime) <= sqlc.arg('month_end')
      AND cc.contract_type = 'on_call'::employee_contract_type_enum
      AND EXISTS (
          SELECT 1
          FROM employee_salary_assignments esa
          WHERE esa.employee_id = s.employee_id
            AND (esa.contract_id IS NULL OR esa.contract_id = cc.id)
            AND esa.effective_from <= DATE(s.start_datetime)
            AND (esa.effective_to IS NULL OR esa.effective_to > DATE(s.start_datetime))
      )

    UNION

    SELECT DISTINCT oe.employee_id
    FROM overtime_entries oe
    JOIN LATERAL (
        SELECT c.id, c.contract_type
        FROM employee_contracts c
        WHERE c.employee_id = oe.employee_id
          AND c.start_date <= oe.entry_date
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= oe.entry_date)
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= oe.entry_date)
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    WHERE oe.entry_date >= sqlc.arg('month_start')
      AND oe.entry_date <= sqlc.arg('month_end')
      AND oe.status IN ('approved'::overtime_status_enum, 'submitted'::overtime_status_enum)
      AND cc.contract_type = 'on_call'::employee_contract_type_enum
      AND (
          oe.status = 'submitted'::overtime_status_enum
          OR EXISTS (
              SELECT 1
              FROM employee_salary_assignments esa
              WHERE esa.employee_id = oe.employee_id
                AND (esa.contract_id IS NULL OR esa.contract_id = cc.id)
                AND esa.effective_from <= oe.entry_date
                AND (esa.effective_to IS NULL OR esa.effective_to > oe.entry_date)
          )
      )

    UNION

    SELECT DISTINCT lpr.employee_id
    FROM leave_payout_requests lpr
    JOIN LATERAL (
        SELECT c.contract_type
        FROM employee_contracts c
        WHERE c.employee_id = lpr.employee_id
          AND c.start_date <= lpr.pay_period_start
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= lpr.pay_period_start)
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= lpr.pay_period_start)
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    WHERE lpr.pay_period_start >= sqlc.arg('month_start')
      AND lpr.pay_period_start <= sqlc.arg('month_end')
      AND lpr.status = 'approved'::payout_request_status_enum
      AND lpr.paid_period_id IS NULL
      AND cc.contract_type = 'on_call'::employee_contract_type_enum
)
SELECT
    ep.id AS employee_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM on_call_month_employees ocme
JOIN employee_profile ep ON ep.id = ocme.employee_id
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

-- name: ListPayPeriodsByEmployeeIDsAndRange :many
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
WHERE pp.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND pp.period_start = sqlc.arg('month_start')
  AND pp.period_end = sqlc.arg('month_end')
ORDER BY pp.employee_id ASC, pp.created_at DESC;

-- name: ListLockedPayPeriodMultiplierSummaries :many
SELECT
    ppl.pay_period_id,
    ppl.applied_rate_percent,
    COALESCE(SUM(CASE WHEN ppl.line_type = 'leave_payout' THEN 0 ELSE ppl.minutes_worked END), 0)::double precision AS worked_minutes,
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
        NULL::uuid AS leave_payout_request_id,
        cc.contract_type,
        css.hourly_rate::double precision AS contract_rate,
        NULL::double precision AS gross_amount_override,
        'none'::text AS irregular_hours_profile
    FROM schedules s
    JOIN employee_profile ep ON ep.id = s.employee_id
    JOIN LATERAL (
        SELECT c.id, c.contract_type
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
        NULL::uuid AS leave_payout_request_id,
        cc.contract_type,
        css.hourly_rate::double precision AS contract_rate,
        NULL::double precision AS gross_amount_override,
        'none'::text AS irregular_hours_profile
    FROM overtime_entries oe
    JOIN employee_profile ep ON ep.id = oe.employee_id
    JOIN LATERAL (
        SELECT c.id, c.contract_type
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
),
leave_payout_items AS (
    SELECT
        lpr.id AS source_id,
        lpr.employee_id,
        ep.first_name AS employee_first_name,
        ep.last_name AS employee_last_name,
        'Leave payout'::text AS label,
        lpr.pay_period_start AS work_date,
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
          AND c.start_date <= lpr.pay_period_start
          AND (c.effective_end_date IS NULL OR c.effective_end_date >= lpr.pay_period_start)
          AND (c.contract_end_date IS NULL OR c.contract_end_date >= lpr.pay_period_start)
        ORDER BY c.start_date DESC, c.created_at DESC
        LIMIT 1
    ) cc ON TRUE
    WHERE lpr.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
      AND lpr.status = 'approved'::payout_request_status_enum
      AND lpr.paid_period_id IS NULL
      AND lpr.pay_period_start >= sqlc.arg('month_start')
      AND lpr.pay_period_start <= sqlc.arg('month_end')
)
SELECT * FROM schedule_items
UNION ALL
SELECT * FROM overtime_items
UNION ALL
SELECT * FROM leave_payout_items
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
