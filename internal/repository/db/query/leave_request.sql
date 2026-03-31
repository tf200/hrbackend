-- name: CreateLeaveRequest :one
INSERT INTO leave_requests (
    employee_id,
    created_by_employee_id,
    leave_type,
    start_date,
    end_date,
    reason,
    requested_at
) VALUES (
    sqlc.arg(employee_id),
    sqlc.arg(created_by_employee_id),
    sqlc.arg(leave_type),
    sqlc.arg(start_date),
    sqlc.arg(end_date),
    sqlc.narg(reason),
    NOW()
)
RETURNING *;

-- name: ListMyLeaveRequestsPaginated :many
SELECT
    lr.id,
    lr.employee_id,
    lr.created_by_employee_id,
    lr.leave_type,
    lr.status,
    lr.start_date,
    lr.end_date,
    lr.reason,
    lr.decision_note,
    lr.decided_by_employee_id,
    lr.requested_at,
    lr.decided_at,
    lr.cancelled_at,
    lr.created_at,
    lr.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM leave_requests lr
JOIN employee_profile ep ON ep.id = lr.employee_id
WHERE lr.employee_id = sqlc.arg(employee_id)
  AND (
    sqlc.narg('status')::leave_request_status_enum IS NULL
    OR lr.status = sqlc.narg('status')::leave_request_status_enum
  )
ORDER BY lr.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetMyLeaveRequestStats :one
SELECT
    COUNT(*) FILTER (
        WHERE lr.status = 'pending'::leave_request_status_enum
    )::BIGINT AS open_requests,
    COUNT(*) FILTER (
        WHERE lr.status = 'approved'::leave_request_status_enum
    )::BIGINT AS approved_requests,
    COUNT(*) FILTER (
        WHERE lr.status = 'rejected'::leave_request_status_enum
    )::BIGINT AS rejected_requests,
    COUNT(*) FILTER (
        WHERE lr.leave_type = 'sick'::leave_request_type_enum
          AND lr.status = 'approved'::leave_request_status_enum
    )::BIGINT AS sickness_absence
FROM leave_requests lr
WHERE lr.employee_id = sqlc.arg(employee_id)
  AND lr.start_date < (DATE_TRUNC('year', NOW()) + INTERVAL '1 year')::date
  AND lr.end_date >= DATE_TRUNC('year', NOW())::date;

-- name: GetLeaveRequestStats :one
SELECT
    COUNT(*) FILTER (
        WHERE lr.status = 'pending'::leave_request_status_enum
    )::BIGINT AS open_requests,
    COUNT(*) FILTER (
        WHERE lr.status = 'approved'::leave_request_status_enum
    )::BIGINT AS approved_requests,
    COUNT(*) FILTER (
        WHERE lr.status = 'rejected'::leave_request_status_enum
    )::BIGINT AS rejected_requests,
    COUNT(*) FILTER (
        WHERE lr.leave_type = 'sick'::leave_request_type_enum
          AND lr.status = 'approved'::leave_request_status_enum
    )::BIGINT AS sickness_absence
FROM leave_requests lr
WHERE lr.start_date < (DATE_TRUNC('year', NOW()) + INTERVAL '1 year')::date
  AND lr.end_date >= DATE_TRUNC('year', NOW())::date;

-- name: ListLeaveRequestsPaginated :many
SELECT
    lr.id,
    lr.employee_id,
    lr.created_by_employee_id,
    lr.leave_type,
    lr.status,
    lr.start_date,
    lr.end_date,
    lr.reason,
    lr.decision_note,
    lr.decided_by_employee_id,
    lr.requested_at,
    lr.decided_at,
    lr.cancelled_at,
    lr.created_at,
    lr.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM leave_requests lr
JOIN employee_profile ep ON ep.id = lr.employee_id
WHERE (
    sqlc.narg('status')::leave_request_status_enum IS NULL
    OR lr.status = sqlc.narg('status')::leave_request_status_enum
)
  AND (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
  )
ORDER BY lr.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: LockLeaveRequestByID :one
SELECT *
FROM leave_requests
WHERE id = $1
FOR UPDATE;

-- name: UpdateLeaveRequestEditableFields :one
UPDATE leave_requests
SET
    leave_type = COALESCE(sqlc.narg('leave_type')::leave_request_type_enum, leave_type),
    start_date = COALESCE(sqlc.narg('start_date')::date, start_date),
    end_date = COALESCE(sqlc.narg('end_date')::date, end_date),
    reason = COALESCE(sqlc.narg('reason')::text, reason),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateLeaveRequestDecision :one
UPDATE leave_requests
SET
    status = sqlc.arg('status')::leave_request_status_enum,
    decision_note = sqlc.narg('decision_note')::text,
    decided_by_employee_id = sqlc.arg(decided_by_employee_id),
    decided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
