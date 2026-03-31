-- name: CreateShiftSwapRequest :one
INSERT INTO shift_swap_requests (
    requester_employee_id,
    recipient_employee_id,
    requester_schedule_id,
    recipient_schedule_id,
    status,
    requested_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), $6
)
RETURNING *;

-- name: ExpirePendingShiftSwapRequests :exec
UPDATE shift_swap_requests
SET
    status = 'expired',
    updated_at = NOW()
WHERE status IN ('pending_recipient', 'pending_admin')
  AND expires_at IS NOT NULL
  AND expires_at <= NOW();

-- name: GetScheduleForSwapValidation :one
SELECT
    id,
    employee_id,
    location_id,
    start_datetime,
    end_datetime
FROM schedules
WHERE id = $1
LIMIT 1;

-- name: GetShiftSwapRequestByID :one
SELECT *
FROM shift_swap_requests
WHERE id = $1
LIMIT 1;

-- name: UpdateShiftSwapStatusAfterRecipientResponse :one
UPDATE shift_swap_requests
SET
    status = $1,
    recipient_responded_at = NOW(),
    recipient_response_note = $2,
    updated_at = NOW()
WHERE id = $3
  AND recipient_employee_id = $4
  AND status = 'pending_recipient'
  AND (expires_at IS NULL OR expires_at > NOW())
RETURNING *;

-- name: UpdateShiftSwapAdminDecision :one
UPDATE shift_swap_requests
SET
    status = $1,
    admin_decided_at = NOW(),
    admin_decision_note = $2,
    admin_employee_id = $3,
    updated_at = NOW()
WHERE id = $4
  AND status = 'pending_admin'
  AND (expires_at IS NULL OR expires_at > NOW())
RETURNING *;

-- name: GetShiftSwapRequestDetailsByID :one
SELECT
    ssr.id,
    ssr.requester_employee_id,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    ssr.recipient_employee_id,
    rec.first_name AS recipient_first_name,
    rec.last_name AS recipient_last_name,
    ssr.requester_schedule_id,
    req_s.start_datetime AS requester_schedule_start_datetime,
    req_s.end_datetime AS requester_schedule_end_datetime,
    ssr.recipient_schedule_id,
    rec_s.start_datetime AS recipient_schedule_start_datetime,
    rec_s.end_datetime AS recipient_schedule_end_datetime,
    (
        CASE
        WHEN ssr.status IN ('pending_recipient', 'pending_admin')
         AND ssr.expires_at IS NOT NULL
         AND ssr.expires_at <= NOW()
        THEN 'expired'::shift_swap_status_enum
        ELSE ssr.status
        END
    )::shift_swap_status_enum AS status,
    ssr.requested_at,
    ssr.recipient_responded_at,
    ssr.admin_decided_at,
    ssr.recipient_response_note,
    ssr.admin_decision_note,
    ssr.admin_employee_id,
    admin_ep.first_name AS admin_first_name,
    admin_ep.last_name AS admin_last_name,
    ssr.expires_at
FROM shift_swap_requests ssr
JOIN employee_profile req ON req.id = ssr.requester_employee_id
JOIN employee_profile rec ON rec.id = ssr.recipient_employee_id
JOIN schedules req_s ON req_s.id = ssr.requester_schedule_id
JOIN schedules rec_s ON rec_s.id = ssr.recipient_schedule_id
LEFT JOIN employee_profile admin_ep ON admin_ep.id = ssr.admin_employee_id
WHERE ssr.id = $1
LIMIT 1;

-- name: LockShiftSwapRequestForAdminDecision :one
SELECT *
FROM shift_swap_requests
WHERE id = $1
FOR UPDATE;

-- name: LockSchedulesByIDsForSwap :many
SELECT
    id,
    employee_id,
    start_datetime,
    end_datetime
FROM schedules
WHERE id = ANY($1::uuid[])
ORDER BY id
FOR UPDATE;

-- name: CountScheduleOverlapsForEmployee :one
SELECT COUNT(*)::bigint
FROM schedules s
WHERE s.employee_id = sqlc.arg(employee_id)
  AND s.id <> ALL(sqlc.arg(excluded_schedule_ids)::uuid[])
  AND s.start_datetime < sqlc.arg(conflict_end)
  AND s.end_datetime > sqlc.arg(conflict_start);

-- name: UpdateScheduleEmployeeAssignment :exec
UPDATE schedules
SET
    employee_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkShiftSwapConfirmed :one
