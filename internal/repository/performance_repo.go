package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"

	"github.com/google/uuid"
)

type PerformanceRepository struct {
	store *db.Store
}

func NewPerformanceRepository(store *db.Store) domain.PerformanceRepository {
	return &PerformanceRepository{store: store}
}

func (r *PerformanceRepository) CreateAssessment(
	ctx context.Context,
	params domain.CreatePerformanceAssessmentParams,
) (*domain.PerformanceAssessment, error) {
	tx, err := r.store.ConnPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var firstName, lastName string
	err = tx.QueryRow(
		ctx,
		`SELECT first_name, last_name
		 FROM employee_profile
		 WHERE id = $1
		   AND is_archived = FALSE
		   AND COALESCE(out_of_service, FALSE) = FALSE`,
		params.EmployeeID,
	).Scan(&firstName, &lastName)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPerformanceInvalidRequest
		}
		return nil, err
	}

	var created domain.PerformanceAssessment
	created.EmployeeID = params.EmployeeID
	created.EmployeeName = strings.TrimSpace(firstName + " " + lastName)
	err = tx.QueryRow(
		ctx,
		`INSERT INTO performance_assessments (
			employee_id,
			assessment_date,
			total_score,
			status,
			notes
		)
		VALUES ($1, $2, $3, 'completed', $4)
		RETURNING id, assessment_date, total_score, status, notes, created_at`,
		params.EmployeeID,
		params.AssessmentDate,
		averageScore(params.Scores),
		params.Notes,
	).Scan(
		&created.ID,
		&created.AssessmentDate,
		&created.TotalScore,
		&created.Status,
		&created.Notes,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, score := range params.Scores {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO performance_assessment_scores (
				assessment_id,
				domain_id,
				item_id,
				rating,
				remarks
			)
			VALUES ($1, $2, $3, $4, $5)`,
			created.ID,
			score.DomainID,
			score.ItemID,
			score.Rating,
			score.Remarks,
		)
		if err != nil {
			return nil, err
		}

		if score.Rating <= 5 {
			_, err = tx.Exec(
				ctx,
				`INSERT INTO performance_work_assignments (
					assessment_id,
					employee_id,
					question_id,
					domain_id,
					question_text,
					score,
					assignment_description,
					due_date,
					status
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, ($8::date + INTERVAL '14 day')::date, 'open')`,
				created.ID,
				params.EmployeeID,
				score.ItemID,
				score.DomainID,
				formatQuestionText(score.ItemID),
				score.Rating,
				formatAssignmentDescription(score.DomainID, score.ItemID, score.Rating),
				params.AssessmentDate,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *PerformanceRepository) ListAssessments(
	ctx context.Context,
	params domain.ListPerformanceAssessmentsParams,
) (*domain.PerformanceAssessmentPage, error) {
	rows, err := r.store.ConnPool.Query(
		ctx,
		`SELECT
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
		 WHERE ($1::uuid IS NULL OR pa.employee_id = $1)
		   AND ($2::text IS NULL OR pa.status::text = $2)
		   AND ($3::date IS NULL OR pa.assessment_date >= $3)
		   AND ($4::date IS NULL OR pa.assessment_date <= $4)
		 ORDER BY pa.assessment_date DESC, pa.created_at DESC
		 LIMIT $5 OFFSET $6`,
		params.EmployeeID,
		params.Status,
		params.FromDate,
		params.ToDate,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PerformanceAssessment, 0)
	var totalCount int64
	for rows.Next() {
		var item domain.PerformanceAssessment
		var firstName, lastName string
		if err := rows.Scan(
			&item.ID,
			&item.EmployeeID,
			&firstName,
			&lastName,
			&item.AssessmentDate,
			&item.TotalScore,
			&item.Status,
			&item.Notes,
			&item.CreatedAt,
			&totalCount,
		); err != nil {
			return nil, err
		}
		item.EmployeeName = strings.TrimSpace(firstName + " " + lastName)
		items = append(items, item)
	}

	return &domain.PerformanceAssessmentPage{Items: items, TotalCount: totalCount}, rows.Err()
}

func (r *PerformanceRepository) GetAssessmentByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.PerformanceAssessment, error) {
	var item domain.PerformanceAssessment
	var firstName, lastName string
	err := r.store.ConnPool.QueryRow(
		ctx,
		`SELECT
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
		 WHERE pa.id = $1`,
		id,
	).Scan(
		&item.ID,
		&item.EmployeeID,
		&firstName,
		&lastName,
		&item.AssessmentDate,
		&item.TotalScore,
		&item.Status,
		&item.Notes,
		&item.CreatedAt,
	)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPerformanceNotFound
		}
		return nil, err
	}
	item.EmployeeName = strings.TrimSpace(firstName + " " + lastName)
	return &item, nil
}

func (r *PerformanceRepository) DeleteAssessment(ctx context.Context, id uuid.UUID) (bool, error) {
	result, err := r.store.ConnPool.Exec(
		ctx,
		`DELETE FROM performance_assessments WHERE id = $1`,
		id,
	)
	if err != nil {
		return false, err
	}
	if result.RowsAffected() == 0 {
		return false, domain.ErrPerformanceNotFound
	}
	return true, nil
}

func (r *PerformanceRepository) ListAssessmentScores(
	ctx context.Context,
	assessmentID uuid.UUID,
) ([]domain.PerformanceAssessmentScore, error) {
	rows, err := r.store.ConnPool.Query(
		ctx,
		`SELECT id, assessment_id, domain_id, item_id, rating, remarks
		 FROM performance_assessment_scores
		 WHERE assessment_id = $1
		 ORDER BY domain_id, item_id`,
		assessmentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PerformanceAssessmentScore, 0)
	for rows.Next() {
		var item domain.PerformanceAssessmentScore
		if err := rows.Scan(
			&item.ID,
			&item.AssessmentID,
			&item.DomainID,
			&item.ItemID,
			&item.Rating,
			&item.Remarks,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PerformanceRepository) ListWorkAssignments(
	ctx context.Context,
	params domain.ListPerformanceWorkAssignmentsParams,
) (*domain.PerformanceWorkAssignmentPage, error) {
	rows, err := r.store.ConnPool.Query(
		ctx,
		`SELECT
			pwa.id,
			pwa.assessment_id,
			pwa.employee_id,
			ep.first_name,
			ep.last_name,
			pwa.question_id,
			pwa.domain_id,
			pwa.question_text,
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
		 WHERE ($1::uuid IS NULL OR pwa.employee_id = $1)
		   AND ($2::text IS NULL OR pwa.status::text = $2)
		   AND ($3::date IS NULL OR pwa.due_date <= $3)
		   AND ($4::date IS NULL OR pwa.due_date >= $4)
		 ORDER BY pwa.created_at DESC
		 LIMIT $5 OFFSET $6`,
		params.EmployeeID,
		params.Status,
		params.DueBefore,
		params.DueAfter,
		params.Limit,
		params.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PerformanceWorkAssignment, 0)
	var totalCount int64
	for rows.Next() {
		var item domain.PerformanceWorkAssignment
		var firstName, lastName string
		if err := rows.Scan(
			&item.ID,
			&item.AssessmentID,
			&item.EmployeeID,
			&firstName,
			&lastName,
			&item.QuestionID,
			&item.DomainID,
			&item.QuestionText,
			&item.Score,
			&item.AssignmentDescription,
			&item.ImprovementNotes,
			&item.Expectations,
			&item.Advice,
			&item.DueDate,
			&item.Status,
			&item.SubmittedAt,
			&item.SubmissionText,
			&item.Feedback,
			&item.ReviewedAt,
			&totalCount,
		); err != nil {
			return nil, err
		}
		item.EmployeeName = strings.TrimSpace(firstName + " " + lastName)
		items = append(items, item)
	}

	return &domain.PerformanceWorkAssignmentPage{Items: items, TotalCount: totalCount}, rows.Err()
}

func (r *PerformanceRepository) GetWorkAssignmentByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.PerformanceWorkAssignment, error) {
	var item domain.PerformanceWorkAssignment
	var firstName, lastName string
	err := r.store.ConnPool.QueryRow(
		ctx,
		`SELECT
			pwa.id,
			pwa.assessment_id,
			pwa.employee_id,
			ep.first_name,
			ep.last_name,
			pwa.question_id,
			pwa.domain_id,
			pwa.question_text,
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
		 WHERE pwa.id = $1`,
		id,
	).Scan(
		&item.ID,
		&item.AssessmentID,
		&item.EmployeeID,
		&firstName,
		&lastName,
		&item.QuestionID,
		&item.DomainID,
		&item.QuestionText,
		&item.Score,
		&item.AssignmentDescription,
		&item.ImprovementNotes,
		&item.Expectations,
		&item.Advice,
		&item.DueDate,
		&item.Status,
		&item.SubmittedAt,
		&item.SubmissionText,
		&item.Feedback,
		&item.ReviewedAt,
	)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPerformanceNotFound
		}
		return nil, err
	}
	item.EmployeeName = strings.TrimSpace(firstName + " " + lastName)
	return &item, nil
}

func (r *PerformanceRepository) DecideWorkAssignment(
	ctx context.Context,
	id uuid.UUID,
	params domain.DecidePerformanceWorkAssignmentParams,
) (*domain.PerformanceWorkAssignment, error) {
	tx, err := r.store.ConnPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var currentStatus string
	err = tx.QueryRow(
		ctx,
		`SELECT status::text FROM performance_work_assignments WHERE id = $1 FOR UPDATE`,
		id,
	).Scan(&currentStatus)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPerformanceNotFound
		}
		return nil, err
	}

	if currentStatus != domain.PerformanceWorkAssignmentStatusSubmitted {
		return nil, domain.ErrPerformanceStateInvalid
	}

	nextStatus := domain.PerformanceWorkAssignmentStatusApproved
	if params.Decision == "request_revision" {
		nextStatus = domain.PerformanceWorkAssignmentStatusRevisionNeeded
	}

	_, err = tx.Exec(
		ctx,
		`UPDATE performance_work_assignments
		 SET status = $2,
		     feedback = $3,
		     reviewed_at = NOW(),
		     updated_at = NOW()
		 WHERE id = $1`,
		id,
		nextStatus,
		params.Feedback,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetWorkAssignmentByID(ctx, id)
}

func (r *PerformanceRepository) ListUpcoming(
	ctx context.Context,
	windowDays int,
) ([]domain.PerformanceUpcomingItem, error) {
	rows, err := r.store.ConnPool.Query(
		ctx,
		`WITH last_completed AS (
			SELECT
				employee_id,
				MAX(assessment_date) AS last_assessment_date
			FROM performance_assessments
			WHERE status = 'completed'
			GROUP BY employee_id
		)
		SELECT
			ep.id,
			ep.first_name,
			ep.last_name,
			lc.last_assessment_date,
			(lc.last_assessment_date + INTERVAL '42 day')::date AS next_assessment_date
		FROM employee_profile ep
		JOIN last_completed lc ON lc.employee_id = ep.id
		WHERE ep.is_archived = FALSE
		  AND COALESCE(ep.out_of_service, FALSE) = FALSE
		ORDER BY next_assessment_date ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := dateOnlyUTCPerf(time.Now().UTC())
	items := make([]domain.PerformanceUpcomingItem, 0)
	for rows.Next() {
		var employeeID uuid.UUID
		var firstName, lastName string
		var lastDate, nextDate time.Time
		if err := rows.Scan(&employeeID, &firstName, &lastName, &lastDate, &nextDate); err != nil {
			return nil, err
		}
		days := int(nextDate.Sub(now).Hours() / 24)
		isOverdue := nextDate.Before(now)
		isDueSoon := !isOverdue && days <= windowDays
		items = append(items, domain.PerformanceUpcomingItem{
			EmployeeID:         employeeID,
			EmployeeName:       strings.TrimSpace(firstName + " " + lastName),
			LastAssessmentDate: lastDate,
			NextAssessmentDate: nextDate,
			IsOverdue:          isOverdue,
			IsDueSoon:          isDueSoon,
			DaysUntilDue:       days,
		})
	}

	return items, rows.Err()
}

func (r *PerformanceRepository) GetStats(ctx context.Context) (*domain.PerformanceStats, error) {
	stats := &domain.PerformanceStats{}

	err := r.store.ConnPool.QueryRow(
		ctx,
		`WITH active_employees AS (
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
			(SELECT AVG(total_score) FROM completed) AS average_score,
			(SELECT COUNT(DISTINCT employee_id) FROM completed) AS covered_count`,
	).Scan(
		&stats.TotalEmployees,
		&stats.CompletedCount,
		&stats.CompletedThisMonth,
		&stats.AverageScore,
		&stats.CoveredCount,
	)
	if err != nil {
		return nil, err
	}

	if stats.TotalEmployees > 0 {
		stats.CoveragePercent = int32((stats.CoveredCount * 100) / stats.TotalEmployees)
	}

	return stats, nil
}

func averageScore(scores []domain.CreatePerformanceAssessmentScoreParams) *float64 {
	if len(scores) == 0 {
		return nil
	}
	var sum float64
	for _, score := range scores {
		sum += score.Rating
	}
	result := sum / float64(len(scores))
	return &result
}

func formatQuestionText(itemID string) string {
	if itemID == "" {
		return "Reflectiepunt"
	}
	return fmt.Sprintf("Reflectiepunt %s", itemID)
}

func formatAssignmentDescription(domainID, itemID string, score float64) string {
	return fmt.Sprintf(
		"Score %.1f op %s (%s). Beschrijf verbeteracties en concrete opvolgstappen.",
		score,
		domainID,
		itemID,
	)
}

func dateOnlyUTCPerf(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
