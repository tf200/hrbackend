
-- name: CreateSchedule :one
WITH inserted_schedule AS (
    INSERT INTO schedules (
        employee_id,
        location_id,
        location_shift_id,
        shift_name_snapshot,
        shift_start_time_snapshot,
        shift_end_time_snapshot,
        is_custom,
        created_by_employee_id,
        start_datetime,
        end_datetime
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
    )
    RETURNING *
)
SELECT 
    s.*,
    l.name as location_name
FROM inserted_schedule s
JOIN location l ON s.location_id = l.id;


-- name: GetSchedulesByLocationInRange :many
SELECT 
  s.id AS shift_id,
  s.employee_id,
  s.location_id,
  s.start_datetime,
  s.end_datetime,
  s.is_custom,
  COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name,
  ls.id AS location_shift_id,
  DATE(s.start_datetime) AS day,
  e.first_name AS employee_first_name,
  e.last_name AS employee_last_name
FROM schedules s
LEFT JOIN location_shift ls 
  ON s.location_shift_id = ls.id
JOIN employee_profile e 
  ON s.employee_id = e.id
WHERE s.location_id = sqlc.arg(location_id)
  AND DATE(s.start_datetime) BETWEEN sqlc.arg(start_date)::date AND sqlc.arg(end_date)::date
ORDER BY day, s.start_datetime;

-- name: GetEmployeeSchedulesByDay :many
SELECT
    s.id AS schedule_id,
    s.employee_id,
    s.location_id,
    l.name AS location_name,
    s.start_datetime,
    s.end_datetime,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name,
    ls.id AS location_shift_id,
    s.is_custom
FROM schedules s
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE s.employee_id = sqlc.arg(employee_id)
  AND DATE(s.start_datetime AT TIME ZONE l.timezone) = sqlc.arg(date)::date
ORDER BY s.start_datetime;


-- name: GetScheduleById :one
SELECT s.*,
    e.first_name AS employee_first_name,
    e.last_name AS employee_last_name,
    l.name AS location_name,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS location_shift_name,
    ls.id AS location_shift_id
FROM schedules s
LEFT JOIN location_shift ls ON s.location_shift_id = ls.id
JOIN employee_profile e ON s.employee_id = e.id
JOIN location l ON s.location_id = l.id
WHERE s.id = $1
LIMIT 1;

-- name: UpdateSchedule :one
WITH updated_schedule AS (
    UPDATE schedules
    SET
        employee_id = $2,
        location_id = $3,
        location_shift_id = $4,
        start_datetime = $5,
        end_datetime = $6,
        shift_name_snapshot = $7,
        shift_start_time_snapshot = $8,
        shift_end_time_snapshot = $9,
        is_custom = $10,
        updated_at = NOW()
    WHERE schedules.id = $1
    RETURNING *
)
SELECT 
    s.*,
    l.name as location_name
FROM updated_schedule s
JOIN location l ON s.location_id = l.id;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE id = $1;



-- name: GetEmployeeSchedules :many
SELECT 
    s.id,
    s.start_datetime,
    s.end_datetime,
    s.location_id,
    l.name as location_name,
    'shift'::text as type
FROM schedules s
JOIN location l ON s.location_id = l.id
WHERE s.employee_id = @employee_id
    AND s.start_datetime >= @period_start
    AND s.start_datetime < @period_end
ORDER BY s.start_datetime;

-- name: GetEmployeeNextShift :one
SELECT
    s.id AS schedule_id,
    s.location_id,
    l.name AS location_name,
    l.street,
    l.house_number,
    l.house_number_addition,
    l.postal_code,
    l.city,
    s.start_datetime,
    s.end_datetime,
    s.is_custom,
    s.shift_name_snapshot AS shift_name,
    DATE(s.start_datetime AT TIME ZONE l.timezone) AS shift_date
FROM schedules s
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE s.employee_id = @employee_id
  AND s.start_datetime > @now
ORDER BY s.start_datetime
LIMIT 1;

-- name: ListEmployeeShiftColleagues :many
SELECT
    colleague.id AS employee_id,
    colleague.first_name,
    colleague.last_name
FROM schedules base
JOIN schedules shared
  ON shared.location_id = base.location_id
 AND shared.start_datetime = base.start_datetime
 AND shared.end_datetime = base.end_datetime
 AND (
    shared.location_shift_id IS NOT DISTINCT FROM base.location_shift_id
 )
