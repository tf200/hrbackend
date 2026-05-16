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
    COUNT(*) FILTER (WHERE contract_type IS DISTINCT FROM 'ZZP') AS total_employees,
    COUNT(*) FILTER (WHERE contract_type = 'ZZP') AS total_subcontractors,
    COUNT(*) FILTER (WHERE is_archived = TRUE) AS total_archived,
    COUNT(*) FILTER (WHERE out_of_service = TRUE) AS total_out_of_service
FROM
    employee_profile;

-- name: GetEmployeeDetailStats :one
WITH ranges AS (
    SELECT
        date_trunc('month', CURRENT_DATE)::date AS month_start,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month')::date AS next_month_start,
        date_trunc('year', CURRENT_DATE)::date AS year_start,
        (date_trunc('year', CURRENT_DATE) + INTERVAL '1 year')::date AS next_year_start,
        EXTRACT(YEAR FROM CURRENT_DATE)::int AS balance_year
), time_entry_minutes AS (
    SELECT
        COALESCE(SUM(worked_minutes) FILTER (
            WHERE te.status = 'approved'::time_entry_status_enum
              AND te.entry_date >= r.month_start
              AND te.entry_date < r.next_month_start
        ), 0)::double precision AS approved_month_minutes,
        COALESCE(SUM(worked_minutes) FILTER (
            WHERE te.status = 'submitted'::time_entry_status_enum
              AND te.entry_date >= r.month_start
              AND te.entry_date < r.next_month_start
        ), 0)::double precision AS pending_month_minutes,
        COALESCE(SUM(worked_minutes) FILTER (
            WHERE te.status = 'approved'::time_entry_status_enum
              AND te.entry_date >= r.year_start
              AND te.entry_date < r.next_year_start
        ), 0)::double precision AS approved_year_minutes
    FROM ranges r
    LEFT JOIN LATERAL (
        SELECT
            te.status,
            te.entry_date,
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
            ) AS worked_minutes
        FROM time_entries te
        WHERE te.employee_id = sqlc.arg('employee_id')
          AND te.hour_type IN (
              'normal'::time_entry_hour_type_enum,
              'overtime'::time_entry_hour_type_enum,
              'travel'::time_entry_hour_type_enum,
              'training'::time_entry_hour_type_enum
          )
          AND te.status IN (
              'approved'::time_entry_status_enum,
              'submitted'::time_entry_status_enum
          )
          AND te.entry_date >= r.year_start
          AND te.entry_date < r.next_year_start
    ) te ON TRUE
), leave_balance AS (
    SELECT
        COALESCE(
            (lb.legal_total_hours - lb.legal_used_hours)
            + (lb.extra_total_hours - lb.extra_used_hours),
            0
        )::int AS remaining_leave_balance_hours
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
    lb.remaining_leave_balance_hours,
    (tem.approved_month_minutes / 60.0)::double precision AS hours_worked_this_month,
    (tem.pending_month_minutes / 60.0)::double precision AS hours_pending_approval,
    (tem.approved_year_minutes / 60.0)::double precision AS total_hours_worked_this_year,
    lr.last_performance_review_score
FROM leave_balance lb
CROSS JOIN time_entry_minutes tem
LEFT JOIN last_review lr ON TRUE;

-- name: ListEmployeesWithContractHours :many
SELECT
    id,
    first_name,
    last_name,
    contract_hours
FROM employee_profile
WHERE id = ANY($1::uuid[])
AND contract_hours IS NOT NULL
AND contract_hours > 0;

-- name: ListEmployeeNamesByIDs :many
SELECT
    id,
    first_name,
    last_name
FROM employee_profile
WHERE id = ANY(sqlc.arg(employee_ids)::uuid[]);
