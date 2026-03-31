-- name: CreateEmployeeProfile :one
INSERT INTO employee_profile (
    user_id,
    first_name,
    last_name,
    bsn,
    street,
    house_number,
    house_number_addition,
    postal_code,
    city,
    position,
    department_id,
    manager_employee_id,
    employee_number,
    employment_number,
    private_email_address,
    work_email_address,
    work_phone_number,
    private_phone_number,
    date_of_birth,
    home_telephone_number,
    gender,
    location_id,
    contract_hours,
    contract_end_date,
    contract_start_date,
    contract_type,
    contract_rate
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27
) RETURNING *;

-- name: ListEmployeeProfile :many
SELECT
    ep.id,
    ep.first_name,
    ep.last_name,
    ep.bsn,
    ep.contract_type,
    d.name AS department_name,
    ep.contract_end_date,
    concat_ws(' ', l.street, l.house_number, l.house_number_addition, l.postal_code, l.city) AS location_address
FROM employee_profile ep
LEFT JOIN location l ON l.id = ep.location_id
LEFT JOIN departments d ON d.id = ep.department_id
WHERE
    (CASE
        WHEN sqlc.narg('include_archived')::boolean IS NULL THEN true
        WHEN sqlc.narg('include_archived')::boolean = false THEN NOT ep.is_archived
        ELSE true
    END) AND
    (CASE
        WHEN sqlc.narg('include_out_of_service')::boolean IS NULL THEN true
        WHEN sqlc.narg('include_out_of_service')::boolean = false THEN NOT COALESCE(ep.out_of_service, false)
        ELSE true
    END) AND
    (ep.location_id = sqlc.narg('location_id') OR sqlc.narg('location_id') IS NULL) AND
    (ep.contract_type = sqlc.narg('contract_type') OR sqlc.narg('contract_type') IS NULL) AND
    (sqlc.narg('search')::TEXT IS NULL OR
        ep.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
        ep.last_name ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY ep.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountEmployeeProfile :one
SELECT COUNT(*)
FROM employee_profile ep
WHERE
    (CASE
        WHEN sqlc.narg('include_archived')::boolean IS NULL THEN true
        WHEN sqlc.narg('include_archived')::boolean = false THEN NOT ep.is_archived
        ELSE true
    END) AND
    (CASE
        WHEN sqlc.narg('include_out_of_service')::boolean IS NULL THEN true
        WHEN sqlc.narg('include_out_of_service')::boolean = false THEN NOT COALESCE(ep.out_of_service, false)
        ELSE true
    END) AND
    (location_id = sqlc.narg('location_id') OR sqlc.narg('location_id') IS NULL) AND
    (contract_type = sqlc.narg('contract_type') OR sqlc.narg('contract_type') IS NULL);

-- name: GetEmployeeProfileByUserID :one
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
SELECT
    cu.id           AS user_id,
    cu.email        AS email,
    cu.last_login   AS last_login,
    cu.two_factor_enabled AS two_factor_enabled,
    ep.id           AS employee_id,
    ep.first_name,
    ep.last_name,
    (
        SELECT COALESCE(json_agg(json_build_object(
            'id',       p.id,
            'name',     p.name,
            'resource', p.resource,
            'method',   p.method
        )), '[]'::json)
        FROM effective_permissions ep2
        JOIN permissions p ON p.id = ep2.permission_id
    )::json AS permissions
FROM custom_user cu
JOIN employee_profile ep ON ep.user_id = cu.id
WHERE cu.id = $1;

-- name: GetEmployeeProfileByID :one
SELECT
    ep.*,
    cu.profile_picture as profile_picture,
    d.name AS department_name,
    mgr.first_name AS manager_first_name,
    mgr.last_name AS manager_last_name
FROM employee_profile ep
JOIN custom_user cu ON ep.user_id = cu.id
LEFT JOIN departments d ON d.id = ep.department_id
LEFT JOIN employee_profile mgr ON mgr.id = ep.manager_employee_id
WHERE ep.id = $1;

-- name: UpdateEmployeeProfile :one
UPDATE employee_profile
SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    position = COALESCE(sqlc.narg('position'), position),
    department_id = COALESCE(sqlc.narg('department_id'), department_id),
    manager_employee_id = COALESCE(sqlc.narg('manager_employee_id'), manager_employee_id),
    employee_number = COALESCE(sqlc.narg('employee_number'), employee_number),
    employment_number = COALESCE(sqlc.narg('employment_number'), employment_number),
    private_email_address = COALESCE(sqlc.narg('private_email_address'), private_email_address),
    work_email_address = COALESCE(sqlc.narg('work_email_address'), work_email_address),
    private_phone_number = COALESCE(sqlc.narg('private_phone_number'), private_phone_number),
    work_phone_number = COALESCE(sqlc.narg('work_phone_number'), work_phone_number),
    date_of_birth = COALESCE(sqlc.narg('date_of_birth'), date_of_birth),
    home_telephone_number = COALESCE(sqlc.narg('home_telephone_number'), home_telephone_number),
    gender = COALESCE(sqlc.narg('gender'), gender),
    location_id = COALESCE(sqlc.narg('location_id'), location_id),
    has_borrowed = COALESCE(sqlc.narg('has_borrowed'), has_borrowed),
    out_of_service = COALESCE(sqlc.narg('out_of_service'), out_of_service),
    is_archived = COALESCE(sqlc.narg('is_archived'), is_archived)
WHERE id = sqlc.arg('id')
RETURNING *;
