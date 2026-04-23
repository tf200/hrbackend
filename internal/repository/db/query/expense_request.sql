-- name: CreateExpenseRequest :one
INSERT INTO expense_requests (
    employee_id,
    created_by_employee_id,
    category,
    expense_date,
    merchant_name,
    description,
    business_purpose,
    currency,
    claimed_amount,
    travel_mode,
    travel_from,
    travel_to,
    distance_km,
    request_note,
    requested_at
) VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('created_by_employee_id'),
    sqlc.arg('category'),
    sqlc.arg('expense_date'),
    sqlc.narg('merchant_name'),
    sqlc.arg('description'),
    sqlc.arg('business_purpose'),
    sqlc.arg('currency'),
    sqlc.arg('claimed_amount'),
    sqlc.narg('travel_mode'),
    sqlc.narg('travel_from'),
    sqlc.narg('travel_to'),
    sqlc.narg('distance_km'),
    sqlc.narg('request_note'),
    NOW()
)
RETURNING *;

-- name: GetExpenseRequestByID :one
SELECT
    er.id,
    er.employee_id,
    er.created_by_employee_id,
    er.category,
    er.expense_date,
    er.merchant_name,
    er.description,
    er.business_purpose,
    er.currency,
    er.claimed_amount,
    er.approved_amount,
    er.travel_mode,
    er.travel_from,
    er.travel_to,
    er.distance_km,
    er.status,
    er.request_note,
    er.decision_note,
    er.decided_by_employee_id,
    er.reimbursed_by_employee_id,
    er.requested_at,
    er.decided_at,
    er.reimbursed_at,
    er.cancelled_at,
    er.created_at,
    er.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name
FROM expense_requests er
JOIN employee_profile ep ON ep.id = er.employee_id
WHERE er.id = sqlc.arg('id');

-- name: ListMyExpenseRequestsPaginated :many
SELECT
    er.id,
    er.employee_id,
    er.created_by_employee_id,
    er.category,
    er.expense_date,
    er.merchant_name,
    er.description,
    er.business_purpose,
    er.currency,
    er.claimed_amount,
    er.approved_amount,
    er.travel_mode,
    er.travel_from,
    er.travel_to,
    er.distance_km,
    er.status,
    er.request_note,
    er.decision_note,
    er.decided_by_employee_id,
    er.reimbursed_by_employee_id,
    er.requested_at,
    er.decided_at,
    er.reimbursed_at,
    er.cancelled_at,
    er.created_at,
    er.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM expense_requests er
JOIN employee_profile ep ON ep.id = er.employee_id
WHERE er.employee_id = sqlc.arg('employee_id')
  AND (
    sqlc.narg('status')::expense_request_status_enum IS NULL
    OR er.status = sqlc.narg('status')::expense_request_status_enum
  )
  AND (
    sqlc.narg('category')::expense_request_category_enum IS NULL
    OR er.category = sqlc.narg('category')::expense_request_category_enum
  )
ORDER BY er.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListExpenseRequestsPaginated :many
SELECT
    er.id,
    er.employee_id,
    er.created_by_employee_id,
    er.category,
    er.expense_date,
    er.merchant_name,
    er.description,
    er.business_purpose,
    er.currency,
    er.claimed_amount,
    er.approved_amount,
    er.travel_mode,
    er.travel_from,
    er.travel_to,
    er.distance_km,
    er.status,
    er.request_note,
    er.decision_note,
    er.decided_by_employee_id,
    er.reimbursed_by_employee_id,
    er.requested_at,
    er.decided_at,
    er.reimbursed_at,
    er.cancelled_at,
    er.created_at,
    er.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    COUNT(*) OVER() AS total_count
FROM expense_requests er
JOIN employee_profile ep ON ep.id = er.employee_id
WHERE (
    sqlc.narg('status')::expense_request_status_enum IS NULL
    OR er.status = sqlc.narg('status')::expense_request_status_enum
)
  AND (
    sqlc.narg('category')::expense_request_category_enum IS NULL
    OR er.category = sqlc.narg('category')::expense_request_category_enum
  )
  AND (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
  )
ORDER BY er.requested_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: LockExpenseRequestByID :one
SELECT *
FROM expense_requests
WHERE id = $1
FOR UPDATE;

-- name: UpdateExpenseRequestEditableFields :one
UPDATE expense_requests
SET
    category = COALESCE(sqlc.narg('category')::expense_request_category_enum, category),
    expense_date = COALESCE(sqlc.narg('expense_date')::date, expense_date),
    merchant_name = COALESCE(sqlc.narg('merchant_name')::text, merchant_name),
    description = COALESCE(sqlc.narg('description')::text, description),
    business_purpose = COALESCE(sqlc.narg('business_purpose')::text, business_purpose),
    currency = COALESCE(sqlc.narg('currency')::bpchar, currency),
    claimed_amount = COALESCE(sqlc.narg('claimed_amount')::numeric, claimed_amount),
    travel_mode = COALESCE(sqlc.narg('travel_mode')::text, travel_mode),
    travel_from = COALESCE(sqlc.narg('travel_from')::text, travel_from),
    travel_to = COALESCE(sqlc.narg('travel_to')::text, travel_to),
    distance_km = COALESCE(sqlc.narg('distance_km')::numeric, distance_km),
    request_note = COALESCE(sqlc.narg('request_note')::text, request_note),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: ApproveExpenseRequest :one
UPDATE expense_requests
SET
    status = 'approved'::expense_request_status_enum,
    approved_amount = sqlc.arg('approved_amount')::numeric,
    decision_note = sqlc.narg('decision_note')::text,
    decided_by_employee_id = sqlc.arg('decided_by_employee_id'),
    decided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: RejectExpenseRequest :one
UPDATE expense_requests
SET
    status = 'rejected'::expense_request_status_enum,
    decision_note = sqlc.narg('decision_note')::text,
    decided_by_employee_id = sqlc.arg('decided_by_employee_id'),
    decided_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: MarkExpenseRequestReimbursed :one
UPDATE expense_requests
SET
    status = 'reimbursed'::expense_request_status_enum,
    reimbursed_by_employee_id = sqlc.arg('reimbursed_by_employee_id'),
    reimbursed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: CancelExpenseRequest :one
UPDATE expense_requests
SET
    status = 'cancelled'::expense_request_status_enum,
    cancelled_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
