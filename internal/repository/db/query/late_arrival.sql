-- name: ListAssignedSchedulesForEmployeeOnDate :many
SELECT
    s.id AS schedule_id,
    s.employee_id,
    s.start_datetime,
    s.end_datetime,
    l.timezone AS location_timezone,
    l.name AS location_name,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name
FROM schedules s
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE s.employee_id = sqlc.arg(employee_id)
  AND DATE(s.start_datetime AT TIME ZONE l.timezone) = sqlc.arg(arrival_date)::date
ORDER BY s.start_datetime;

-- name: CreateLateArrival :one
INSERT INTO late_arrivals (
    schedule_id,
    employee_id,
    created_by_employee_id,
    arrival_date,
    arrival_time,
    reason
) VALUES (
    sqlc.arg(schedule_id),
    sqlc.arg(employee_id),
    sqlc.arg(created_by_employee_id),
    sqlc.arg(arrival_date),
    sqlc.arg(arrival_time),
    sqlc.arg(reason)
)
RETURNING *;

-- name: ListMyLateArrivalsPaginated :many
SELECT
    lr.id,
    lr.schedule_id,
    lr.employee_id,
    lr.created_by_employee_id,
    lr.arrival_date,
    lr.arrival_time,
    lr.reason,
    lr.created_at,
    lr.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    s.start_datetime AS shift_start_datetime,
    s.end_datetime AS shift_end_datetime,
    l.name AS location_name,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name,
    COUNT(*) OVER() AS total_count
FROM late_arrivals lr
JOIN employee_profile ep ON ep.id = lr.employee_id
JOIN schedules s ON s.id = lr.schedule_id
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE lr.employee_id = sqlc.arg(employee_id)
  AND (
    sqlc.narg('date_from')::date IS NULL
    OR lr.arrival_date >= sqlc.narg('date_from')::date
  )
  AND (
    sqlc.narg('date_to')::date IS NULL
    OR lr.arrival_date <= sqlc.narg('date_to')::date
  )
ORDER BY lr.arrival_date DESC, lr.arrival_time DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListLateArrivalsPaginated :many
SELECT
    lr.id,
    lr.schedule_id,
    lr.employee_id,
    lr.created_by_employee_id,
    lr.arrival_date,
    lr.arrival_time,
    lr.reason,
    lr.created_at,
    lr.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    s.start_datetime AS shift_start_datetime,
    s.end_datetime AS shift_end_datetime,
    l.name AS location_name,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name,
    COUNT(*) OVER() AS total_count
FROM late_arrivals lr
JOIN employee_profile ep ON ep.id = lr.employee_id
JOIN schedules s ON s.id = lr.schedule_id
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
  )
  AND (
    sqlc.narg('date_from')::date IS NULL
    OR lr.arrival_date >= sqlc.narg('date_from')::date
  )
  AND (
    sqlc.narg('date_to')::date IS NULL
    OR lr.arrival_date <= sqlc.narg('date_to')::date
  )
ORDER BY lr.arrival_date DESC, lr.arrival_time DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
