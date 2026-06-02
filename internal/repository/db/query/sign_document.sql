-- name: CreateSignDocument :one
INSERT INTO sign_documents (
    title,
    source_attachment_id,
    source_file_key,
    created_by_employee_id,
    related_entity_type,
    related_entity_id,
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CreateSignDocumentRecipient :one
INSERT INTO sign_document_recipients (
    document_id,
    employee_id,
    name,
    email,
    signing_order
)
SELECT
    sqlc.arg('document_id'),
    ep.id,
    NULLIF(TRIM(CONCAT_WS(' ', ep.first_name, ep.last_name)), ''),
    cu.email,
    sqlc.arg('signing_order')
FROM employee_profile ep
LEFT JOIN custom_user cu ON cu.id = ep.user_id
WHERE ep.id = sqlc.arg('employee_id')
RETURNING *;

-- name: GetSignDocumentByID :one
SELECT *
FROM sign_documents
WHERE id = $1
LIMIT 1;

-- name: ListSignDocumentsByCreator :many
SELECT *
FROM sign_documents
WHERE created_by_employee_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSignDocumentsForEmployee :many
SELECT sd.*
FROM sign_documents sd
JOIN sign_document_recipients sdr ON sdr.document_id = sd.id
WHERE sdr.employee_id = $1
ORDER BY sd.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSignDocumentRecipients :many
SELECT *
FROM sign_document_recipients
WHERE document_id = $1
ORDER BY signing_order ASC, created_at ASC;

-- name: GetSignDocumentRecipientForEmployee :one
SELECT *
FROM sign_document_recipients
WHERE document_id = $1 AND employee_id = $2
LIMIT 1;

-- name: DeleteSignDocumentFields :exec
DELETE FROM sign_document_fields
WHERE document_id = $1;

-- name: CreateSignDocumentField :one
INSERT INTO sign_document_fields (
    document_id,
    recipient_id,
    type,
    page_number,
    x,
    y,
    width,
    height,
    required,
    label
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListSignDocumentFields :many
SELECT *
FROM sign_document_fields
WHERE document_id = $1
ORDER BY page_number ASC, created_at ASC;

-- name: ListSignDocumentFieldsForRecipient :many
SELECT *
FROM sign_document_fields
WHERE document_id = $1 AND recipient_id = $2
ORDER BY page_number ASC, created_at ASC;

-- name: SendSignDocument :one
UPDATE sign_documents
SET status = 'sent', sent_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: MarkSignDocumentRecipientViewed :one
UPDATE sign_document_recipients
SET status = CASE WHEN status = 'pending' THEN 'viewed' ELSE status END,
    viewed_at = COALESCE(viewed_at, CURRENT_TIMESTAMP),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('pending', 'viewed')
RETURNING *;

-- name: CountUnsignedPriorSignDocumentRecipients :one
SELECT COUNT(*)::int
FROM sign_document_recipients current_recipient
JOIN sign_document_recipients prior_recipient
  ON prior_recipient.document_id = current_recipient.document_id
 AND prior_recipient.signing_order < current_recipient.signing_order
WHERE current_recipient.id = $1
  AND prior_recipient.status <> 'signed';

-- name: CreateEmployeeSignatureProfile :one
WITH reset_default AS (
    UPDATE employee_signature_profiles
    SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP
    WHERE employee_signature_profiles.employee_id = sqlc.arg('employee_id')
      AND sqlc.arg('is_default')::boolean
)
INSERT INTO employee_signature_profiles (
    employee_id,
    type,
    typed_name,
    image_file_key,
    is_default
)
VALUES (
    sqlc.arg('employee_id'),
    sqlc.arg('type'),
    sqlc.narg('typed_name'),
    sqlc.narg('image_file_key'),
    sqlc.arg('is_default')
)
RETURNING *;

-- name: GetDefaultEmployeeSignatureProfile :one
SELECT *
FROM employee_signature_profiles
WHERE employee_id = $1 AND is_default
LIMIT 1;

-- name: CreateSignDocumentSignature :one
INSERT INTO sign_document_signatures (
    document_id,
    recipient_id,
    employee_id,
    signature_profile_id,
    signature_text,
    signature_image_file_key,
    consent_text,
    ip_address,
    user_agent,
    signature_hash
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateSignDocumentFieldValue :one
UPDATE sign_document_fields
SET value = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND recipient_id = $2
RETURNING *;

-- name: MarkSignDocumentRecipientSigned :one
UPDATE sign_document_recipients
SET status = 'signed', signed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('pending', 'viewed')
RETURNING *;

-- name: CountUnsignedSignDocumentRecipients :one
SELECT COUNT(*)::int
FROM sign_document_recipients
WHERE document_id = $1 AND status <> 'signed';

-- name: MarkSignDocumentPartiallySigned :one
UPDATE sign_documents
SET status = 'partially_signed', updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('sent', 'partially_signed')
RETURNING *;

-- name: MarkSignDocumentCompleted :one
UPDATE sign_documents
SET status = 'completed', signed_file_key = $2, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('sent', 'partially_signed')
RETURNING *;

-- name: CancelSignDocument :one
UPDATE sign_documents
SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status IN ('draft', 'sent', 'partially_signed')
RETURNING *;

-- name: CreateSignDocumentEvent :one
INSERT INTO sign_document_events (
    document_id,
    recipient_id,
    actor_employee_id,
    event,
    ip_address,
    user_agent,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, '{}'::jsonb))
RETURNING *;

-- name: ListSignDocumentEvents :many
SELECT *
FROM sign_document_events
WHERE document_id = $1
ORDER BY created_at DESC;

-- name: ListSignDocumentSignatures :many
SELECT *
FROM sign_document_signatures
WHERE document_id = $1
ORDER BY signed_at ASC;
