package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PerformanceRepository struct {
	store *db.Store
}

func NewPerformanceRepository(store *db.Store) domain.PerformanceRepository {
	return &PerformanceRepository{store: store}
}

func (r *PerformanceRepository) ListAssessmentCatalog(ctx context.Context) ([]domain.PerformanceDomain, error) {
	rows, err := r.store.ListPerformanceAssessmentCatalog(ctx)
	if err != nil {
		return nil, err
	}

	domains := make([]domain.PerformanceDomain, 0)
	domainByCode := make(map[string]int)
	for _, row := range rows {
		idx, exists := domainByCode[row.DomainCode]
		if !exists {
			idx = len(domains)
			domainByCode[row.DomainCode] = idx
			domains = append(domains, domain.PerformanceDomain{
				Code:      row.DomainCode,
				NameNL:    row.DomainNameNl,
				NameEN:    row.DomainNameEn,
				SortOrder: row.DomainSortOrder,
				Questions: make([]domain.PerformanceQuestion, 0, 5),
			})
		}

		domains[idx].Questions = append(domains[idx].Questions, domain.PerformanceQuestion{
			Code:          row.QuestionCode,
			DomainCode:    row.DomainCode,
			TitleNL:       row.TitleNl,
			TitleEN:       row.TitleEn,
			DescriptionNL: row.DescriptionNl,
			DescriptionEN: row.DescriptionEn,
			SortOrder:     row.QuestionSortOrder,
		})
	}

	return domains, nil
}

func (r *PerformanceRepository) CreateAssessment(
	ctx context.Context,
	params domain.CreatePerformanceAssessmentParams,
) (*domain.PerformanceAssessment, error) {
	var created *domain.PerformanceAssessment
	err := r.store.ExecTx(ctx, func(q *db.Queries) error {
		employee, err := q.GetActiveEmployeeNameForPerformance(ctx, params.EmployeeID)
		if err != nil {
			if isDBNotFound(err) {
				return domain.ErrPerformanceInvalidRequest
			}
			return err
		}

		row, err := q.CreatePerformanceAssessment(ctx, db.CreatePerformanceAssessmentParams{
			EmployeeID:     params.EmployeeID,
			AssessmentDate: conv.PgDateFromTime(params.AssessmentDate),
			TotalScore:     averageScore(params.Scores),
			Notes:          params.Notes,
		})
		if err != nil {
			return err
		}

		assessment := toDomainPerformanceAssessment(
			row.ID,
			row.EmployeeID,
			employee.FirstName,
			employee.LastName,
			row.AssessmentDate,
			row.TotalScore,
			string(row.Status),
			row.Notes,
			row.CreatedAt,
		)

		for _, score := range params.Scores {
			question, err := q.GetActivePerformanceQuestion(ctx, score.QuestionCode)
			if err != nil {
				if isDBNotFound(err) {
					return domain.ErrPerformanceInvalidRequest
				}
				return err
			}

			err = q.CreatePerformanceAssessmentScore(ctx, db.CreatePerformanceAssessmentScoreParams{
				AssessmentID: assessment.ID,
				QuestionCode: score.QuestionCode,
				Rating:       score.Rating,
				Remarks:      score.Remarks,
			})
			if err != nil {
				return err
			}

			if score.Rating <= 5 {
				err = q.CreatePerformanceWorkAssignment(ctx, db.CreatePerformanceWorkAssignmentParams{
					AssessmentID:          assessment.ID,
					EmployeeID:            params.EmployeeID,
					QuestionCode:          question.Code,
					DomainCode:            question.DomainCode,
					QuestionTextNl:        question.TitleNl,
					QuestionTextEn:        question.TitleEn,
					Score:                 score.Rating,
					AssignmentDescription: formatAssignmentDescription(question.DomainCode, question.Code, score.Rating),
					AssessmentDate:        conv.PgDateFromTime(params.AssessmentDate),
				})
				if err != nil {
					return err
				}
			}
		}

		created = &assessment
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *PerformanceRepository) ListAssessments(
	ctx context.Context,
	params domain.ListPerformanceAssessmentsParams,
) (*domain.PerformanceAssessmentPage, error) {
	rows, err := r.store.ListPerformanceAssessments(ctx, db.ListPerformanceAssessmentsParams{
		Limit:    params.Limit,
		Offset:   params.Offset,
		Search:   params.Search,
		Status:   params.Status,
		FromDate: pgDateFromTimePtr(params.FromDate),
		ToDate:   pgDateFromTimePtr(params.ToDate),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.PerformanceAssessment, 0, len(rows))
	var totalCount int64
	for _, row := range rows {
		items = append(items, toDomainPerformanceAssessment(
			row.ID,
			row.EmployeeID,
			row.FirstName,
			row.LastName,
			row.AssessmentDate,
			row.TotalScore,
			string(row.Status),
			row.Notes,
			row.CreatedAt,
		))
		totalCount = row.TotalCount
	}

	return &domain.PerformanceAssessmentPage{Items: items, TotalCount: totalCount}, nil
}

func (r *PerformanceRepository) GetAssessmentByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.PerformanceAssessment, error) {
	row, err := r.store.GetPerformanceAssessmentByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPerformanceNotFound
		}
		return nil, err
	}

	item := toDomainPerformanceAssessment(
		row.ID,
		row.EmployeeID,
		row.FirstName,
		row.LastName,
		row.AssessmentDate,
		row.TotalScore,
		string(row.Status),
		row.Notes,
		row.CreatedAt,
	)
	return &item, nil
}

