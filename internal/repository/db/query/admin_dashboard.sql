-- name: ListRecentDashboardEmployees :many
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT
    ep.id,
    ep.first_name,
    ep.last_name,
    org_role.name AS organizational_role_name,
    d.name AS department_name,
    l.name AS location_name,
    ep.created_at
FROM employee_profile ep
LEFT JOIN latest_contract ec ON ec.employee_id = ep.id
LEFT JOIN departments d ON d.id = ec.department_id
LEFT JOIN location l ON l.id = ec.location_id
LEFT JOIN organizational_roles org_role ON org_role.id = ec.organizational_role_id
WHERE ep.is_archived = FALSE AND COALESCE(ep.out_of_service, FALSE) = FALSE
ORDER BY ep.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountRecentDashboardEmployees :one
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT COUNT(*)
FROM employee_profile ep
LEFT JOIN latest_contract ec ON ec.employee_id = ep.id
WHERE ep.is_archived = FALSE AND COALESCE(ep.out_of_service, FALSE) = FALSE;

-- name: ListFullTimeEmployeesByDepartment :many
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT
    d.id AS department_id,
    d.name AS department_name,
    COUNT(*)::bigint AS total_employees
FROM employee_profile ep
JOIN latest_contract ec ON ec.employee_id = ep.id
JOIN departments d ON d.id = ec.department_id
WHERE ep.is_archived = FALSE
  AND COALESCE(ep.out_of_service, FALSE) = FALSE
  AND ec.contract_type = 'permanent'
  AND ec.start_date <= CURRENT_DATE
  AND (ec.effective_end_date IS NULL OR ec.effective_end_date >= CURRENT_DATE)
  AND (ec.contract_end_date IS NULL OR ec.contract_end_date >= CURRENT_DATE)
GROUP BY d.id, d.name
ORDER BY d.name;

-- name: ListFullTimeEmployeesByLocation :many
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT
    l.id AS location_id,
    l.name AS location_name,
    COUNT(*)::bigint AS total_employees
FROM employee_profile ep
JOIN latest_contract ec ON ec.employee_id = ep.id
JOIN location l ON l.id = ec.location_id
WHERE ep.is_archived = FALSE
  AND COALESCE(ep.out_of_service, FALSE) = FALSE
  AND ec.contract_type = 'permanent'
  AND ec.start_date <= CURRENT_DATE
  AND (ec.effective_end_date IS NULL OR ec.effective_end_date >= CURRENT_DATE)
  AND (ec.contract_end_date IS NULL OR ec.contract_end_date >= CURRENT_DATE)
GROUP BY l.id, l.name
ORDER BY l.name;

-- name: ListLeaveAbsenceTrendPoints :many
WITH months AS (
    SELECT generate_series(
        date_trunc('month', sqlc.arg('from_date')::date),
        date_trunc('month', sqlc.arg('to_date')::date),
        interval '1 month'
    )::date AS month_start
), month_ranges AS (
    SELECT
        month_start,
        (month_start + interval '1 month')::date AS month_end_exclusive
    FROM months
)
SELECT
    mr.month_start,
    COUNT(DISTINCT lr.employee_id)::bigint AS employees_out
FROM month_ranges mr
LEFT JOIN leave_requests lr
    ON lr.status = 'approved'
   AND lr.start_date < mr.month_end_exclusive
   AND lr.end_date >= mr.month_start
GROUP BY mr.month_start
ORDER BY mr.month_start;

-- name: ListEndingContractAlerts :many
WITH latest_contract AS (
    SELECT DISTINCT ON (employee_id) *
    FROM employee_contracts
    ORDER BY employee_id, start_date DESC, created_at DESC
)
SELECT
    ep.id AS employee_id,
    concat_ws(' ', ep.first_name, ep.last_name) AS employee_name,
    ec.id AS contract_id,
    ec.contract_type::text AS contract_type,
    ec.contract_end_date,
    (ec.contract_end_date - CURRENT_DATE)::int AS days_remaining,
    d.name AS department,
    l.name AS location
