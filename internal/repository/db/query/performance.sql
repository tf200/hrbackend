-- name: ListPerformanceAssessmentCatalog :many
SELECT
    d.code AS domain_code,
    d.name_nl AS domain_name_nl,
    d.name_en AS domain_name_en,
    d.sort_order AS domain_sort_order,
    q.code AS question_code,
    q.title_nl,
    q.title_en,
    q.description_nl,
    q.description_en,
    q.sort_order AS question_sort_order
FROM performance_domains d
JOIN performance_questions q ON q.domain_code = d.code
WHERE d.is_active = TRUE
  AND q.is_active = TRUE
ORDER BY d.sort_order, q.sort_order;

-- name: GetActiveEmployeeNameForPerformance :one
SELECT first_name, last_name
FROM employee_profile
WHERE id = $1
  AND is_archived = FALSE
  AND COALESCE(out_of_service, FALSE) = FALSE;

-- name: CreatePerformanceAssessment :one
INSERT INTO performance_assessments (
    employee_id,
    assessment_date,
    total_score,
    status,
    notes
)
VALUES ($1, $2, $3, 'completed', $4)
RETURNING id, employee_id, assessment_date, total_score, status, notes, created_at;

-- name: GetActivePerformanceQuestion :one
SELECT code, domain_code, title_nl, title_en, description_nl, description_en, sort_order
FROM performance_questions
WHERE code = $1
  AND is_active = TRUE;

-- name: CreatePerformanceAssessmentScore :exec
INSERT INTO performance_assessment_scores (
    assessment_id,
    question_code,
    rating,
    remarks
)
VALUES ($1, $2, $3, $4);

-- name: ListPerformanceAssessments :many
SELECT
    pa.id,
    pa.employee_id,
    ep.first_name,
    ep.last_name,
    pa.assessment_date,
    pa.total_score,
    pa.status,
    pa.notes,
    pa.created_at,
    COUNT(*) OVER() AS total_count
FROM performance_assessments pa
JOIN employee_profile ep ON ep.id = pa.employee_id
WHERE (sqlc.narg(search)::text IS NULL OR LOWER(ep.first_name || ' ' || ep.last_name) LIKE '%' || LOWER(sqlc.narg(search)) || '%')
  AND (sqlc.narg(status)::text IS NULL OR pa.status::text = sqlc.narg(status))
  AND (sqlc.narg(from_date)::date IS NULL OR pa.assessment_date >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::date IS NULL OR pa.assessment_date <= sqlc.narg(to_date))
ORDER BY pa.assessment_date DESC, pa.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPerformanceAssessmentByID :one
SELECT
    pa.id,
    pa.employee_id,
    ep.first_name,
    ep.last_name,
    pa.assessment_date,
    pa.total_score,
    pa.status,
    pa.notes,
    pa.created_at
FROM performance_assessments pa
JOIN employee_profile ep ON ep.id = pa.employee_id
WHERE pa.id = $1;

-- name: DeletePerformanceAssessment :execrows
DELETE FROM performance_assessments WHERE id = $1;

-- name: ListPerformanceAssessmentScores :many
SELECT
    pas.id,
    pas.assessment_id,
    pas.question_code,
    pq.domain_code,
    pq.title_nl,
    pq.title_en,
    pq.description_nl,
    pq.description_en,
    pas.rating,
    pas.remarks
FROM performance_assessment_scores pas
JOIN performance_questions pq ON pq.code = pas.question_code
WHERE pas.assessment_id = $1
ORDER BY pq.domain_code, pq.sort_order;

-- name: ListPerformanceWorkAssignments :many
SELECT
    pwa.id,
    pwa.assessment_id,
    pwa.employee_id,
    ep.first_name,
    ep.last_name,
    pwa.question_code,
    pwa.domain_code,
    pwa.question_text_nl,
    pwa.question_text_en,
    pwa.score,
    pwa.assignment_description,
    pwa.improvement_notes,
    pwa.expectations,
    pwa.advice,
    pwa.due_date,
    pwa.status,
    pwa.submitted_at,
    pwa.submission_text,
    pwa.feedback,
    pwa.reviewed_at,
    COUNT(*) OVER() AS total_count
