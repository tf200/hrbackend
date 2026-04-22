-- name: ListPayrollMonthEmployeesPaginated :many
WITH month_employees AS (
    SELECT DISTINCT pp.employee_id
    FROM pay_periods pp
    WHERE pp.period_start = sqlc.arg('month_start')
      AND pp.period_end = sqlc.arg('month_end')

    UNION

    SELECT DISTINCT te.employee_id
    FROM time_entries te
    WHERE te.entry_date >= sqlc.arg('month_start')
      AND te.entry_date <= sqlc.arg('month_end')
      AND te.hour_type IN (
          'normal'::time_entry_hour_type_enum,
          'overtime'::time_entry_hour_type_enum,
          'travel'::time_entry_hour_type_enum,
          'training'::time_entry_hour_type_enum
      )
      AND te.status IN (
          'approved'::time_entry_status_enum,
          'draft'::time_entry_status_enum,
          'submitted'::time_entry_status_enum
      )
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

    SELECT DISTINCT te.employee_id
    FROM time_entries te
    WHERE te.entry_date >= sqlc.arg('month_start')
      AND te.entry_date <= sqlc.arg('month_end')
      AND te.hour_type IN (
          'normal'::time_entry_hour_type_enum,
          'overtime'::time_entry_hour_type_enum,
          'travel'::time_entry_hour_type_enum,
          'training'::time_entry_hour_type_enum
      )
      AND te.status IN (
          'approved'::time_entry_status_enum,
          'draft'::time_entry_status_enum,
          'submitted'::time_entry_status_enum
      )
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

-- name: ListPayrollMonthApprovedTimeEntriesByEmployeeIDs :many
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
WHERE te.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND te.status = 'approved'::time_entry_status_enum
  AND te.hour_type IN (
      'normal'::time_entry_hour_type_enum,
      'overtime'::time_entry_hour_type_enum,
      'travel'::time_entry_hour_type_enum,
      'training'::time_entry_hour_type_enum
  )
  AND te.entry_date >= sqlc.arg('month_start')
  AND te.entry_date <= sqlc.arg('month_end')
ORDER BY te.employee_id ASC, te.entry_date ASC, te.start_time ASC, te.created_at ASC;

-- name: ListPayrollMonthPendingSummariesByEmployeeIDs :many
SELECT
    te.employee_id,
    COUNT(*)::INT AS pending_entry_count,
    COALESCE(
        SUM(
            GREATEST(
                0,
                (
                    CASE
                        WHEN te.end_time > te.start_time THEN
                            EXTRACT(EPOCH FROM te.end_time) - EXTRACT(EPOCH FROM te.start_time)
                        ELSE
                            EXTRACT(EPOCH FROM te.end_time) + 86400 - EXTRACT(EPOCH FROM te.start_time)
                    END
                ) / 60 - te.break_minutes
            )
        ),
        0
    )::INT AS pending_worked_minutes
FROM time_entries te
WHERE te.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND te.status IN (
      'draft'::time_entry_status_enum,
      'submitted'::time_entry_status_enum
  )
  AND te.hour_type IN (
      'normal'::time_entry_hour_type_enum,
      'overtime'::time_entry_hour_type_enum,
      'travel'::time_entry_hour_type_enum,
      'training'::time_entry_hour_type_enum
  )
  AND te.entry_date >= sqlc.arg('month_start')
  AND te.entry_date <= sqlc.arg('month_end')
GROUP BY te.employee_id
ORDER BY te.employee_id ASC;

-- name: ListPayrollMonthPendingEntriesByEmployeeIDs :many
SELECT
    te.employee_id,
    GREATEST(
        0,
        (
            CASE
                WHEN te.end_time > te.start_time THEN
                    EXTRACT(EPOCH FROM te.end_time) - EXTRACT(EPOCH FROM te.start_time)
                ELSE
                    EXTRACT(EPOCH FROM te.end_time) + 86400 - EXTRACT(EPOCH FROM te.start_time)
            END
        ) / 60 - te.break_minutes
    )::INT AS worked_minutes,
    COALESCE(cc.contract_type, ep.contract_type) AS contract_type
FROM time_entries te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN LATERAL (
    SELECT
        c.contract_type
    FROM employee_contract_changes c
    WHERE c.employee_id = te.employee_id
      AND c.effective_from <= te.entry_date
    ORDER BY c.effective_from DESC, c.created_at DESC
    LIMIT 1
) cc ON TRUE
WHERE te.employee_id = ANY(sqlc.arg('employee_ids')::uuid[])
  AND te.status IN (
      'draft'::time_entry_status_enum,
      'submitted'::time_entry_status_enum
  )
  AND te.hour_type IN (
      'normal'::time_entry_hour_type_enum,
      'overtime'::time_entry_hour_type_enum,
      'travel'::time_entry_hour_type_enum,
      'training'::time_entry_hour_type_enum
  )
  AND te.entry_date >= sqlc.arg('month_start')
  AND te.entry_date <= sqlc.arg('month_end')
ORDER BY te.employee_id ASC, te.entry_date ASC, te.start_time ASC, te.created_at ASC;
