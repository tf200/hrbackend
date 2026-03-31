
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
