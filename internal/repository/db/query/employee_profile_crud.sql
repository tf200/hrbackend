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
    manager_employee_id,
    employee_number,
    employment_number,
    private_email_address,
    work_email_address,
    work_phone_number,
    private_phone_number,
    date_of_birth,
    home_telephone_number,
    gender
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
) RETURNING *;

-- name: ListEmployeeProfile :many
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT
    ep.id,
    ep.first_name,
    ep.last_name,
    ep.bsn,
    ec.contract_type,
    d.name AS department_name,
    ec.contract_end_date,
    concat_ws(' ', l.street, l.house_number, l.house_number_addition, l.postal_code, l.city) AS location_address
FROM employee_profile ep
LEFT JOIN latest_contract ec ON ec.employee_id = ep.id
LEFT JOIN location l ON l.id = ec.location_id
LEFT JOIN departments d ON d.id = ec.department_id
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
    (sqlc.narg('location_id')::uuid IS NULL OR ec.location_id = sqlc.narg('location_id')::uuid) AND
    (sqlc.narg('contract_type')::employee_contract_type_enum IS NULL OR ec.contract_type = sqlc.narg('contract_type')::employee_contract_type_enum) AND
    (sqlc.narg('search')::TEXT IS NULL OR
        ep.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
        ep.last_name ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY ep.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountEmployeeProfile :one
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT COUNT(*)
FROM employee_profile ep
LEFT JOIN latest_contract ec ON ec.employee_id = ep.id
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
    (sqlc.narg('location_id')::uuid IS NULL OR ec.location_id = sqlc.narg('location_id')::uuid) AND
    (sqlc.narg('contract_type')::employee_contract_type_enum IS NULL OR ec.contract_type = sqlc.narg('contract_type')::employee_contract_type_enum);

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
    COALESCE(r.name, '') AS role,
    ur.role_id      AS role_id,
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
LEFT JOIN user_roles ur ON ur.user_id = cu.id
LEFT JOIN roles r ON r.id = ur.role_id
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
LEFT JOIN LATERAL (
    SELECT *
    FROM employee_contracts c
    WHERE c.employee_id = ep.id
    ORDER BY c.start_date DESC, c.created_at DESC
    LIMIT 1
) ec ON true
LEFT JOIN departments d ON d.id = ec.department_id
LEFT JOIN employee_profile mgr ON mgr.id = ep.manager_employee_id
WHERE ep.id = $1;

-- name: UpdateEmployeeProfile :one
UPDATE employee_profile
SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
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
    out_of_service = COALESCE(sqlc.narg('out_of_service'), out_of_service),
    is_archived = COALESCE(sqlc.narg('is_archived'), is_archived)
WHERE id = sqlc.arg('id')
RETURNING *;
