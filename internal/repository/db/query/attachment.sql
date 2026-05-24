-- name: CreateAttachment :one
INSERT INTO attachment_file (
    name,
    file,
    size,
    is_used,
    tag
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetAttachment :one
SELECT * FROM attachment_file
WHERE uuid = $1 LIMIT 1;

-- name: UpdateAttachmentUsed :one
UPDATE attachment_file
SET is_used = $2,
    updated = CURRENT_TIMESTAMP
WHERE uuid = $1
RETURNING *;

-- name: DeleteAttachment :exec
DELETE FROM attachment_file
WHERE uuid = $1;

-- name: CreateEmployeeAttachments :exec
INSERT INTO employee_attachment (
    employee_id,
    attachment_id,
    category
)
SELECT 
    sqlc.arg(employee_id)::uuid,
    unnest(sqlc.arg(attachment_ids)::uuid[]),
    sqlc.arg(category)::varchar
;

-- name: UpdateAttachmentsUsed :exec
UPDATE attachment_file
SET is_used = sqlc.arg(is_used)::boolean,
    updated = CURRENT_TIMESTAMP
WHERE uuid = ANY(sqlc.arg(attachment_ids)::uuid[]);

-- name: ListEmployeeAttachments :many
SELECT 
    ea.id,
    ea.employee_id,
    ea.attachment_id,
    ea.category,
    ea.created_at,
    ea.updated_at,
    af.name,
    af.file,
    af.size,
    af.tag
FROM employee_attachment ea
JOIN attachment_file af ON af.uuid = ea.attachment_id
WHERE ea.employee_id = $1;

