/*
 *  RBAC – Role & Permission Management
 *  This file contains all sqlc queries for the simple role-based
 *  access-control system.  They are grouped in the following order:
 *    1. Core roles
 *    2. Core permissions
 *    3. Role <-> Permission mapping
 *    4. User <-> Role mapping
 *    5. User <-> Permission mapping
 *    6. Read queries
 *    7. Helper utilities
 *
 *  sqlc will generate Go types and funcs named after the `-- name:` tags.
 */

/* ---------- 1. ROLES ---------- */

-- name: CreateRole :one
/* Insert a new role and return the created row. */
INSERT INTO roles (name, description)
VALUES (sqlc.arg('name'), sqlc.narg('description'))
RETURNING *;

-- name: GetAdminRoleId :one
/* Returns the ID of the admin role. */
SELECT id
FROM roles
WHERE name = 'admin';

-- name: GetRoleByID :one
/* Returns a single role by ID. */
SELECT *
FROM roles
WHERE id = $1;

-- name: ListRoles :many
/* Returns every role ordered by id with count of permissions and employees. */
SELECT 
    r.id,
    r.name,
    r.description,
    COALESCE(COUNT(DISTINCT rp.permission_id), 0)::BIGINT AS permission_count,
    COALESCE(COUNT(DISTINCT ep.id), 0)::BIGINT AS employee_count
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN user_roles ur ON r.id = ur.role_id
LEFT JOIN employee_profile ep ON ep.user_id = ur.user_id
GROUP BY r.id, r.name, r.description
ORDER BY r.id;

/* ---------- 2. PERMISSIONS ---------- */

-- name: ListAllPermissions :many
/* Returns every permission ordered by id. */
SELECT *
FROM permissions
ORDER BY group_key, section_key, sort_order, name;

/* ---------- 3. ROLE-PERMISSION MAPPING ---------- */

-- name: ListRolePermissions :many
/* Returns all permissions attached to a single role. */
SELECT p.id AS permission_id,
       p.name AS permission_name,
       p.resource AS permission_resource,
       p.method AS permission_method,
       p.group_key,
       p.section_key,
       p.display_name,
       p.description,
       p.sort_order
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.group_key, p.section_key, p.sort_order, p.name;


-- name: AddPermissionsToRole :exec
/* Bulk-insert permission IDs into a role (idempotent). */
INSERT INTO role_permissions (role_id, permission_id)
SELECT sqlc.arg('role_id'), unnest(sqlc.arg('permission_ids')::uuid[])
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: RemovePermissionsFromRole :exec
/* Removes *all* permissions from the given role. */
DELETE FROM role_permissions
WHERE role_id = $1;

/* ---------- 4. USER-ROLE MAPPING ---------- */

-- name: GetUserRoles :many
/* Returns every role granted to a user. */
SELECT r.id, r.name
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET role_id = $2;

/* ---------- 5. USER-PERMISSION OVERRIDES ---------- */

-- name: ListInheritedUserPermissions :many
/* Returns permissions inherited from the user's assigned role. */
SELECT p.id   AS permission_id,
       p.name AS permission_name,
       p.resource
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ur.user_id = $1
ORDER BY p.id;

-- name: ListUserPermissionOverrides :many
/* Returns explicit allow/deny overrides configured for a user. */
SELECT upo.permission_id,
       p.name AS permission_name,
       p.resource,
       upo.effect
FROM user_permission_overrides upo
JOIN permissions p ON p.id = upo.permission_id
WHERE upo.user_id = $1
ORDER BY p.id;

-- name: ListEffectiveUserPermissions :many
/* Returns effective permissions after applying role inheritance and overrides. */
WITH inherited_permissions AS (
    SELECT rp.permission_id
    FROM user_roles ur
    JOIN role_permissions rp ON rp.role_id = ur.role_id
    WHERE ur.user_id = $1
),
allowed_overrides AS (
    SELECT permission_id
    FROM user_permission_overrides
    WHERE user_id = $1
      AND effect = 'allow'
),
base_permissions AS (
    SELECT permission_id FROM inherited_permissions
    UNION
    SELECT permission_id FROM allowed_overrides
),
effective_permissions AS (
    SELECT permission_id
    FROM base_permissions
    EXCEPT
    SELECT permission_id
    FROM user_permission_overrides
    WHERE user_id = $1
      AND effect = 'deny'
)
SELECT p.id AS permission_id,
       p.name AS permission_name,
       p.resource
FROM effective_permissions ep
JOIN permissions p ON p.id = ep.permission_id
ORDER BY p.id;

-- name: AddUserPermissionOverrides :exec
/* Bulk-insert explicit overrides for a user (idempotent by replacement flow). */
INSERT INTO user_permission_overrides (user_id, permission_id, effect)
SELECT sqlc.arg('user_id'),
       unnest(sqlc.arg('permission_ids')::uuid[]),
       sqlc.arg('effect')
ON CONFLICT (user_id, permission_id) DO UPDATE SET effect = EXCLUDED.effect;

-- name: DeleteUserPermissionOverrides :exec
/* Removes all explicit overrides for the given user. */
DELETE FROM user_permission_overrides
WHERE user_id = $1;

/* ---------- 6. CHECK UTILITIES ---------- */

-- name: CheckUserPermission :one
/* Returns true/false whether the user has the named permission. */
WITH inherited_permissions AS (
    SELECT rp.permission_id
    FROM user_roles ur
    JOIN role_permissions rp ON rp.role_id = ur.role_id
    WHERE ur.user_id = $1
),
allowed_overrides AS (
    SELECT permission_id
    FROM user_permission_overrides
    WHERE user_id = $1
      AND effect = 'allow'
),
base_permissions AS (
    SELECT permission_id FROM inherited_permissions
    UNION
    SELECT permission_id FROM allowed_overrides
),
effective_permissions AS (
    SELECT permission_id
    FROM base_permissions
    EXCEPT
    SELECT permission_id
    FROM user_permission_overrides
    WHERE user_id = $1
      AND effect = 'deny'
)
SELECT EXISTS (
    SELECT 1
    FROM effective_permissions ep
    JOIN permissions p ON p.id = ep.permission_id
    WHERE p.name = $2
) AS has_permission;