FROM performance_work_assignments pwa
JOIN employee_profile ep ON ep.id = pwa.employee_id
WHERE (sqlc.narg(employee_id)::uuid IS NULL OR pwa.employee_id = sqlc.narg(employee_id))
  AND (sqlc.narg(status)::text IS NULL OR pwa.status::text = sqlc.narg(status))
  AND (sqlc.narg(due_before)::date IS NULL OR pwa.due_date <= sqlc.narg(due_before))
  AND (sqlc.narg(due_after)::date IS NULL OR pwa.due_date >= sqlc.narg(due_after))
ORDER BY pwa.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPerformanceWorkAssignmentByID :one
SELECT
    pwa.id,
    pwa.assessment_id,
    pwa.employee_id,
    ep.first_name,
    ep.last_name,
    pwa.question_code,
    pwa.domain_code,
    pwa.question_text_nl,
    pwa.question_text_en,
    pwa.score,
    pwa.assignment_description,
    pwa.improvement_notes,
    pwa.expectations,
    pwa.advice,
    pwa.due_date,
    pwa.status,
    pwa.submitted_at,
    pwa.submission_text,
    pwa.feedback,
    pwa.reviewed_at
FROM performance_work_assignments pwa
JOIN employee_profile ep ON ep.id = pwa.employee_id
WHERE pwa.id = $1;

-- name: GetPerformanceWorkAssignmentStatusForUpdate :one
SELECT status::text
FROM performance_work_assignments
WHERE id = $1
FOR UPDATE;

-- name: UpdatePerformanceWorkAssignmentDecision :exec
UPDATE performance_work_assignments
SET status = $2,
    feedback = $3,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ListPerformanceUpcoming :many
WITH employee_review_status AS (
    SELECT
        ep.id,
        ep.first_name,
        ep.last_name,
        ep.contract_start_date,
        (
            SELECT MAX(pa.assessment_date)::date
            FROM performance_assessments pa
            WHERE pa.employee_id = ep.id
              AND pa.status = 'completed'
        ) AS last_assessment_date
    FROM employee_profile ep
    WHERE ep.is_archived = FALSE
      AND COALESCE(ep.out_of_service, FALSE) = FALSE
),
employee_next_dates AS (
    SELECT
        id,
        first_name,
        last_name,
        last_assessment_date,
        CASE
            WHEN last_assessment_date IS NOT NULL
                THEN (last_assessment_date + INTERVAL '42 days')::date
            WHEN contract_start_date IS NOT NULL
                THEN (contract_start_date + INTERVAL '42 days')::date
        END::date AS next_assessment_date
    FROM employee_review_status
)
SELECT
    id,
    first_name,
    last_name,
    last_assessment_date,
    next_assessment_date
FROM employee_next_dates
WHERE next_assessment_date IS NOT NULL
  AND next_assessment_date <= CURRENT_DATE + $1::integer * INTERVAL '1 day'
ORDER BY next_assessment_date ASC;

-- name: GetPerformanceStats :one
WITH active_employees AS (
    SELECT id
    FROM employee_profile
    WHERE is_archived = FALSE
      AND COALESCE(out_of_service, FALSE) = FALSE
),
completed AS (
    SELECT employee_id, assessment_date, total_score
    FROM performance_assessments
    WHERE status = 'completed'
)
SELECT
    (SELECT COUNT(*) FROM active_employees) AS total_employees,
    (SELECT COUNT(*) FROM completed) AS completed_count,
    (SELECT COUNT(*) FROM completed WHERE date_trunc('month', assessment_date) = date_trunc('month', CURRENT_DATE)) AS completed_this_month,
    (SELECT COALESCE(AVG(total_score), 0)::numeric FROM completed) AS average_score,
    (SELECT COUNT(DISTINCT employee_id) FROM completed) AS covered_count;
