-- name: ComputeLegalLeaveTotalForYear :one
SELECT calculate_legal_leave_minutes(
    sqlc.arg('employee_id'),
    sqlc.arg('year')::int,
    sqlc.arg('as_of')::timestamptz
)::int AS legal_total_minutes;

-- name: LockEmployeeProfileForLeaveBalance :one
SELECT id
FROM employee_profile
WHERE id = sqlc.arg('employee_id')
FOR UPDATE;

-- name: ComputeLegalLeaveUsedForYear :one
SELECT COALESCE(SUM(lr.requested_minutes), 0)::int AS legal_used_minutes
FROM leave_requests lr
JOIN leave_policies lp ON lp.leave_type = lr.leave_type
WHERE lr.employee_id = sqlc.arg('employee_id')
  AND lr.status = 'approved'::leave_request_status_enum
  AND lp.deducts_balance = TRUE
  AND lr.start_date >= make_date(sqlc.arg('year')::int, 1, 1)
  AND lr.start_date < make_date(sqlc.arg('year')::int + 1, 1, 1);

-- name: GetEmployeeContractForLeave :one
SELECT
    hours_per_week AS contract_hours,
    contract_type
FROM employee_contracts
WHERE employee_id = sqlc.arg('employee_id')
ORDER BY start_date DESC, created_at DESC
LIMIT 1;

-- name: ListLeaveBalancesPaginated :many
WITH input AS (
    SELECT COALESCE(sqlc.narg('year')::int, EXTRACT(YEAR FROM CURRENT_DATE)::int)::int AS year
)
SELECT
    lb.id AS employee_id,
    input.year,
    ent.legal_calculated_minutes::int AS legal_total_minutes,
    used.legal_used_minutes::int AS legal_used_minutes,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    COUNT(*) OVER() AS total_count
