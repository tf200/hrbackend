-- name: CreateNotifications :many
INSERT INTO notifications (user_id, type, message, data, created_at)
SELECT
    unnest(sqlc.arg('user_ids')::uuid[]),
    sqlc.arg('type'),
    sqlc.arg('message'),
    sqlc.arg('data')::jsonb,
    sqlc.arg('created_at')
RETURNING *;

-- name: ListNotificationUserIDsByEmployeeIDs :many
SELECT DISTINCT ep.user_id
FROM employee_profile ep
JOIN custom_user cu ON cu.id = ep.user_id
WHERE ep.id = ANY(sqlc.arg('employee_ids')::uuid[])
AND cu.is_active = TRUE
AND ep.is_archived = FALSE;

-- name: ListNotificationUserIDsByRoles :many
SELECT DISTINCT cu.id
FROM custom_user cu
JOIN user_roles ur ON ur.user_id = cu.id
JOIN roles r ON r.id = ur.role_id
WHERE r.name = ANY(sqlc.arg('role_names')::text[])
AND cu.is_active = TRUE;

-- name: ListNotificationUserIDsByPermissions :many
WITH inherited_permissions AS (
    SELECT ur.user_id
    FROM user_roles ur
    JOIN role_permissions rp ON rp.role_id = ur.role_id
    JOIN permissions p ON p.id = rp.permission_id
    WHERE p.name = ANY(sqlc.arg('permission_names')::text[])
),
allowed_overrides AS (
    SELECT upo.user_id
    FROM user_permission_overrides upo
    JOIN permissions p ON p.id = upo.permission_id
    WHERE p.name = ANY(sqlc.arg('permission_names')::text[])
    AND upo.effect = 'allow'
),
denied_overrides AS (
    SELECT upo.user_id
    FROM user_permission_overrides upo
    JOIN permissions p ON p.id = upo.permission_id
    WHERE p.name = ANY(sqlc.arg('permission_names')::text[])
    AND upo.effect = 'deny'
),
effective_users AS (
    SELECT user_id FROM inherited_permissions
    UNION
    SELECT user_id FROM allowed_overrides
    EXCEPT
    SELECT user_id FROM denied_overrides
)
SELECT DISTINCT cu.id
FROM effective_users eu
JOIN custom_user cu ON cu.id = eu.user_id
WHERE cu.is_active = TRUE;
