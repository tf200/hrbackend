-- name: GetActiveLeavePolicyByType :one
SELECT
    leave_type,
    requires_approval,
    deducts_balance,
    is_active,
    created_at,
    updated_at
FROM leave_policies
WHERE leave_type = sqlc.arg('leave_type')::leave_request_type_enum
  AND is_active = TRUE;