func (r *PerformanceRepository) DeleteAssessment(ctx context.Context, id uuid.UUID) (bool, error) {
	rowsAffected, err := r.store.DeletePerformanceAssessment(ctx, id)
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, domain.ErrPerformanceNotFound
	}
	return true, nil
}

func (r *PerformanceRepository) ListAssessmentScores(
	ctx context.Context,
	assessmentID uuid.UUID,
) ([]domain.PerformanceAssessmentScore, error) {
	rows, err := r.store.ListPerformanceAssessmentScores(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PerformanceAssessmentScore, len(rows))
	for i, row := range rows {
		items[i] = domain.PerformanceAssessmentScore{
			ID:            row.ID,
			AssessmentID:  row.AssessmentID,
			QuestionCode:  row.QuestionCode,
			DomainCode:    row.DomainCode,
			TitleNL:       row.TitleNl,
			TitleEN:       row.TitleEn,
			DescriptionNL: row.DescriptionNl,
			DescriptionEN: row.DescriptionEn,
			Rating:        row.Rating,
			Remarks:       row.Remarks,
		}
	}

	return items, nil
}

func (r *PerformanceRepository) ListWorkAssignments(
	ctx context.Context,
	params domain.ListPerformanceWorkAssignmentsParams,
) (*domain.PerformanceWorkAssignmentPage, error) {
	rows, err := r.store.ListPerformanceWorkAssignments(ctx, db.ListPerformanceWorkAssignmentsParams{
		Limit:      params.Limit,
		Offset:     params.Offset,
		EmployeeID: params.EmployeeID,
		Status:     params.Status,
		DueBefore:  pgDateFromTimePtr(params.DueBefore),
		DueAfter:   pgDateFromTimePtr(params.DueAfter),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.PerformanceWorkAssignment, 0, len(rows))
	var totalCount int64
	for _, row := range rows {
		items = append(items, toDomainPerformanceWorkAssignment(
			row.ID,
			row.AssessmentID,
			row.EmployeeID,
			row.FirstName,
			row.LastName,
			row.QuestionCode,
			row.DomainCode,
			row.QuestionTextNl,
			row.QuestionTextEn,
			row.Score,
			row.AssignmentDescription,
			row.ImprovementNotes,
			row.Expectations,
			row.Advice,
			row.DueDate,
			string(row.Status),
			row.SubmittedAt,
			row.SubmissionText,
			row.Feedback,
			row.ReviewedAt,
		))
		totalCount = row.TotalCount
	}

	return &domain.PerformanceWorkAssignmentPage{Items: items, TotalCount: totalCount}, nil
}

func (r *PerformanceRepository) GetWorkAssignmentByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.PerformanceWorkAssignment, error) {
	row, err := r.store.GetPerformanceWorkAssignmentByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPerformanceNotFound
		}
		return nil, err
	}

	item := toDomainPerformanceWorkAssignment(
		row.ID,
		row.AssessmentID,
		row.EmployeeID,
		row.FirstName,
		row.LastName,
		row.QuestionCode,
		row.DomainCode,
		row.QuestionTextNl,
		row.QuestionTextEn,
		row.Score,
		row.AssignmentDescription,
		row.ImprovementNotes,
		row.Expectations,
		row.Advice,
		row.DueDate,
		string(row.Status),
		row.SubmittedAt,
		row.SubmissionText,
		row.Feedback,
		row.ReviewedAt,
	)
	return &item, nil
}

func (r *PerformanceRepository) DecideWorkAssignment(
	ctx context.Context,
	id uuid.UUID,
	params domain.DecidePerformanceWorkAssignmentParams,
) (*domain.PerformanceWorkAssignment, error) {
	err := r.store.ExecTx(ctx, func(q *db.Queries) error {
		currentStatus, err := q.GetPerformanceWorkAssignmentStatusForUpdate(ctx, id)
		if err != nil {
			if isDBNotFound(err) {
				return domain.ErrPerformanceNotFound
			}
			return err
		}

		if currentStatus != domain.PerformanceWorkAssignmentStatusSubmitted {
			return domain.ErrPerformanceStateInvalid
		}

		nextStatus := db.PerformanceWorkAssignmentStatusEnumApproved
		if params.Decision == "request_revision" {
			nextStatus = db.PerformanceWorkAssignmentStatusEnumRevisionNeeded
		}

		return q.UpdatePerformanceWorkAssignmentDecision(ctx, db.UpdatePerformanceWorkAssignmentDecisionParams{
			ID:       id,
			Status:   nextStatus,
			Feedback: params.Feedback,
		})
	})
	if err != nil {
		return nil, err
	}

	return r.GetWorkAssignmentByID(ctx, id)
}

