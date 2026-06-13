-- name: GetEmployeePayoutContract :one
WITH latest_contract AS (
    SELECT
        ec.id,
        ec.contract_type
    FROM employee_contracts ec
    WHERE ec.employee_id = $1
    ORDER BY ec.start_date DESC, ec.created_at DESC
    LIMIT 1
)
SELECT
    lc.contract_type,
    css.hourly_rate::double precision AS contract_rate
FROM latest_contract lc
JOIN LATERAL (
    SELECT esa.salary_scale_step_id
    FROM employee_salary_assignments esa
    WHERE esa.employee_id = $1
      AND (esa.contract_id IS NULL OR esa.contract_id = lc.id)
      AND esa.effective_from <= CURRENT_DATE
      AND (esa.effective_to IS NULL OR esa.effective_to > CURRENT_DATE)
    ORDER BY
      (esa.contract_id = lc.id) DESC,
      esa.effective_from DESC,
      esa.created_at DESC
    LIMIT 1
) latest_salary ON TRUE
JOIN cao_salary_scale_steps css ON css.id = latest_salary.salary_scale_step_id;

-- name: ComputeReservedPayoutMinutesForYear :one
SELECT COALESCE(SUM(requested_hours * 60), 0)::int AS reserved_payout_minutes
FROM leave_payout_requests
WHERE employee_id = sqlc.arg('employee_id')
  AND balance_year = sqlc.arg('balance_year')
  AND status IN (
    'pending'::payout_request_status_enum,
    'approved'::payout_request_status_enum,
    'paid'::payout_request_status_enum
  );

-- name: CreatePayoutRequest :one
INSERT INTO leave_payout_requests (
    employee_id,
    created_by_employee_id,
    requested_hours,
    balance_year,
    hourly_rate,
    gross_amount,
    request_note,
    requested_at
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('created_by_employee_id'),
    sqlc.arg('requested_hours'),
    sqlc.arg('balance_year'),
    sqlc.arg('hourly_rate'),
    sqlc.arg('gross_amount'),
    sqlc.narg('request_note'),
    NOW()
)
RETURNING *;

-- name: ListMyPayoutRequestsPaginated :many
SELECT
    pr.id,
    pr.employee_id,
    pr.created_by_employee_id,
    pr.requested_hours,
    pr.balance_year,
    pr.hourly_rate,
    pr.gross_amount,
    pr.pay_period_start,
    pr.status,
    pr.request_note,
    pr.decision_note,
    pr.decided_by_employee_id,
    pr.paid_by_employee_id,
    pr.requested_at,
    pr.decided_at,
    pr.paid_at,
    pr.created_at,
    pr.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM leave_payout_requests pr
JOIN employee_profile ep ON ep.id = pr.employee_id
WHERE pr.employee_id = sqlc.arg('employee_id')
  AND (
    sqlc.narg('status')::payout_request_status_enum IS NULL
    OR pr.status = sqlc.narg('status')::payout_request_status_enum
  )
ORDER BY pr.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListPayoutRequestsPaginated :many
SELECT
    pr.id,
    pr.employee_id,
    pr.created_by_employee_id,
    pr.requested_hours,
    pr.balance_year,
    pr.hourly_rate,
    pr.gross_amount,
    pr.pay_period_start,
    pr.status,
    pr.request_note,
    pr.decision_note,
    pr.decided_by_employee_id,
    pr.paid_by_employee_id,
    pr.requested_at,
    pr.decided_at,
    pr.paid_at,
    pr.created_at,
    pr.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM leave_payout_requests pr
JOIN employee_profile ep ON ep.id = pr.employee_id
WHERE (
    sqlc.narg('status')::payout_request_status_enum IS NULL
    OR pr.status = sqlc.narg('status')::payout_request_status_enum
)
  AND (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
  )
ORDER BY pr.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: LockPayoutRequestByID :one
SELECT *
FROM leave_payout_requests
WHERE id = $1
FOR UPDATE;

-- name: ApprovePayoutRequest :one
UPDATE leave_payout_requests
SET
    status = 'approved'::payout_request_status_enum,
    decision_note = sqlc.narg('decision_note')::text,
    decided_by_employee_id = sqlc.arg('decided_by_employee_id'),
    pay_period_start = sqlc.arg('pay_period_start')::date,
    decided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: RejectPayoutRequest :one
UPDATE leave_payout_requests
SET
    status = 'rejected'::payout_request_status_enum,
    decision_note = sqlc.narg('decision_note')::text,
    decided_by_employee_id = sqlc.arg('decided_by_employee_id'),
    decided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: MarkPayoutRequestPaid :one
UPDATE leave_payout_requests
SET
    status = 'paid'::payout_request_status_enum,
    paid_by_employee_id = sqlc.arg('paid_by_employee_id'),
    paid_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdatePayoutRequest :one
UPDATE leave_payout_requests
SET
    requested_hours = sqlc.arg('requested_hours'),
    balance_year = sqlc.arg('balance_year'),
    gross_amount = sqlc.arg('gross_amount'),
    request_note = sqlc.narg('request_note'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
