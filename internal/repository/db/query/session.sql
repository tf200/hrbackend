-- name: CreateSession :one
INSERT INTO sessions (
    id,
    refresh_token,
    user_agent,
    client_ip,
    is_blocked,
    expires_at,
    created_at,
    user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: ListActiveSessionsByUserID :many
SELECT id, user_agent, client_ip, is_blocked, expires_at, created_at
FROM sessions
WHERE user_id = $1
  AND is_blocked = FALSE
  AND expires_at > NOW()
ORDER BY created_at DESC;
