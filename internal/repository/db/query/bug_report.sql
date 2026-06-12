-- name: CreateBugReport :one
INSERT INTO bug_reports (
    user_id,
    subject,
    category,
    severity,
    description,
    steps,
    debug_info
) VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('subject'),
    sqlc.arg('category')::bug_report_category_enum,
    sqlc.arg('severity')::bug_report_severity_enum,
    sqlc.arg('description'),
    sqlc.narg('steps'),
    COALESCE(sqlc.narg('debug_info'), '{}'::jsonb)
)
RETURNING *;
