-- name: GetUserIDByEmployeeID :one
SELECT user_id FROM employee_profile
WHERE id = $1 LIMIT 1;

-- name: ListUserIDsByEmployeeIDs :many
SELECT user_id
FROM employee_profile
WHERE id = ANY($1::uuid[]);

-- name: SetEmployeeProfilePicture :one
UPDATE custom_user
SET profile_picture = $2
WHERE id = (
    SELECT user_id
    FROM employee_profile
    WHERE employee_profile.id = $1
)
RETURNING *;

-- name: SearchEmployeesByNameOrEmail :many
SELECT
    id,
    first_name,
    last_name,
    work_email_address
FROM employee_profile
WHERE
    first_name ILIKE '%' || @search || '%' OR
    last_name ILIKE '%' || @search || '%' OR
    email ILIKE '%' || @search || '%'
LIMIT 10;

-- name: GetEmployeeCounts :one
SELECT
    COUNT(*) FILTER (WHERE ec.contract_type = 'permanent') AS total_permanent,
    COUNT(*) FILTER (WHERE ec.contract_type = 'temporary') AS total_temporary,
    COUNT(*) FILTER (WHERE ec.contract_type = 'on_call') AS total_on_call,
    COUNT(*) FILTER (WHERE ep.out_of_service = TRUE) AS total_out_of_service
FROM employee_profile ep
LEFT JOIN LATERAL (
    SELECT c.contract_type
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
      AND c.start_date <= CURRENT_DATE
      AND (c.effective_end_date IS NULL OR c.effective_end_date >= CURRENT_DATE)
      AND (c.contract_end_date IS NULL OR c.contract_end_date >= CURRENT_DATE)
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON TRUE;

-- name: GetEmployeeDetailStats :one
WITH ranges AS (
    SELECT
        date_trunc('month', CURRENT_DATE)::date AS month_start,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month')::date AS next_month_start,
        date_trunc('year', CURRENT_DATE)::date AS year_start,
        (date_trunc('year', CURRENT_DATE) + INTERVAL '1 year')::date AS next_year_start,
        EXTRACT(YEAR FROM CURRENT_DATE)::int AS balance_year
), overtime_stats AS (
    SELECT
        COALESCE(SUM(oe.minutes) FILTER (
            WHERE oe.status = 'approved'::overtime_status_enum
              AND oe.entry_date >= r.month_start
              AND oe.entry_date < r.next_month_start
        ), 0)::double precision AS approved_month_minutes,
        COALESCE(SUM(oe.minutes) FILTER (
            WHERE oe.status = 'submitted'::overtime_status_enum
              AND oe.entry_date >= r.month_start
              AND oe.entry_date < r.next_month_start
        ), 0)::double precision AS pending_month_minutes,
        COALESCE(SUM(oe.minutes) FILTER (
            WHERE oe.status = 'approved'::overtime_status_enum
              AND oe.entry_date >= r.year_start
              AND oe.entry_date < r.next_year_start
        ), 0)::double precision AS approved_year_minutes
    FROM ranges r
    LEFT JOIN overtime_entries oe ON oe.employee_id = sqlc.arg('employee_id')
      AND oe.entry_date >= r.year_start
      AND oe.entry_date < r.next_year_start
), schedule_stats AS (
    SELECT
        COALESCE(SUM(
            EXTRACT(EPOCH FROM (s.end_datetime - s.start_datetime)) / 60
        ) FILTER (
            WHERE DATE(s.start_datetime) >= r.month_start
              AND DATE(s.start_datetime) < r.next_month_start
        ), 0)::double precision AS scheduled_month_minutes,
        COALESCE(SUM(
            EXTRACT(EPOCH FROM (s.end_datetime - s.start_datetime)) / 60
        ) FILTER (
            WHERE DATE(s.start_datetime) >= r.year_start
              AND DATE(s.start_datetime) < r.next_year_start
        ), 0)::double precision AS scheduled_year_minutes
    FROM ranges r
    LEFT JOIN schedules s ON s.employee_id = sqlc.arg('employee_id')
      AND DATE(s.start_datetime) >= r.year_start
      AND DATE(s.start_datetime) < r.next_year_start
), leave_balance AS (
    SELECT
        COALESCE(
            (
                calculate_legal_leave_minutes(
                    sqlc.arg('employee_id'),
                    r.balance_year,
                    CURRENT_TIMESTAMP
                ) + COALESCE(lb.legal_adjustment_minutes, 0) - COALESCE(lb.legal_used_minutes, 0)
            )
            + (COALESCE(lb.extra_total_minutes, 0) - COALESCE(lb.extra_used_minutes, 0)),
            0
        )::int AS remaining_leave_balance_minutes
    FROM ranges r
    LEFT JOIN leave_balances lb ON lb.employee_id = sqlc.arg('employee_id')
      AND lb.year = r.balance_year
), last_review AS (
    SELECT pa.total_score::double precision AS last_performance_review_score
    FROM performance_assessments pa
    WHERE pa.employee_id = sqlc.arg('employee_id')
      AND pa.status = 'completed'::performance_assessment_status_enum
    ORDER BY pa.assessment_date DESC, pa.created_at DESC
    LIMIT 1
)
SELECT
    lb.remaining_leave_balance_minutes,
    ((os.approved_month_minutes + ss.scheduled_month_minutes) / 60.0)::double precision AS hours_worked_this_month,
    (os.pending_month_minutes / 60.0)::double precision AS hours_pending_approval,
    ((os.approved_year_minutes + ss.scheduled_year_minutes) / 60.0)::double precision AS total_hours_worked_this_year,
    lr.last_performance_review_score
FROM leave_balance lb
CROSS JOIN overtime_stats os
CROSS JOIN schedule_stats ss
LEFT JOIN last_review lr ON TRUE;

-- name: ListEmployeesWithContractHours :many
SELECT
    ep.id,
    ep.first_name,
    ep.last_name,
    ec.hours_per_week AS contract_hours
FROM employee_profile ep
JOIN LATERAL (
    SELECT hours_per_week
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
WHERE ep.id = ANY($1::uuid[])
AND ec.hours_per_week IS NOT NULL
AND ec.hours_per_week > 0;

-- name: ListEmployeeNamesByIDs :many
SELECT
    id,
    first_name,
    last_name
FROM employee_profile
WHERE id = ANY(sqlc.arg(employee_ids)::uuid[]);