UPDATE shift_swap_requests
SET
    status = 'confirmed',
    admin_decided_at = NOW(),
    admin_decision_note = $2,
    admin_employee_id = $3,
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending_admin'
RETURNING *;

-- name: ListMyShiftSwapRequests :many
SELECT
    ssr.id,
    ssr.requester_employee_id,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    ssr.recipient_employee_id,
    rec.first_name AS recipient_first_name,
    rec.last_name AS recipient_last_name,
    ssr.requester_schedule_id,
    req_s.start_datetime AS requester_schedule_start_datetime,
    req_s.end_datetime AS requester_schedule_end_datetime,
    ssr.recipient_schedule_id,
    rec_s.start_datetime AS recipient_schedule_start_datetime,
    rec_s.end_datetime AS recipient_schedule_end_datetime,
    (
        CASE
        WHEN ssr.status IN ('pending_recipient', 'pending_admin')
         AND ssr.expires_at IS NOT NULL
         AND ssr.expires_at <= NOW()
        THEN 'expired'::shift_swap_status_enum
        ELSE ssr.status
        END
    )::shift_swap_status_enum AS status,
    ssr.requested_at,
    ssr.recipient_responded_at,
    ssr.admin_decided_at,
    ssr.recipient_response_note,
    ssr.admin_decision_note,
    ssr.admin_employee_id,
    admin_ep.first_name AS admin_first_name,
    admin_ep.last_name AS admin_last_name,
    ssr.expires_at
FROM shift_swap_requests ssr
JOIN employee_profile req ON req.id = ssr.requester_employee_id
JOIN employee_profile rec ON rec.id = ssr.recipient_employee_id
JOIN schedules req_s ON req_s.id = ssr.requester_schedule_id
JOIN schedules rec_s ON rec_s.id = ssr.recipient_schedule_id
LEFT JOIN employee_profile admin_ep ON admin_ep.id = ssr.admin_employee_id
WHERE ssr.requester_employee_id = $1
   OR ssr.recipient_employee_id = $1
ORDER BY ssr.requested_at DESC;

-- name: ListShiftSwapRequestsPaginated :many
SELECT
    ssr.id,
    ssr.requester_employee_id,
    req.first_name AS requester_first_name,
    req.last_name AS requester_last_name,
    ssr.recipient_employee_id,
    rec.first_name AS recipient_first_name,
    rec.last_name AS recipient_last_name,
    ssr.requester_schedule_id,
    req_s.start_datetime AS requester_schedule_start_datetime,
    req_s.end_datetime AS requester_schedule_end_datetime,
    ssr.recipient_schedule_id,
    rec_s.start_datetime AS recipient_schedule_start_datetime,
    rec_s.end_datetime AS recipient_schedule_end_datetime,
    (
        CASE
        WHEN ssr.status IN ('pending_recipient', 'pending_admin')
         AND ssr.expires_at IS NOT NULL
         AND ssr.expires_at <= NOW()
        THEN 'expired'::shift_swap_status_enum
        ELSE ssr.status
        END
    )::shift_swap_status_enum AS status,
    ssr.requested_at,
    ssr.recipient_responded_at,
    ssr.admin_decided_at,
    ssr.recipient_response_note,
    ssr.admin_decision_note,
    ssr.admin_employee_id,
    admin_ep.first_name AS admin_first_name,
    admin_ep.last_name AS admin_last_name,
    ssr.expires_at,
    COUNT(*) OVER() AS total_count
FROM shift_swap_requests ssr
JOIN employee_profile req ON req.id = ssr.requester_employee_id
JOIN employee_profile rec ON rec.id = ssr.recipient_employee_id
JOIN schedules req_s ON req_s.id = ssr.requester_schedule_id
JOIN schedules rec_s ON rec_s.id = ssr.recipient_schedule_id
LEFT JOIN employee_profile admin_ep ON admin_ep.id = ssr.admin_employee_id
WHERE (
    sqlc.narg('status')::shift_swap_status_enum IS NULL
    OR (
        (
            CASE
            WHEN ssr.status IN ('pending_recipient', 'pending_admin')
             AND ssr.expires_at IS NOT NULL
             AND ssr.expires_at <= NOW()
            THEN 'expired'::shift_swap_status_enum
            ELSE ssr.status
            END
        )::shift_swap_status_enum
    ) = sqlc.narg('status')::shift_swap_status_enum
)
  AND (
    sqlc.narg('employee_id')::uuid IS NULL
    OR ssr.requester_employee_id = sqlc.narg('employee_id')::uuid
    OR ssr.recipient_employee_id = sqlc.narg('employee_id')::uuid
  )
ORDER BY ssr.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
