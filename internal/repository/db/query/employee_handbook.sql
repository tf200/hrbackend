-- name: CreateHandbookTemplateForDepartment :one
WITH next_version AS (
    SELECT COALESCE(MAX(version), 0) + 1 AS v
    FROM handbook_templates
    WHERE department_id = sqlc.arg('department_id')
)
INSERT INTO handbook_templates (
    department_id,
    title,
    description,
    version,
    status,
    created_by_employee_id
)
SELECT
    sqlc.arg('department_id'),
    sqlc.arg('title'),
    sqlc.narg('description'),
    next_version.v,
    'draft',
    sqlc.narg('created_by_employee_id')
FROM next_version
RETURNING *;

-- name: CloneHandbookTemplateToDraft :one
WITH source_template AS (
    SELECT id, department_id, title, description
    FROM handbook_templates
    WHERE handbook_templates.id = sqlc.arg('source_template_id')
    LIMIT 1
), next_version AS (
    SELECT COALESCE(MAX(version), 0) + 1 AS v
    FROM handbook_templates
    WHERE department_id = (SELECT department_id FROM source_template)
), cloned_template AS (
    INSERT INTO handbook_templates (
        department_id,
        title,
        description,
        version,
        status,
        created_by_employee_id
    )
    SELECT
        st.department_id,
        st.title,
        st.description,
        nv.v,
        'draft',
        sqlc.narg('created_by_employee_id')
    FROM source_template st
    CROSS JOIN next_version nv
    RETURNING *
), cloned_steps AS (
    INSERT INTO handbook_steps (
        template_id,
        sort_order,
        kind,
        title,
        body,
        content,
        is_required
    )
    SELECT
        ct.id,
        hs.sort_order,
        hs.kind,
        hs.title,
        hs.body,
        hs.content,
        hs.is_required
    FROM cloned_template ct
    JOIN source_template st ON TRUE
    JOIN handbook_steps hs ON hs.template_id = st.id
    RETURNING 1
)
SELECT * FROM cloned_template;

-- name: PublishHandbookTemplate :one
WITH target AS (
    SELECT id, department_id
    FROM handbook_templates
    WHERE handbook_templates.id = sqlc.arg('template_id')
      AND status = 'draft'
    LIMIT 1
), archived AS (
    UPDATE handbook_templates ht
    SET
        status = 'archived',
        archived_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    FROM target t
    WHERE ht.department_id = t.department_id
      AND ht.status = 'published'
)
UPDATE handbook_templates ht
SET
    status = 'published',
    published_by_employee_id = sqlc.narg('published_by_employee_id'),
    published_at = CURRENT_TIMESTAMP,
    archived_at = NULL,
    updated_at = CURRENT_TIMESTAMP
FROM target t
WHERE ht.id = t.id
RETURNING ht.*;

-- name: ListHandbookTemplatesByDepartment :many
SELECT *
FROM handbook_templates
WHERE department_id = $1
ORDER BY version DESC;

-- name: GetActiveHandbookTemplateByDepartment :one
SELECT *
FROM handbook_templates
WHERE department_id = $1 AND status = 'published'
LIMIT 1;

-- name: GetHandbookTemplateByID :one
SELECT *
FROM handbook_templates
WHERE id = $1
LIMIT 1;

-- name: UpdateHandbookTemplateMetadata :one
UPDATE handbook_templates
SET
    title = CASE
        WHEN sqlc.arg('set_title')::boolean THEN COALESCE(sqlc.narg('title'), title)
        ELSE title
    END,
    description = CASE
        WHEN sqlc.arg('set_description')::boolean THEN sqlc.narg('description')
        ELSE description
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('template_id')
  AND status = 'draft'
RETURNING *;

-- name: CreateHandbookStep :one
INSERT INTO handbook_steps (
    template_id,
    sort_order,
    kind,
    title,
    body,
    content,
    is_required
)
VALUES (
    sqlc.arg('template_id'),
    sqlc.arg('sort_order'),
    sqlc.arg('kind'),
    sqlc.arg('title'),
    sqlc.narg('body'),
    COALESCE(sqlc.narg('content'), '{}'::jsonb),
    COALESCE(sqlc.narg('is_required')::boolean, TRUE)
)
RETURNING *;

-- name: ListHandbookStepsByTemplate :many
SELECT *
FROM handbook_steps
WHERE template_id = $1
ORDER BY sort_order ASC;

