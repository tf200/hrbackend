-- name: CreateOvertimeEntry :one
WITH inserted_overtime_entry AS (
    INSERT INTO overtime_entries (
        employee_id,
        schedule_id,
        entry_date,
        minutes,
        reason,
        description
    ) VALUES (
        sqlc.arg(employee_id),
        sqlc.narg(schedule_id),
        sqlc.arg(entry_date),
        sqlc.arg(minutes),
        sqlc.arg(reason),
        sqlc.narg(description)
    )
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        minutes,
        reason,
        description,
        status,
        submitted_at,
        approved_at,
        approved_by_employee_id,
        rejection_reason,
        paid_period_id,
        created_at,
        updated_at
)
SELECT
    oe.id,
    oe.employee_id,
    oe.schedule_id,
    oe.entry_date,
    oe.minutes,
    oe.reason,
    oe.description,
    oe.status,
    oe.submitted_at,
    oe.approved_at,
    oe.approved_by_employee_id,
    oe.rejection_reason,
    oe.paid_period_id,
    oe.created_at,
    oe.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM inserted_overtime_entry oe
JOIN employee_profile ep ON ep.id = oe.employee_id
LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id;

-- name: GetOvertimeEntryByID :one
SELECT
    oe.id,
    oe.employee_id,
    oe.schedule_id,
    oe.entry_date,
    oe.minutes,
    oe.reason,
    oe.description,
    oe.status,
    oe.submitted_at,
    oe.approved_at,
    oe.approved_by_employee_id,
    oe.rejection_reason,
    oe.paid_period_id,
    oe.created_at,
    oe.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM overtime_entries oe
JOIN employee_profile ep ON ep.id = oe.employee_id
LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id
WHERE oe.id = sqlc.arg(id);

-- name: LockOvertimeEntryByID :one
SELECT *
FROM overtime_entries
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: DeleteOvertimeEntry :one
DELETE FROM overtime_entries
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: ListOvertimeEntriesPaginated :many
WITH filtered AS (
    SELECT
        oe.id,
        oe.employee_id,
        oe.schedule_id,
        oe.entry_date,
        oe.minutes,
        oe.reason,
        oe.description,
        oe.status,
        oe.submitted_at,
        oe.approved_at,
        oe.approved_by_employee_id,
        oe.rejection_reason,
        oe.paid_period_id,
        oe.created_at,
        oe.updated_at,
        ep.first_name AS employee_first_name,
        ep.last_name AS employee_last_name,
        ap.first_name AS approved_by_first_name,
        ap.last_name AS approved_by_last_name,
        COUNT(*) OVER() AS total_count
    FROM overtime_entries oe
    JOIN employee_profile ep ON ep.id = oe.employee_id
    LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id
    WHERE (
        sqlc.narg(status)::overtime_status_enum IS NULL
        OR oe.status = sqlc.narg(status)::overtime_status_enum
    )
    ORDER BY oe.created_at DESC
)
SELECT *
FROM filtered
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: ListMyOvertimeEntriesPaginated :many
WITH filtered AS (
    SELECT
        oe.id,
        oe.employee_id,
        oe.schedule_id,
        oe.entry_date,
        oe.minutes,
        oe.reason,
        oe.description,
        oe.status,
        oe.submitted_at,
        oe.approved_at,
        oe.approved_by_employee_id,
        oe.rejection_reason,
        oe.paid_period_id,
        oe.created_at,
        oe.updated_at,
        ep.first_name AS employee_first_name,
        ep.last_name AS employee_last_name,
        ap.first_name AS approved_by_first_name,
        ap.last_name AS approved_by_last_name,
        COUNT(*) OVER() AS total_count
    FROM overtime_entries oe
    JOIN employee_profile ep ON ep.id = oe.employee_id
    LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id
    WHERE oe.employee_id = sqlc.arg(employee_id)
      AND (
        sqlc.narg(status)::overtime_status_enum IS NULL
        OR oe.status = sqlc.narg(status)::overtime_status_enum
    )
    ORDER BY oe.created_at DESC
)
SELECT *
FROM filtered
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: ApproveOvertimeEntry :one
WITH updated_overtime_entry AS (
    UPDATE overtime_entries
    SET
        status = 'approved'::overtime_status_enum,
        approved_at = NOW(),
        approved_by_employee_id = sqlc.arg(approved_by_employee_id),
        rejection_reason = NULL,
        updated_at = NOW()
    WHERE overtime_entries.id = sqlc.arg(id)
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        minutes,
        reason,
        description,
        status,
        submitted_at,
        approved_at,
        approved_by_employee_id,
        rejection_reason,
        paid_period_id,
        created_at,
        updated_at
)
SELECT
    oe.id,
    oe.employee_id,
    oe.schedule_id,
    oe.entry_date,
    oe.minutes,
    oe.reason,
    oe.description,
    oe.status,
    oe.submitted_at,
    oe.approved_at,
    oe.approved_by_employee_id,
    oe.rejection_reason,
    oe.paid_period_id,
    oe.created_at,
    oe.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM updated_overtime_entry oe
