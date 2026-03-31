-- name: CreateUser :one
INSERT INTO custom_user (
    password,
    email,
    is_active,
    profile_picture
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByID :one
SELECT cu.*, e.id as employee_id FROM custom_user cu
JOIN employee_profile e ON e.user_id = cu.id
WHERE cu.id = $1 LIMIT 1;


-- name: GetUserByEmail :one
SELECT cu.*, e.id as employee_id FROM custom_user cu
JOIN employee_profile e ON e.user_id = cu.id
WHERE cu.email = $1 LIMIT 1;

-- name: UpdatePassword :exec
UPDATE custom_user
SET password = $2
WHERE id = $1;


-- name: CreateTemp2FaSecret :execrows
UPDATE custom_user
SET two_factor_secret_temp = $2
WHERE id = $1;

-- name: GetTemp2FaSecret :one
SELECT two_factor_secret_temp FROM custom_user
WHERE id = $1 LIMIT 1;


-- name: Enable2Fa :execrows
UPDATE custom_user
SET two_factor_secret = $2,
    two_factor_secret_temp = NULL,
    two_factor_enabled = true,
    recovery_codes = $3
WHERE id = $1;


-- name: GetAllAdminUsers :many
SELECT * FROM custom_user
WHERE role_id = (SELECT id FROM roles WHERE name = 'admin')
;