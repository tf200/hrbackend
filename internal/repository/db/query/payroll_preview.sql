-- name: ListPayrollPreviewWorkItems :many
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
      AND DATE(s.start_datetime) <= sqlc.arg(period_end)
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
      AND oe.entry_date <= sqlc.arg(period_end)
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
    WHERE lpr.employee_id = sqlc.arg(employee_id)
      AND lpr.status = 'approved'::payout_request_status_enum
      AND lpr.paid_period_id IS NULL
      AND lpr.pay_period_start = sqlc.arg(period_start)
)
SELECT * FROM schedule_items
UNION ALL
SELECT * FROM overtime_items
UNION ALL
SELECT * FROM leave_payout_items
ORDER BY work_date ASC, source_type ASC, source_id ASC;

-- name: ListNationalHolidaysInRange :many
SELECT
    holiday_date,
    name
FROM national_holidays
WHERE country_code = sqlc.arg(country_code)
  AND is_national = TRUE
  AND holiday_date >= sqlc.arg(start_date)
  AND holiday_date <= sqlc.arg(end_date)
ORDER BY holiday_date ASC;