func (r *PerformanceRepository) ListUpcoming(
	ctx context.Context,
	windowDays int,
) ([]domain.PerformanceUpcomingItem, error) {
	rows, err := r.store.ListPerformanceUpcoming(ctx, int32(windowDays))
	if err != nil {
		return nil, err
	}

	now := dateOnlyUTCPerf(time.Now().UTC())
	items := make([]domain.PerformanceUpcomingItem, 0, len(rows))
	for _, row := range rows {
		nextDate := conv.TimeFromPgDate(row.NextAssessmentDate)
		days := int(nextDate.Sub(now).Hours() / 24)
		isOverdue := nextDate.Before(now)
		items = append(items, domain.PerformanceUpcomingItem{
			EmployeeID:         row.ID,
			EmployeeName:       strings.TrimSpace(row.FirstName + " " + row.LastName),
			LastAssessmentDate: conv.TimePtrFromPgDate(row.LastAssessmentDate),
			NextAssessmentDate: nextDate,
			IsOverdue:          isOverdue,
			IsDueSoon:          !isOverdue && days <= windowDays,
			DaysUntilDue:       days,
			IsFirstReview:      !row.LastAssessmentDate.Valid,
		})
	}

	return items, nil
}

func (r *PerformanceRepository) GetStats(ctx context.Context) (*domain.PerformanceStats, error) {
	row, err := r.store.GetPerformanceStats(ctx)
	if err != nil {
		return nil, err
	}

	averageScore := row.AverageScore
	stats := &domain.PerformanceStats{
		TotalEmployees:     row.TotalEmployees,
		CompletedCount:     row.CompletedCount,
		CompletedThisMonth: row.CompletedThisMonth,
		AverageScore:       &averageScore,
		CoveredCount:       row.CoveredCount,
	}
	if stats.TotalEmployees > 0 {
		stats.CoveragePercent = int32((stats.CoveredCount * 100) / stats.TotalEmployees)
	}

	return stats, nil
}

func toDomainPerformanceAssessment(
	id uuid.UUID,
	employeeID uuid.UUID,
	firstName string,
	lastName string,
	assessmentDate pgtype.Date,
	totalScore *float64,
	status string,
	notes *string,
	createdAt pgtype.Timestamptz,
) domain.PerformanceAssessment {
	return domain.PerformanceAssessment{
		ID:             id,
		EmployeeID:     employeeID,
		EmployeeName:   strings.TrimSpace(firstName + " " + lastName),
		AssessmentDate: conv.TimeFromPgDate(assessmentDate),
		TotalScore:     totalScore,
		Status:         status,
		Notes:          notes,
		CreatedAt:      conv.TimeFromPgTimestamptz(createdAt),
	}
}

func toDomainPerformanceWorkAssignment(
	id uuid.UUID,
	assessmentID uuid.UUID,
	employeeID uuid.UUID,
	firstName string,
	lastName string,
	questionCode string,
	domainCode string,
	questionTextNL string,
	questionTextEN string,
	score float64,
	assignmentDescription string,
	improvementNotes *string,
	expectations *string,
	advice *string,
	dueDate pgtype.Date,
	status string,
	submittedAt pgtype.Timestamptz,
	submissionText *string,
	feedback *string,
	reviewedAt pgtype.Timestamptz,
) domain.PerformanceWorkAssignment {
	return domain.PerformanceWorkAssignment{
		ID:                    id,
		AssessmentID:          assessmentID,
		EmployeeID:            employeeID,
		EmployeeName:          strings.TrimSpace(firstName + " " + lastName),
		QuestionCode:          questionCode,
		DomainCode:            domainCode,
		QuestionTextNL:        questionTextNL,
		QuestionTextEN:        questionTextEN,
		Score:                 score,
		AssignmentDescription: assignmentDescription,
		ImprovementNotes:      improvementNotes,
		Expectations:          expectations,
		Advice:                advice,
		DueDate:               conv.TimePtrFromPgDate(dueDate),
		Status:                status,
		SubmittedAt:           timePtrFromPgTimestamptz(submittedAt),
		SubmissionText:        submissionText,
		Feedback:              feedback,
		ReviewedAt:            timePtrFromPgTimestamptz(reviewedAt),
	}
}

func pgDateFromTimePtr(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return conv.PgDateFromTime(*value)
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

func formatAssignmentDescription(domainCode, questionCode string, score float64) string {
	return fmt.Sprintf(
		"Score %.1f op %s (%s). Beschrijf verbeteracties en concrete opvolgstappen.",
		score,
		domainCode,
		questionCode,
	)
}

func dateOnlyUTCPerf(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