-- name: GetHandbookStepByID :one
SELECT *
FROM handbook_steps
WHERE id = $1
LIMIT 1;

-- name: UpdateHandbookStepByID :one
UPDATE handbook_steps
SET
    title = CASE
        WHEN sqlc.arg('set_title')::boolean THEN COALESCE(sqlc.narg('title'), title)
        ELSE title
    END,
    body = CASE
        WHEN sqlc.arg('set_body')::boolean THEN sqlc.narg('body')
        ELSE body
    END,
    content = CASE
        WHEN sqlc.arg('set_content')::boolean THEN COALESCE(sqlc.narg('content'), 'null'::jsonb)
        ELSE content
    END,
    is_required = CASE
        WHEN sqlc.arg('set_is_required')::boolean THEN COALESCE(sqlc.narg('is_required')::boolean, is_required)
        ELSE is_required
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('step_id')
RETURNING *;

-- name: UpdateHandbookStepSortOrder :exec
UPDATE handbook_steps
SET
    sort_order = sqlc.arg('sort_order'),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('step_id');

-- name: DeleteHandbookStepByID :exec
DELETE FROM handbook_steps
WHERE id = $1;

-- name: CreateEmployeeHandbookFromTemplate :one
WITH hb AS (
    INSERT INTO employee_handbooks (
        employee_id,
        template_id,
        template_version,
        assigned_by_employee_id
    )
    SELECT
        sqlc.arg('employee_id'),
        t.id,
        t.version,
        sqlc.narg('assigned_by_employee_id')
    FROM handbook_templates t
    WHERE t.id = sqlc.arg('template_id')
    RETURNING *
), progress AS (
    INSERT INTO employee_handbook_step_progress (employee_handbook_id, step_id)
    SELECT hb.id, hs.id
    FROM hb
    JOIN handbook_steps hs ON hs.template_id = hb.template_id
)
SELECT * FROM hb;

-- name: ListEmployeesEligibleForDepartmentHandbookSeed :many
SELECT ep.id
FROM employee_profile ep
LEFT JOIN employee_handbooks eh
    ON eh.employee_id = ep.id
   AND eh.status IN ('not_started', 'in_progress')
WHERE ep.department_id = $1
  AND NOT ep.is_archived
  AND NOT COALESCE(ep.out_of_service, false)
  AND eh.id IS NULL
ORDER BY ep.created_at DESC
LIMIT $2;

-- name: ListEligibleEmployeesForHandbookAssignment :many
WITH active_assignments AS (
    SELECT DISTINCT employee_id
    FROM employee_handbooks
    WHERE status IN ('not_started', 'in_progress')
)
SELECT
    ep.id AS employee_id,
    ep.first_name,
    ep.last_name,
    ep.department_id,
    d.name AS department_name