FROM employee_profile lb
CROSS JOIN input
JOIN employee_profile ep ON ep.id = lb.id
JOIN LATERAL (
    SELECT calculate_legal_leave_minutes(
        lb.id,
        input.year,
        CASE
            WHEN input.year < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(input.year, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN input.year = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(input.year, 1, 1)::timestamptz
        END
    )::int AS legal_calculated_minutes
) ent ON true
JOIN LATERAL (
    SELECT COALESCE(SUM(lr.requested_minutes), 0)::int AS legal_used_minutes
    FROM leave_requests lr
    JOIN leave_policies lp ON lp.leave_type = lr.leave_type
    WHERE lr.employee_id = lb.id
      AND lr.status = 'approved'::leave_request_status_enum
      AND lp.deducts_balance = TRUE
      AND lr.start_date >= make_date(input.year, 1, 1)
      AND lr.start_date < make_date(input.year + 1, 1, 1)
) used ON true
LEFT JOIN LATERAL (
    SELECT *
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
WHERE (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
)
ORDER BY ep.first_name ASC, ep.last_name ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListMyLeaveBalancesPaginated :many
WITH input AS (
    SELECT COALESCE(sqlc.narg('year')::int, EXTRACT(YEAR FROM CURRENT_DATE)::int)::int AS year
)
SELECT
    lb.id AS employee_id,
    input.year::int AS year,
    ent.legal_calculated_minutes::int AS legal_total_minutes,
    used.legal_used_minutes::int AS legal_used_minutes,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    ec.effective_end_date,
    COUNT(*) OVER() AS total_count
FROM employee_profile lb
CROSS JOIN input
JOIN employee_profile ep ON ep.id = lb.id
JOIN LATERAL (
    SELECT calculate_legal_leave_minutes(
        lb.id,
        input.year,
        CASE
            WHEN input.year < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(input.year, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN input.year = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(input.year, 1, 1)::timestamptz
        END
    )::int AS legal_calculated_minutes
) ent ON true
JOIN LATERAL (
    SELECT COALESCE(SUM(lr.requested_minutes), 0)::int AS legal_used_minutes
    FROM leave_requests lr
    JOIN leave_policies lp ON lp.leave_type = lr.leave_type
    WHERE lr.employee_id = lb.id
      AND lr.status = 'approved'::leave_request_status_enum
      AND lp.deducts_balance = TRUE
      AND lr.start_date >= make_date(input.year, 1, 1)
      AND lr.start_date < make_date(input.year + 1, 1, 1)
) used ON true
LEFT JOIN LATERAL (
    SELECT *
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
WHERE lb.id = sqlc.arg('employee_id')
ORDER BY input.year DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetLeaveBalanceDetails :one
SELECT
    lb.id AS employee_id,
    sqlc.arg('year')::int AS year,
    ent.legal_calculated_minutes::int AS legal_total_minutes,
    used.legal_used_minutes::int AS legal_used_minutes,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ec.hours_per_week AS contract_hours,
    ec.contract_type,
    ec.start_date AS contract_start_date,
    ec.contract_end_date,
    ec.effective_end_date
FROM employee_profile lb
JOIN employee_profile ep ON ep.id = lb.id
JOIN LATERAL (
    SELECT calculate_legal_leave_minutes(
        lb.id,
        sqlc.arg('year')::int,
        CASE
            WHEN sqlc.arg('year')::int < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(sqlc.arg('year')::int, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN sqlc.arg('year')::int = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(sqlc.arg('year')::int, 1, 1)::timestamptz
        END
    )::int AS legal_calculated_minutes
) ent ON true
JOIN LATERAL (
    SELECT COALESCE(SUM(lr.requested_minutes), 0)::int AS legal_used_minutes
    FROM leave_requests lr
    JOIN leave_policies lp ON lp.leave_type = lr.leave_type
    WHERE lr.employee_id = lb.id
      AND lr.status = 'approved'::leave_request_status_enum
      AND lp.deducts_balance = TRUE
      AND lr.start_date >= make_date(sqlc.arg('year')::int, 1, 1)
      AND lr.start_date < make_date(sqlc.arg('year')::int + 1, 1, 1)
) used ON true
LEFT JOIN LATERAL (
    SELECT *
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
WHERE lb.id = sqlc.arg('employee_id')
LIMIT 1;

-- name: ListLeaveContractAccrualsForYear :many
WITH input AS (
    SELECT
        sqlc.arg('employee_id')::uuid AS employee_id,
        sqlc.arg('year')::int AS year,
        CASE
            WHEN sqlc.arg('year')::int < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(sqlc.arg('year')::int, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN sqlc.arg('year')::int = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(sqlc.arg('year')::int, 1, 1)::timestamptz
        END AS as_of
), segments AS (
    SELECT
        ec.id AS contract_id,
        ec.employee_id,
        ec.contract_type,
        ec.hours_per_week AS contract_hours,
        ec.start_date AS contract_start_date,
        ec.contract_end_date,
        ec.effective_end_date,
        GREATEST(ec.start_date, make_date(input.year, 1, 1)) AS segment_start_date,
        LEAST(
            COALESCE(ec.contract_end_date, make_date(input.year, 12, 31)),
            COALESCE(ec.effective_end_date, make_date(input.year, 12, 31)),
            COALESCE(
                (
                    LEAD(ec.start_date) OVER (
                        PARTITION BY ec.employee_id
                        ORDER BY ec.start_date
                    ) - INTERVAL '1 day'
                )::date,
                make_date(input.year, 12, 31)
            ),
            make_date(input.year, 12, 31),
            input.as_of::date
        ) AS segment_end_date,
        COALESCE(ec.hours_per_week, 0) AS weekly_hours,
        input.year,
        input.as_of
    FROM employee_contracts ec
    JOIN input ON input.employee_id = ec.employee_id
)
SELECT
    segments.contract_id,
    segments.contract_type,
    segments.contract_hours,
    segments.contract_start_date,
    segments.contract_end_date,
    segments.effective_end_date,
    segments.segment_start_date::date AS segment_start_date,
    segments.segment_end_date::date AS segment_end_date,
    (make_date(segments.year + 1, 1, 1) - make_date(segments.year, 1, 1))::int AS year_days,
    (segments.segment_end_date - segments.segment_start_date + 1)::int AS segment_days,
    CASE
        WHEN segments.contract_type = 'on_call' THEN 0
        ELSE ROUND(segments.weekly_hours * 60.0 * 4.0)::int
    END AS full_year_minutes,
    COALESCE(worked.schedule_minutes, 0)::int AS schedule_minutes,
    COALESCE(worked.overtime_minutes, 0)::int AS overtime_minutes,
    GREATEST(
        0,
        ROUND(
            CASE
                WHEN segments.contract_type = 'on_call' THEN
                    (COALESCE(worked.schedule_minutes, 0) + COALESCE(worked.overtime_minutes, 0)) / 13.0
                WHEN segments.weekly_hours <= 0 THEN
                    0
                ELSE
                    (segments.weekly_hours * 60.0 * 4.0) * (
                        (segments.segment_end_date - segments.segment_start_date + 1)::numeric /
                        (make_date(segments.year + 1, 1, 1) - make_date(segments.year, 1, 1))::numeric
                    )
            END
        )::int
    )::int AS gained_minutes
FROM segments
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM(
            EXTRACT(EPOCH FROM (s.end_datetime - s.start_datetime)) / 60
        ), 0)::numeric AS schedule_minutes,
        COALESCE((
            SELECT SUM(oe.minutes)::numeric
            FROM overtime_entries oe
            WHERE oe.employee_id = segments.employee_id
              AND oe.status = 'approved'
              AND oe.entry_date >= segments.segment_start_date
              AND oe.entry_date <= segments.segment_end_date
        ), 0)::numeric AS overtime_minutes
    FROM schedules s
    WHERE s.employee_id = segments.employee_id
      AND DATE(s.start_datetime) >= segments.segment_start_date
      AND DATE(s.start_datetime) <= segments.segment_end_date
      AND s.end_datetime <= segments.as_of
) worked ON segments.contract_type = 'on_call'
WHERE segments.segment_end_date >= segments.segment_start_date
ORDER BY segments.segment_start_date ASC, segments.contract_start_date ASC;