JOIN employee_profile ep ON ep.id = oe.employee_id
LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id;

-- name: RejectOvertimeEntry :one
WITH updated_overtime_entry AS (
    UPDATE overtime_entries
    SET
        status = 'rejected'::overtime_status_enum,
        rejection_reason = sqlc.narg(rejection_reason)::text,
        approved_at = NULL,
        approved_by_employee_id = NULL,
        updated_at = NOW()
    WHERE overtime_entries.id = sqlc.arg(id)
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        minutes,
        reason,
        description,
        status,
        submitted_at,
        approved_at,
        approved_by_employee_id,
        rejection_reason,
        paid_period_id,
        created_at,
        updated_at
)
SELECT
    oe.id,
    oe.employee_id,
    oe.schedule_id,
    oe.entry_date,
    oe.minutes,
    oe.reason,
    oe.description,
    oe.status,
    oe.submitted_at,
    oe.approved_at,
    oe.approved_by_employee_id,
    oe.rejection_reason,
    oe.paid_period_id,
    oe.created_at,
    oe.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM updated_overtime_entry oe
JOIN employee_profile ep ON ep.id = oe.employee_id
LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id;

-- name: UpdateOvertimeEntryByAdmin :one
WITH updated_overtime_entry AS (
    UPDATE overtime_entries
    SET
        schedule_id = COALESCE(sqlc.narg(schedule_id)::uuid, schedule_id),
        entry_date = COALESCE(sqlc.narg(entry_date)::date, entry_date),
        minutes = COALESCE(sqlc.narg(minutes)::integer, minutes),
        reason = COALESCE(sqlc.narg(reason)::overtime_reason_enum, reason),
        description = COALESCE(sqlc.narg(description)::text, description),
        updated_at = NOW()
    WHERE overtime_entries.id = sqlc.arg(id)
    RETURNING
        id,
        employee_id,
        schedule_id,
        entry_date,
        minutes,
        reason,
        description,
        status,
        submitted_at,
        approved_at,
        approved_by_employee_id,
        rejection_reason,
        paid_period_id,
        created_at,
        updated_at
)
SELECT
    oe.id,
    oe.employee_id,
    oe.schedule_id,
    oe.entry_date,
    oe.minutes,
    oe.reason,
    oe.description,
    oe.status,
    oe.submitted_at,
    oe.approved_at,
    oe.approved_by_employee_id,
    oe.rejection_reason,
    oe.paid_period_id,
    oe.created_at,
    oe.updated_at,
    ep.first_name AS employee_first_name,
    ep.last_name AS employee_last_name,
    ap.first_name AS approved_by_first_name,
    ap.last_name AS approved_by_last_name
FROM updated_overtime_entry oe
JOIN employee_profile ep ON ep.id = oe.employee_id
LEFT JOIN employee_profile ap ON ap.id = oe.approved_by_employee_id;

-- name: GetCurrentMonthOvertimeStats :one
WITH ranges AS (
    SELECT
        date_trunc('month', CURRENT_DATE)::date AS month_start,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month')::date AS next_month_start
)
SELECT
    COALESCE(SUM(oe.minutes) FILTER (
        WHERE oe.status = 'approved'::overtime_status_enum
    ), 0)::BIGINT AS total_approved_minutes,
    COALESCE(COUNT(*) FILTER (
        WHERE oe.status = 'submitted'::overtime_status_enum
    ), 0)::BIGINT AS total_awaiting_approval,
    COALESCE(COUNT(*) FILTER (
        WHERE oe.status = 'approved'::overtime_status_enum
    ), 0)::BIGINT AS total_approved,
    COALESCE(COUNT(*) FILTER (
        WHERE oe.status = 'submitted'::overtime_status_enum
    ), 0)::BIGINT AS total_submitted
FROM ranges r
CROSS JOIN overtime_entries oe
WHERE oe.entry_date >= r.month_start
  AND oe.entry_date < r.next_month_start;

-- name: GetMyCurrentMonthOvertimeStats :one
WITH ranges AS (
    SELECT
        date_trunc('month', CURRENT_DATE)::date AS month_start,
        (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month')::date AS next_month_start
)
SELECT
    COALESCE(SUM(oe.minutes) FILTER (
        WHERE oe.status = 'approved'::overtime_status_enum
    ), 0)::BIGINT AS total_approved_minutes,
    COALESCE(COUNT(*) FILTER (
        WHERE oe.status = 'submitted'::overtime_status_enum
    ), 0)::BIGINT AS total_awaiting_approval,
    COALESCE(COUNT(*) FILTER (
        WHERE oe.status = 'approved'::overtime_status_enum
    ), 0)::BIGINT AS total_approved,
    COALESCE(COUNT(*) FILTER (
        WHERE oe.status = 'submitted'::overtime_status_enum
    ), 0)::BIGINT AS total_submitted
FROM ranges r
CROSS JOIN overtime_entries oe
WHERE oe.employee_id = sqlc.arg(employee_id)
  AND oe.entry_date >= r.month_start
  AND oe.entry_date < r.next_month_start;