FROM employee_profile ep
LEFT JOIN departments d ON d.id = ep.department_id
LEFT JOIN active_assignments aa ON aa.employee_id = ep.id
WHERE
    NOT ep.is_archived AND
    NOT COALESCE(ep.out_of_service, false) AND
    aa.employee_id IS NULL AND
    (ep.department_id = sqlc.narg('department_id') OR sqlc.narg('department_id') IS NULL) AND
    (sqlc.narg('search')::TEXT IS NULL OR
        ep.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
        ep.last_name ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY ep.first_name ASC, ep.last_name ASC, ep.id ASC
LIMIT $1 OFFSET $2;

-- name: CountEligibleEmployeesForHandbookAssignment :one
WITH active_assignments AS (
    SELECT DISTINCT employee_id
    FROM employee_handbooks
    WHERE status IN ('not_started', 'in_progress')
)
SELECT COUNT(*)
FROM employee_profile ep
LEFT JOIN active_assignments aa ON aa.employee_id = ep.id
WHERE
    NOT ep.is_archived AND
    NOT COALESCE(ep.out_of_service, false) AND
    aa.employee_id IS NULL AND
    (ep.department_id = sqlc.narg('department_id') OR sqlc.narg('department_id') IS NULL) AND
    (sqlc.narg('search')::TEXT IS NULL OR
        ep.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
        ep.last_name ILIKE '%' || sqlc.narg('search') || '%');

-- name: ListEmployeeHandbookAssignments :many
WITH latest_handbooks AS (
    SELECT
        eh.*,
        ROW_NUMBER() OVER (PARTITION BY eh.employee_id ORDER BY eh.assigned_at DESC) AS rn
    FROM employee_handbooks eh
)
SELECT
    ar.employee_id,
    ar.first_name,
    ar.last_name,
    ar.employee_department_id,
    ar.department_name,
    ar.employee_handbook_id,
    ar.handbook_template_id,
    ar.template_title,
    ar.template_version,
    ar.employee_handbook_status,
    ar.assigned_at,
    ar.started_at,
    ar.completed_at,
    ar.due_at,
    ar.required_steps_total,
    ar.required_steps_completed
FROM (
    SELECT
        ep.id AS employee_id,
        ep.first_name,
        ep.last_name,
        ep.department_id AS employee_department_id,
        d.name AS department_name,
        latest_handbooks.id AS employee_handbook_id,
        latest_handbooks.template_id AS handbook_template_id,
        ht.title AS template_title,
        latest_handbooks.template_version,
        COALESCE(latest_handbooks.status::text, 'unassigned') AS employee_handbook_status,
        latest_handbooks.assigned_at,
        latest_handbooks.started_at,
        latest_handbooks.completed_at,
        latest_handbooks.due_at,
        COALESCE(progress.required_steps_total, 0)::INT AS required_steps_total,
        COALESCE(progress.required_steps_completed, 0)::INT AS required_steps_completed,
        ep.is_archived
    FROM employee_profile ep
    LEFT JOIN departments d ON d.id = ep.department_id
    LEFT JOIN latest_handbooks ON latest_handbooks.employee_id = ep.id AND latest_handbooks.rn = 1
    LEFT JOIN handbook_templates ht ON ht.id = latest_handbooks.template_id
    LEFT JOIN LATERAL (
        SELECT
            COUNT(*) FILTER (WHERE hs.is_required = TRUE)::INT AS required_steps_total,
            COUNT(*) FILTER (WHERE hs.is_required = TRUE AND ehsp.status = 'completed')::INT AS required_steps_completed
        FROM employee_handbook_step_progress ehsp
        JOIN handbook_steps hs ON hs.id = ehsp.step_id
        WHERE ehsp.employee_handbook_id = latest_handbooks.id
    ) progress ON TRUE
    WHERE
        (ep.department_id = sqlc.narg('department_id') OR sqlc.narg('department_id') IS NULL) AND
        (sqlc.narg('status_filter')::TEXT IS NULL OR COALESCE(latest_handbooks.status::text, 'unassigned') = sqlc.narg('status_filter')::TEXT) AND
        (sqlc.narg('search')::TEXT IS NULL OR
            ep.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
            ep.last_name ILIKE '%' || sqlc.narg('search') || '%')
) AS ar
ORDER BY ar.assigned_at DESC NULLS LAST, ar.employee_id
LIMIT $1 OFFSET $2;

-- name: CountEmployeeHandbookAssignments :one
WITH latest_handbooks AS (
    SELECT
        eh.*,
        ROW_NUMBER() OVER (PARTITION BY eh.employee_id ORDER BY eh.assigned_at DESC) AS rn
    FROM employee_handbooks eh
)
SELECT COUNT(*)
FROM (
    SELECT
        ep.id AS employee_id,
        ep.first_name,
        ep.last_name,
        ep.department_id AS employee_department_id,
        latest_handbooks.template_id AS handbook_template_id,
        COALESCE(latest_handbooks.status::text, 'unassigned') AS employee_handbook_status,
        ep.is_archived
    FROM employee_profile ep
    LEFT JOIN latest_handbooks ON latest_handbooks.employee_id = ep.id AND latest_handbooks.rn = 1
    WHERE
        (ep.department_id = sqlc.narg('department_id') OR sqlc.narg('department_id') IS NULL) AND
        (sqlc.narg('status_filter')::TEXT IS NULL OR COALESCE(latest_handbooks.status::text, 'unassigned') = sqlc.narg('status_filter')::TEXT) AND
        (sqlc.narg('search')::TEXT IS NULL OR
            ep.first_name ILIKE '%' || sqlc.narg('search') || '%' OR
            ep.last_name ILIKE '%' || sqlc.narg('search') || '%')
) AS ar
;

-- name: GetActiveEmployeeHandbookByEmployeeID :one
SELECT
    eh.*,
    ht.title AS template_title,
    ht.description AS template_description,
    d.id AS department_id,
    d.name AS department_name
FROM employee_handbooks eh
JOIN handbook_templates ht ON ht.id = eh.template_id
JOIN departments d ON d.id = ht.department_id
WHERE eh.employee_id = $1
  AND eh.status IN ('not_started', 'in_progress')
ORDER BY eh.assigned_at DESC
LIMIT 1;

-- name: GetEmployeeHandbookByID :one
SELECT *
FROM employee_handbooks
WHERE id = $1
LIMIT 1;

-- name: GetEmployeeHandbookDetailsByID :one
SELECT
    eh.*,
    ht.title AS template_title,
    ht.description AS template_description,
    d.id AS department_id,
    d.name AS department_name,
    ep.first_name,
    ep.last_name
FROM employee_handbooks eh
JOIN handbook_templates ht ON ht.id = eh.template_id
JOIN departments d ON d.id = ht.department_id
JOIN employee_profile ep ON ep.id = eh.employee_id
WHERE eh.id = $1
LIMIT 1;

-- name: ListEmployeeHandbookStepsByHandbookID :many
SELECT
    hs.id AS step_id,
    hs.sort_order,
    hs.kind,
    hs.title,
    hs.body,
    hs.content,
    hs.is_required,
    ehsp.status AS progress_status,
    ehsp.started_at AS progress_started_at,
    ehsp.completed_at AS progress_completed_at,
    ehsp.response AS progress_response
FROM employee_handbook_step_progress ehsp
JOIN handbook_steps hs ON hs.id = ehsp.step_id
WHERE ehsp.employee_handbook_id = $1
ORDER BY hs.sort_order ASC;

-- name: MarkEmployeeHandbookStarted :one
UPDATE employee_handbooks
SET
    status = 'in_progress',
    started_at = COALESCE(started_at, CURRENT_TIMESTAMP)
WHERE id = $1
RETURNING *;

-- name: WaiveActiveEmployeeHandbooksByEmployeeID :exec
UPDATE employee_handbooks
SET
    status = 'waived',
    completed_at = COALESCE(completed_at, CURRENT_TIMESTAMP)
WHERE employee_id = $1
  AND status IN ('not_started', 'in_progress');

-- name: WaiveEmployeeHandbookByID :one
UPDATE employee_handbooks
SET
    status = 'waived',
    completed_at = COALESCE(completed_at, CURRENT_TIMESTAMP)
WHERE id = $1
  AND status IN ('not_started', 'in_progress')
RETURNING *;

-- name: CreateEmployeeHandbookAssignmentHistory :one
INSERT INTO employee_handbook_assignment_history (
    employee_handbook_id,
    employee_id,
    template_id,
    template_version,
    event,
    actor_employee_id,
    metadata
)
VALUES (
    sqlc.narg('employee_handbook_id'),
    sqlc.arg('employee_id'),
    sqlc.arg('template_id'),
    sqlc.arg('template_version'),
    sqlc.arg('event'),
    sqlc.narg('actor_employee_id'),
    COALESCE(sqlc.narg('metadata'), '{}'::jsonb)
)
RETURNING *;

-- name: ListEmployeeHandbookAssignmentHistoryByEmployeeID :many
SELECT *
FROM employee_handbook_assignment_history
WHERE employee_id = sqlc.arg('employee_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CompleteEmployeeHandbookStep :one
UPDATE employee_handbook_step_progress
SET
    status = 'completed',
    started_at = COALESCE(started_at, CURRENT_TIMESTAMP),
    completed_at = CURRENT_TIMESTAMP,
    response = COALESCE(sqlc.narg('response'), response)
WHERE employee_handbook_id = sqlc.arg('employee_handbook_id')
  AND step_id = sqlc.arg('step_id')
RETURNING *;

-- name: CountRemainingRequiredHandbookSteps :one
SELECT COUNT(*)::INT
FROM employee_handbook_step_progress p
JOIN handbook_steps s ON s.id = p.step_id
WHERE p.employee_handbook_id = $1
  AND s.is_required = TRUE
  AND p.status <> 'completed';

-- name: MarkEmployeeHandbookCompleted :one
UPDATE employee_handbooks
SET
    status = 'completed',
    completed_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CountHandbookStepsByTemplateID :one
SELECT COUNT(*)::INT
FROM handbook_steps
WHERE template_id = $1;
