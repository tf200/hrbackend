-- name: GetEmployeeAvailabilityStats :one
WITH today_schedules AS (
    SELECT DISTINCT employee_id
    FROM schedules
    WHERE start_datetime::date <= CURRENT_DATE
      AND end_datetime::date >= CURRENT_DATE
),
today_sick_leaves AS (
    SELECT DISTINCT employee_id
    FROM leave_requests
    WHERE leave_type = 'sick'
      AND start_date <= CURRENT_DATE
      AND end_date >= CURRENT_DATE
      AND status != 'cancelled'
),
today_approved_leaves AS (
    SELECT DISTINCT employee_id
    FROM leave_requests
    WHERE leave_type IN ('vacation', 'personal', 'unpaid', 'pregnancy', 'other')
      AND status = 'approved'
      AND start_date <= CURRENT_DATE
      AND end_date >= CURRENT_DATE
),
classified AS (
    SELECT
        CASE
            WHEN ep.out_of_service = TRUE THEN 'out_of_service'
            WHEN tsl.employee_id IS NOT NULL THEN 'sick'
            WHEN tal.employee_id IS NOT NULL THEN 'on_leave'
            WHEN ts.employee_id IS NOT NULL THEN 'active'
            ELSE 'unscheduled'
        END AS category
    FROM employee_profile ep
    LEFT JOIN today_schedules ts ON ts.employee_id = ep.id
    LEFT JOIN today_sick_leaves tsl ON tsl.employee_id = ep.id
    LEFT JOIN today_approved_leaves tal ON tal.employee_id = ep.id
    WHERE ep.is_archived = FALSE
)
SELECT
    COUNT(*) FILTER (WHERE category = 'active')::bigint AS active,
    COUNT(*) FILTER (WHERE category = 'on_leave')::bigint AS on_leave,
    COUNT(*) FILTER (WHERE category = 'sick')::bigint AS sick,
    COUNT(*) FILTER (WHERE category = 'out_of_service')::bigint AS out_of_service,
    COUNT(*) FILTER (WHERE category = 'unscheduled')::bigint AS unscheduled
FROM classified;

-- name: ListOpenShiftCoverage :many
WITH date_range AS (
    SELECT generate_series(
        CURRENT_DATE,
        CURRENT_DATE + ((sqlc.arg(days)::int - 1) * INTERVAL '1 day'),
        INTERVAL '1 day'
    )::date AS shift_date
), expected_shifts AS (
    SELECT
        l.id AS location_id,
        l.name AS location_name,
        l.street,
        l.house_number,
        l.house_number_addition,
        l.postal_code,
        l.city,
        l.timezone,
        ls.id AS shift_id,
        ls.shift_name,
        dr.shift_date
    FROM location l
    JOIN location_shift ls ON ls.location_id = l.id
    CROSS JOIN date_range dr
), uncovered_shifts AS (
    SELECT es.*
    FROM expected_shifts es
    WHERE NOT EXISTS (
        SELECT 1
        FROM schedules s
        WHERE s.location_id = es.location_id
          AND s.location_shift_id = es.shift_id
          AND DATE(s.start_datetime AT TIME ZONE es.timezone) = es.shift_date
    )
)
SELECT
    location_id,
    location_name,
    street,
    house_number,
    house_number_addition,
    postal_code,
    city,
    shift_id,
    shift_name,
    shift_date,
    1::bigint AS open_slots
FROM uncovered_shifts
ORDER BY location_name, shift_date, shift_name;

-- name: GetCriticalActionStats :one
SELECT
    COALESCE(lr.pending, 0)::bigint AS pending_leave_requests,
    COALESCE(ss.pending_admin, 0)::bigint AS pending_shift_swaps,
    COALESCE(er.pending, 0)::bigint AS pending_expense_claims,
    COALESCE(te.submitted, 0)::bigint AS pending_time_entries
FROM (
    SELECT COUNT(*)::bigint AS pending
    FROM leave_requests
    WHERE status = 'pending'::leave_request_status_enum
) lr,
(
    SELECT COUNT(*)::bigint AS pending_admin
    FROM shift_swap_requests
    WHERE status = 'pending_admin'::shift_swap_status_enum
      AND (expires_at IS NULL OR expires_at > NOW())
) ss,
(
    SELECT COUNT(*)::bigint AS pending
    FROM expense_requests
    WHERE status = 'pending'::expense_request_status_enum
) er,
(
    SELECT COUNT(*)::bigint AS submitted
    FROM time_entries
    WHERE status = 'submitted'::time_entry_status_enum
) te;

-- name: GetPayrollTotalStats :one
WITH current_month AS (
    SELECT
        date_trunc('month', CURRENT_DATE)::date AS month_start,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::date AS month_end
), line_totals AS (
    SELECT
        ppl.contract_type,
        COALESCE(SUM(ppl.base_amount), 0)::double precision AS base_amount,
        COALESCE(SUM(ppl.premium_amount), 0)::double precision AS premium_amount
    FROM current_month cm
    JOIN pay_periods pp
      ON pp.period_start = cm.month_start
     AND pp.period_end = cm.month_end
    JOIN pay_period_line_items ppl ON ppl.pay_period_id = pp.id
    GROUP BY ppl.contract_type
)
SELECT
    cm.month_start,
    COALESCE(
        SUM(base_amount) FILTER (WHERE contract_type = 'loondienst'::employee_contract_type_enum),
        0
    )::double precision AS salary_total,
    COALESCE(
        SUM(base_amount + premium_amount) FILTER (WHERE contract_type = 'ZZP'::employee_contract_type_enum),
        0
    )::double precision AS zzp_total,
    COALESCE(SUM(premium_amount), 0)::double precision AS ort_total
