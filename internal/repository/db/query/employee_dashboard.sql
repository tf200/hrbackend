-- name: GetEmployeeDashboardKPIs :one
WITH input AS (
    SELECT
        sqlc.arg('employee_id')::uuid AS employee_id,
        sqlc.arg('year')::int AS year
), employee_row AS (
    SELECT ep.id
    FROM employee_profile ep
    JOIN input ON input.employee_id = ep.id
    WHERE ep.is_archived = FALSE
      AND COALESCE(ep.out_of_service, FALSE) = FALSE
)
SELECT
    input.year::int AS year,
    calculate_legal_leave_minutes(
        input.employee_id,
        input.year,
        CASE
            WHEN input.year < EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                (make_date(input.year, 12, 31)::timestamp + INTERVAL '23 hours 59 minutes 59 seconds')::timestamptz
            WHEN input.year = EXTRACT(YEAR FROM CURRENT_DATE)::int THEN
                CURRENT_TIMESTAMP
            ELSE
                make_date(input.year, 1, 1)::timestamptz
        END
    )::int AS leave_total_minutes,
    (COALESCE(leave_used.used_minutes, 0) + COALESCE(payout_used.used_minutes, 0))::int AS leave_used_minutes,
    COALESCE(pending_leave.pending_count, 0)::bigint AS pending_leave_requests,
    COALESCE(pending_signatures.pending_count, 0)::bigint AS pending_signatures
FROM input
JOIN employee_row ON employee_row.id = input.employee_id
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(lr.requested_minutes), 0)::int AS used_minutes
    FROM leave_requests lr
    JOIN leave_policies lp ON lp.leave_type = lr.leave_type
    WHERE lr.employee_id = input.employee_id
      AND lr.status = 'approved'::leave_request_status_enum
      AND lp.deducts_balance = TRUE
      AND lr.start_date >= make_date(input.year, 1, 1)
      AND lr.start_date < make_date(input.year + 1, 1, 1)
) leave_used ON TRUE
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(lpr.requested_hours * 60), 0)::int AS used_minutes
    FROM leave_payout_requests lpr
    WHERE lpr.employee_id = input.employee_id
      AND lpr.balance_year = input.year
      AND lpr.status IN (
          'approved'::payout_request_status_enum,
          'paid'::payout_request_status_enum
      )
) payout_used ON TRUE
LEFT JOIN LATERAL (
    SELECT COUNT(*)::bigint AS pending_count
    FROM leave_requests lr
    WHERE lr.employee_id = input.employee_id
      AND lr.status = 'pending'::leave_request_status_enum
      AND lr.start_date < make_date(input.year + 1, 1, 1)
      AND lr.end_date >= make_date(input.year, 1, 1)
) pending_leave ON TRUE
LEFT JOIN LATERAL (
    SELECT COUNT(*)::bigint AS pending_count
    FROM sign_document_recipients sdr
    JOIN sign_documents sd ON sd.id = sdr.document_id
    WHERE sdr.employee_id = input.employee_id
      AND sdr.status IN (
          'pending'::sign_document_recipient_status_enum,
          'viewed'::sign_document_recipient_status_enum
      )
      AND sd.status IN (
          'sent'::sign_document_status_enum,
          'partially_signed'::sign_document_status_enum
      )
      AND (sd.expires_at IS NULL OR sd.expires_at > CURRENT_TIMESTAMP)
) pending_signatures ON TRUE;

-- name: ListEmployeeDashboardPendingRequests :many
WITH pending_requests AS (
    SELECT
        lr.id,
        'leave'::text AS request_type,
        lr.status::text AS status,
        lr.requested_at AS submitted_at,
        lr.start_date AS request_date,
        lr.leave_type::text AS title,
        lr.reason AS description,
        lr.requested_minutes::int AS duration_minutes,
        NULL::double precision AS amount,
        NULL::text AS currency
    FROM leave_requests lr
    WHERE lr.employee_id = sqlc.arg('employee_id')
      AND lr.status = 'pending'::leave_request_status_enum
      AND lr.requested_at >= sqlc.arg('since')::timestamptz

    UNION ALL

    SELECT
        oe.id,
        'overtime'::text AS request_type,
        oe.status::text AS status,
        oe.submitted_at AS submitted_at,
        oe.entry_date AS request_date,
        oe.reason::text AS title,
        oe.description AS description,
        oe.minutes::int AS duration_minutes,
        NULL::double precision AS amount,
        NULL::text AS currency
    FROM overtime_entries oe
    WHERE oe.employee_id = sqlc.arg('employee_id')
      AND oe.status = 'submitted'::overtime_status_enum
      AND oe.submitted_at >= sqlc.arg('since')::timestamptz

    UNION ALL

    SELECT
        er.id,
        'expense'::text AS request_type,
        er.status::text AS status,
        er.requested_at AS submitted_at,
        er.expense_date AS request_date,
        er.category::text AS title,
        NULLIF(er.description, '') AS description,
        0::int AS duration_minutes,
        er.claimed_amount::double precision AS amount,
        er.currency::text AS currency
    FROM expense_requests er
    WHERE er.employee_id = sqlc.arg('employee_id')
      AND er.status = 'pending'::expense_request_status_enum
      AND er.requested_at >= sqlc.arg('since')::timestamptz
)
SELECT *
FROM pending_requests
ORDER BY submitted_at DESC, id DESC
LIMIT sqlc.arg('limit');