FROM latest_contract ec
JOIN employee_profile ep ON ep.id = ec.employee_id
JOIN departments d ON d.id = ec.department_id
JOIN location l ON l.id = ec.location_id
WHERE ep.is_archived = FALSE
  AND COALESCE(ep.out_of_service, FALSE) = FALSE
  AND ec.start_date <= CURRENT_DATE
  AND (ec.effective_end_date IS NULL OR ec.effective_end_date >= CURRENT_DATE)
  AND ec.contract_end_date BETWEEN CURRENT_DATE AND sqlc.arg('to_date')::date
ORDER BY ec.contract_end_date ASC, ep.first_name ASC, ep.last_name ASC
LIMIT sqlc.arg('limit');

-- name: ListExpiringCredentialAlerts :many
SELECT
    ep.id AS employee_id,
    concat_ws(' ', ep.first_name, ep.last_name) AS employee_name,
    credential_id,
    credential_type,
    name,
    expiry_date,
    (expiry_date - CURRENT_DATE)::int AS days_remaining
FROM (
    SELECT
        eq.employee_id,
        eq.id AS credential_id,
        'qualification'::text AS credential_type,
        q.name,
        eq.expiration_date AS expiry_date
    FROM employee_qualifications eq
    JOIN qualifications q ON q.id = eq.qualification_id
    WHERE q.is_active = TRUE
      AND eq.expiration_date BETWEEN CURRENT_DATE AND sqlc.arg('to_date')::date

    UNION ALL

    SELECT
        ea.employee_id,
        ea.id AS credential_id,
        'authorization'::text AS credential_type,
        a.name,
        ea.expiry_date AS expiry_date
    FROM employee_authorizations ea
    JOIN authorizations a ON a.id = ea.authorization_id
    WHERE a.is_active = TRUE
      AND ea.is_active = TRUE
      AND ea.expiry_date BETWEEN CURRENT_DATE AND sqlc.arg('to_date')::date
) credentials
JOIN employee_profile ep ON ep.id = credentials.employee_id
WHERE ep.is_archived = FALSE
  AND COALESCE(ep.out_of_service, FALSE) = FALSE
ORDER BY expiry_date ASC, ep.first_name ASC, ep.last_name ASC, name ASC
LIMIT sqlc.arg('limit');

-- name: ListReturningFromLeaveAlerts :many
SELECT
    ep.id AS employee_id,
    concat_ws(' ', ep.first_name, ep.last_name) AS employee_name,
    lr.id AS leave_request_id,
    lr.leave_type::text AS leave_type,
    lr.end_date AS leave_end_date,
    (lr.end_date + 1)::date AS return_date,
    ((lr.end_date + 1)::date - CURRENT_DATE)::int AS days_until_return
FROM leave_requests lr
JOIN employee_profile ep ON ep.id = lr.employee_id
WHERE ep.is_archived = FALSE
  AND COALESCE(ep.out_of_service, FALSE) = FALSE
  AND lr.status = 'approved'
  AND (lr.end_date + 1)::date BETWEEN CURRENT_DATE AND sqlc.arg('to_date')::date
ORDER BY return_date ASC, ep.first_name ASC, ep.last_name ASC
LIMIT sqlc.arg('limit');

-- name: GetAdminDashboardKPIs :one
SELECT
    (
        SELECT COUNT(*)
        FROM employee_profile
        WHERE is_archived = FALSE AND COALESCE(out_of_service, FALSE) = FALSE
    ) AS total_employees,
    (
        SELECT COUNT(*)
        FROM employee_profile ep
        WHERE ep.is_archived = FALSE
          AND COALESCE(ep.out_of_service, FALSE) = FALSE
          AND NOT EXISTS (
              SELECT 1
              FROM leave_requests lr
              WHERE lr.employee_id = ep.id
                AND lr.status = 'approved'
                AND CURRENT_DATE BETWEEN lr.start_date AND lr.end_date
          )
    ) AS employees_present,
    (
        SELECT COUNT(*)
        FROM attachment_file
    ) AS total_documents,
    (
        SELECT COUNT(*)
        FROM sign_documents
        WHERE status IN ('sent', 'partially_signed')
    ) AS processing_docs;
