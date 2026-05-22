-- name: ListPayrollPreviewTimeEntries :many
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
    cc.contract_type,
    NULL::numeric AS contract_rate,
    NULL::text AS irregular_hours_profile
FROM time_entries te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN LATERAL (
    SELECT
        c.contract_type,
        c.contract_end_date
    FROM employee_contracts c
    WHERE c.employee_id = te.employee_id
      AND c.start_date <= te.entry_date
      AND (c.contract_end_date IS NULL OR c.contract_end_date >= te.entry_date)
    ORDER BY c.start_date DESC, c.created_at DESC
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
ORDER BY te.entry_date ASC, te.start_time ASC, te.created_at ASC;

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