JOIN employee_profile colleague ON colleague.id = shared.employee_id
WHERE base.id = @schedule_id
  AND shared.employee_id <> @employee_id
ORDER BY colleague.first_name, colleague.last_name;

-- name: GetEmployeeShiftOverviewStats :one
SELECT
    COUNT(*) FILTER (
        WHERE start_datetime >= @month_start
          AND start_datetime < @month_end
          AND start_datetime > @now
    )::bigint AS upcoming_count,
    COUNT(*) FILTER (
        WHERE start_datetime >= @month_start
          AND start_datetime < @month_end
          AND end_datetime <= @now
    )::bigint AS completed_count,
    COALESCE(SUM(
        EXTRACT(EPOCH FROM (end_datetime - start_datetime)) / 3600.0
    ) FILTER (
        WHERE start_datetime >= @week_start
          AND start_datetime < @week_end
          AND start_datetime > @now
    ), 0)::float8 AS planned_hours
FROM schedules
WHERE employee_id = @employee_id
  AND start_datetime >= LEAST(@week_start, @month_start)
  AND start_datetime < GREATEST(@week_end, @month_end);

-- name: ListEmployeeWeekShiftCounts :many
SELECT
    DATE(s.start_datetime AT TIME ZONE l.timezone) AS day,
    COUNT(*)::bigint AS shift_count
FROM schedules s
JOIN location l ON l.id = s.location_id
WHERE s.employee_id = @employee_id
  AND s.start_datetime >= @week_start
  AND s.start_datetime < @week_end
GROUP BY day
ORDER BY day;

-- name: ListEmployeeMonthShiftCounts :many
SELECT
    EXTRACT(DAY FROM s.start_datetime AT TIME ZONE l.timezone)::int AS day,
    COUNT(*)::bigint AS shift_count
FROM schedules s
JOIN location l ON l.id = s.location_id
WHERE s.employee_id = @employee_id
  AND s.start_datetime >= @month_start
  AND s.start_datetime < @month_end
GROUP BY day
ORDER BY day;

-- name: GetEmployeeScheduleManager :one
SELECT
    mgr.first_name,
    mgr.last_name
FROM employee_profile employee
JOIN employee_profile mgr ON mgr.id = employee.manager_employee_id
WHERE employee.id = @employee_id
LIMIT 1;

-- name: ListEmployeeUpcomingShifts :many
SELECT
    s.id AS schedule_id,
    s.location_id,
    l.name AS location_name,
    l.street,
    l.house_number,
    l.house_number_addition,
    l.postal_code,
    l.city,
    s.start_datetime,
    s.end_datetime,
    s.is_custom,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name,
    DATE(s.start_datetime AT TIME ZONE l.timezone) AS shift_date
FROM schedules s
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE s.employee_id = @employee_id
  AND s.start_datetime > @now
  AND s.start_datetime < @window_end
ORDER BY s.start_datetime;

-- name: ListEmployeePastShiftsPaginated :many
SELECT
    s.id AS schedule_id,
    s.location_id,
    l.name AS location_name,
    l.street,
    l.house_number,
    l.house_number_addition,
    l.postal_code,
    l.city,
    s.start_datetime,
    s.end_datetime,
    s.is_custom,
    COALESCE(s.shift_name_snapshot, ls.shift_name, 'Custom Shift') AS shift_name,
    DATE(s.start_datetime AT TIME ZONE l.timezone) AS shift_date,
    COUNT(*) OVER() AS total_count
FROM schedules s
JOIN location l ON l.id = s.location_id
LEFT JOIN location_shift ls ON ls.id = s.location_shift_id
WHERE s.employee_id = @employee_id
  AND s.start_datetime < @now
ORDER BY s.start_datetime DESC
LIMIT @limit_count OFFSET @offset_count;

-- name: ListShiftColleaguesByScheduleIDs :many
SELECT
    s.id AS schedule_id,
    colleague.id AS employee_id,
    colleague.first_name,
    colleague.last_name
FROM schedules s
JOIN schedules shared
  ON shared.location_id = s.location_id
 AND shared.start_datetime = s.start_datetime
 AND shared.end_datetime = s.end_datetime
 AND (shared.location_shift_id IS NOT DISTINCT FROM s.location_shift_id)
JOIN employee_profile colleague ON colleague.id = shared.employee_id
WHERE s.id = ANY(sqlc.arg(schedule_ids)::uuid[])
  AND shared.employee_id <> @employee_id
ORDER BY s.id, colleague.first_name, colleague.last_name;
