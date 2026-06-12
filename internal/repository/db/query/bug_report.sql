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

-- name: UpdateBugReportTrelloCard :one
UPDATE bug_reports
SET
    trello_card_id = sqlc.arg('trello_card_id'),
    trello_card_url = sqlc.arg('trello_card_url'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
