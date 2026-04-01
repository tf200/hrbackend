-- name: CreateTimeEntry :one
WITH inserted_time_entry AS (
    INSERT INTO time_entries (
        employee_id,
        schedule_id,
        entry_date,
        hours,
        hour_type,
        project_name,
        project_number,
        client_name,
        activity_category,
        activity_description,
        notes
    ) VALUES (
        sqlc.arg(employee_id),
        sqlc.narg(schedule_id),
        sqlc.arg(entry_date),
        sqlc.arg(hours),
        sqlc.arg(hour_type),
        sqlc.narg(project_name),
        sqlc.narg(project_number),
        sqlc.narg(client_name),
        sqlc.narg(activity_category),
        sqlc.narg(activity_description),
        sqlc.narg(notes)
    )
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        hours,
        hour_type,
        project_name,
        project_number,
        client_name,
        activity_category,
        activity_description,
        status,
        submitted_at,
        approved_at,
        approved_by_employee_id,
        rejection_reason,
        notes,
        created_at,
        updated_at
)
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.hours,
    te.hour_type,
    te.project_name,
    te.project_number,
    te.client_name,
    te.activity_category,
    te.activity_description,
    te.status,
    te.submitted_at,
    te.approved_at,
    te.approved_by_employee_id,
    te.rejection_reason,
    te.notes,
    te.created_at,
    te.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM inserted_time_entry te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id;

-- name: GetTimeEntryByID :one
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.hours,
    te.hour_type,
    te.project_name,
    te.project_number,
    te.client_name,
    te.activity_category,
    te.activity_description,
    te.status,
    te.submitted_at,
    te.approved_at,
    te.approved_by_employee_id,
    te.rejection_reason,
    te.notes,
    te.created_at,
    te.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM time_entries te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id
WHERE te.id = $1
LIMIT 1;

-- name: ListTimeEntriesPaginated :many
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.hours,
    te.hour_type,
    te.project_name,
    te.project_number,
    te.client_name,
    te.activity_category,
    te.activity_description,
    te.status,
    te.submitted_at,
    te.approved_at,
    te.approved_by_employee_id,
    te.rejection_reason,
    te.notes,
    te.created_at,
    te.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name,
    COUNT(*) OVER() AS total_count
FROM time_entries te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id
WHERE (
    sqlc.narg('employee_id')::uuid IS NULL
    OR te.employee_id = sqlc.narg('employee_id')::uuid
)
  AND (
    sqlc.narg('status')::time_entry_status_enum IS NULL
    OR te.status = sqlc.narg('status')::time_entry_status_enum
  )
  AND (
    sqlc.narg('employee_search')::text IS NULL
    OR sqlc.narg('employee_search')::text = ''
    OR ep.first_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR ep.last_name ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.first_name || ' ' || ep.last_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
    OR (ep.last_name || ' ' || ep.first_name) ILIKE '%' || sqlc.narg('employee_search')::text || '%'
  )
ORDER BY te.entry_date DESC, te.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListMyTimeEntriesPaginated :many
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.hours,
    te.hour_type,
    te.project_name,
    te.project_number,
    te.client_name,
    te.activity_category,
    te.activity_description,
    te.status,
    te.submitted_at,
    te.approved_at,
    te.approved_by_employee_id,
    te.rejection_reason,
    te.notes,
    te.created_at,
    te.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name,
    COUNT(*) OVER() AS total_count
FROM time_entries te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id
WHERE te.employee_id = sqlc.arg(employee_id)
  AND (
    sqlc.narg('status')::time_entry_status_enum IS NULL
    OR te.status = sqlc.narg('status')::time_entry_status_enum
  )
ORDER BY te.entry_date DESC, te.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