FROM current_month cm
LEFT JOIN line_totals lt ON TRUE
GROUP BY cm.month_start;

-- name: GetRiskRadarStats :one
WITH current_month AS (
    SELECT
        date_trunc('month', CURRENT_DATE)::date AS month_start,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month - 1 day')::date AS month_end
)
SELECT
    cm.month_start,
    COALESCE(contracts_ending.total, 0)::bigint AS contracts_ending_this_month,
    COALESCE(overdue_training.total, 0)::bigint AS overdue_training,
    COALESCE(late_arrivals.total, 0)::bigint AS late_arrivals_this_month
FROM current_month cm,
(
    SELECT COUNT(*)::bigint AS total
    FROM current_month cm2
    JOIN employee_profile ep
      ON ep.contract_end_date >= cm2.month_start
     AND ep.contract_end_date <= cm2.month_end
    WHERE ep.is_archived = FALSE
) contracts_ending,
(
    SELECT COUNT(*)::bigint AS total
    FROM employee_training_assignments eta
    WHERE eta.due_at IS NOT NULL
      AND eta.due_at < NOW()
      AND eta.status NOT IN (
          'completed'::training_assignment_status_enum,
          'cancelled'::training_assignment_status_enum
      )
) overdue_training,
(
    SELECT COUNT(*)::bigint AS total
    FROM current_month cm2
    JOIN late_arrivals lr
      ON lr.arrival_date >= cm2.month_start
     AND lr.arrival_date <= cm2.month_end
) late_arrivals;

-- name: ListTeamHealthByDepartment :many
WITH active_employees AS (
    SELECT
        ep.id,
        ep.department_id
    FROM employee_profile ep
    WHERE ep.is_archived = FALSE
      AND COALESCE(ep.out_of_service, FALSE) = FALSE
      AND ep.department_id IS NOT NULL
), today_schedules AS (
    SELECT DISTINCT s.employee_id
    FROM schedules s
    WHERE s.start_datetime::date <= CURRENT_DATE
      AND s.end_datetime::date >= CURRENT_DATE
), today_absences AS (
    SELECT DISTINCT lr.employee_id
    FROM leave_requests lr
    WHERE lr.start_date <= CURRENT_DATE
      AND lr.end_date >= CURRENT_DATE
      AND lr.status != 'cancelled'::leave_request_status_enum
      AND (
          lr.leave_type = 'sick'::leave_request_type_enum
          OR (
              lr.leave_type IN (
                  'vacation'::leave_request_type_enum,
                  'personal'::leave_request_type_enum,
                  'unpaid'::leave_request_type_enum,
                  'pregnancy'::leave_request_type_enum,
                  'other'::leave_request_type_enum
              )
              AND lr.status = 'approved'::leave_request_status_enum
          )
      )
), department_employee_stats AS (
    SELECT
        ae.department_id,
        COUNT(*)::double precision AS active_employee_count,
        COUNT(ts.employee_id)::double precision AS scheduled_today_count,
        COUNT(ta.employee_id)::double precision AS absent_today_count
    FROM active_employees ae
    LEFT JOIN today_schedules ts ON ts.employee_id = ae.id
    LEFT JOIN today_absences ta ON ta.employee_id = ae.id
    GROUP BY ae.department_id
), department_training_stats AS (
    SELECT
        ae.department_id,
        COUNT(eta.id) FILTER (
            WHERE eta.status != 'cancelled'::training_assignment_status_enum
        )::double precision AS training_assignment_count,
        COUNT(eta.id) FILTER (
            WHERE eta.status != 'cancelled'::training_assignment_status_enum
              AND eta.due_at IS NOT NULL
              AND eta.due_at < NOW()
              AND eta.status != 'completed'::training_assignment_status_enum
        )::double precision AS overdue_training_count
    FROM active_employees ae
    LEFT JOIN employee_training_assignments eta ON eta.employee_id = ae.id
    GROUP BY ae.department_id
), percentages AS (
    SELECT
        d.id AS department_id,
        d.name AS department_name,
        ROUND((des.scheduled_today_count / NULLIF(des.active_employee_count, 0) * 100)::numeric, 0)::double precision AS staffing_percent,
        ROUND((des.absent_today_count / NULLIF(des.active_employee_count, 0) * 100)::numeric, 0)::double precision AS absence_percent,
        CASE
            WHEN COALESCE(dts.training_assignment_count, 0) = 0 THEN 100::double precision
            ELSE ROUND(((dts.training_assignment_count - dts.overdue_training_count) / dts.training_assignment_count * 100)::numeric, 0)::double precision
        END AS training_percent
    FROM departments d
    JOIN department_employee_stats des ON des.department_id = d.id
    LEFT JOIN department_training_stats dts ON dts.department_id = d.id
), scored AS (
    SELECT
        p.department_id,
        p.department_name,
        p.staffing_percent,
        p.absence_percent,
        p.training_percent,
        ROUND((
            (
                p.staffing_percent * 0.40 +
                GREATEST(0::double precision, 100::double precision - (p.absence_percent * 2)) * 0.30 +
                p.training_percent * 0.30
            ) / 10
        )::numeric, 1)::double precision AS score
    FROM percentages p
)
SELECT
    department_id,
    department_name,
    staffing_percent,
    absence_percent,
    training_percent,
    score,
    CASE
        WHEN score >= 8.0 THEN 'low'
        WHEN score >= 7.5 THEN 'medium'
        ELSE 'high'
    END::text AS risk
FROM scored
ORDER BY score ASC, department_name ASC;
