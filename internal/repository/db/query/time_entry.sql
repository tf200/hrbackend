-- name: CreateTimeEntry :one
WITH inserted_time_entry AS (
    INSERT INTO time_entries (
        employee_id,
        schedule_id,
        entry_date,
        start_time,
        end_time,
        break_minutes,
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
        sqlc.arg(start_time),
        sqlc.arg(end_time),
        sqlc.arg(break_minutes),
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
        start_time,
        end_time,
        break_minutes,
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
        updated_at,
        paid_period_id
)
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM inserted_time_entry te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id;

-- name: LockTimeEntryByID :one
SELECT *
FROM time_entries
WHERE id = $1
FOR UPDATE;

-- name: ApproveTimeEntry :one
WITH updated_time_entry AS (
    UPDATE time_entries
    SET
        status = 'approved'::time_entry_status_enum,
        approved_at = NOW(),
        approved_by_employee_id = sqlc.arg('approved_by_employee_id'),
        rejection_reason = NULL,
        updated_at = NOW()
    WHERE time_entries.id = sqlc.arg('id')
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        start_time,
        end_time,
        break_minutes,
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
        updated_at,
        paid_period_id
)
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM updated_time_entry te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id;

-- name: RejectTimeEntry :one
WITH updated_time_entry AS (
    UPDATE time_entries
    SET
        status = 'rejected'::time_entry_status_enum,
        rejection_reason = sqlc.narg('rejection_reason')::text,
        approved_at = NULL,
        approved_by_employee_id = NULL,
        updated_at = NOW()
    WHERE time_entries.id = sqlc.arg('id')
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        start_time,
        end_time,
        break_minutes,
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
        updated_at,
        paid_period_id
)
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM updated_time_entry te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id;

-- name: UpdateTimeEntryByAdmin :one
WITH updated_time_entry AS (
    UPDATE time_entries
    SET
        schedule_id = COALESCE(sqlc.narg('schedule_id')::uuid, schedule_id),
        entry_date = COALESCE(sqlc.narg('entry_date')::date, entry_date),
        start_time = COALESCE(sqlc.narg('start_time')::time, start_time),
        end_time = COALESCE(sqlc.narg('end_time')::time, end_time),
        break_minutes = COALESCE(sqlc.narg('break_minutes')::integer, break_minutes),
        hour_type = COALESCE(sqlc.narg('hour_type')::time_entry_hour_type_enum, hour_type),
        project_name = COALESCE(sqlc.narg('project_name')::text, project_name),
        project_number = COALESCE(sqlc.narg('project_number')::text, project_number),
        client_name = COALESCE(sqlc.narg('client_name')::text, client_name),
        activity_category = COALESCE(sqlc.narg('activity_category')::text, activity_category),
        activity_description = COALESCE(
            sqlc.narg('activity_description')::text,
            activity_description
        ),
        notes = COALESCE(sqlc.narg('notes')::text, notes),
        status = CASE
            WHEN sqlc.arg('set_submitted')::boolean
                THEN 'submitted'::time_entry_status_enum
            ELSE status
        END,
        submitted_at = CASE
            WHEN sqlc.arg('set_submitted')::boolean THEN NOW()
            ELSE submitted_at
        END,
        approved_at = CASE
            WHEN sqlc.arg('set_submitted')::boolean THEN NULL
            ELSE approved_at
        END,
        approved_by_employee_id = CASE
            WHEN sqlc.arg('set_submitted')::boolean THEN NULL
            ELSE approved_by_employee_id
        END,
        rejection_reason = CASE
            WHEN sqlc.arg('set_submitted')::boolean THEN NULL
            ELSE rejection_reason
        END,
        updated_at = NOW()
    WHERE time_entries.id = sqlc.arg('id')
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        start_time,
        end_time,
        break_minutes,
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
        updated_at,
        paid_period_id
)
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM updated_time_entry te
JOIN employee_profile ep ON ep.id = te.employee_id
LEFT JOIN employee_profile ap ON ap.id = te.approved_by_employee_id;

-- name: CreateTimeEntryUpdateAudit :exec
INSERT INTO time_entry_update_audits (
    time_entry_id,
    admin_employee_id,
    admin_update_note,
    before_snapshot,
    after_snapshot
) VALUES (
    sqlc.arg('time_entry_id'),
    sqlc.arg('admin_employee_id'),
    sqlc.arg('admin_update_note'),
    sqlc.arg('before_snapshot'),
    sqlc.arg('after_snapshot')
);

-- name: GetTimeEntryByID :one
SELECT
    te.id,
    te.employee_id,
    te.schedule_id,
    te.entry_date,
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
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
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
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
    te.start_time,
    te.end_time,
    te.break_minutes,
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
    te.paid_period_id,
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

-- name: GetCurrentMonthTimeEntryStats :one
SELECT
    COALESCE(
        SUM(
            GREATEST(
                0,
                (
                    CASE
                        WHEN te.end_time > te.start_time THEN
                            EXTRACT(EPOCH FROM te.end_time) - EXTRACT(EPOCH FROM te.start_time)
                        ELSE
                            EXTRACT(EPOCH FROM te.end_time) + 86400 - EXTRACT(EPOCH FROM te.start_time)
                    END
                ) / 60 - te.break_minutes
            )
        ),
        0
    )::BIGINT AS total_worked_minutes,
    COUNT(*) FILTER (
        WHERE te.status = 'submitted'::time_entry_status_enum
    )::BIGINT AS total_awaiting_approval,
    COUNT(*) FILTER (
        WHERE te.status = 'approved'::time_entry_status_enum
    )::BIGINT AS total_approved,
    COUNT(*) FILTER (
        WHERE te.status = 'draft'::time_entry_status_enum
    )::BIGINT AS total_concepts
FROM time_entries te
WHERE te.entry_date >= DATE_TRUNC('month', NOW())::date
  AND te.entry_date < (DATE_TRUNC('month', NOW()) + INTERVAL '1 month')::date;
