package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PerformanceService struct {
	repository domain.PerformanceRepository
	logger     domain.Logger
}

func NewPerformanceService(
	repository domain.PerformanceRepository,
	logger domain.Logger,
) domain.PerformanceService {
	return &PerformanceService{repository: repository, logger: logger}
}

func (s *PerformanceService) ListAssessmentCatalog(ctx context.Context) ([]domain.PerformanceDomain, error) {
	items, err := s.repository.ListAssessmentCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assessment catalog: %w", err)
	}
	return items, nil
}

func (s *PerformanceService) CreateAssessment(
	ctx context.Context,
	params domain.CreatePerformanceAssessmentParams,
) (*domain.PerformanceAssessment, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPerformanceInvalidRequest
	}
	if len(params.Scores) == 0 {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	if params.AssessmentDate.IsZero() {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	normalized := params
	normalized.AssessmentDate = performanceDateOnlyUTC(params.AssessmentDate)
	normalized.Notes = performanceTrimStringPtr(params.Notes)

	catalog, err := s.repository.ListAssessmentCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("list assessment catalog: %w", err)
	}

	activeQuestions := make(map[string]struct{})
	for _, domainItem := range catalog {
		for _, question := range domainItem.Questions {
			activeQuestions[question.Code] = struct{}{}
		}
	}
	if len(activeQuestions) == 0 {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	seen := make(map[string]struct{}, len(params.Scores))
	for i, score := range params.Scores {
		questionCode := strings.TrimSpace(score.QuestionCode)
		if questionCode == "" {
			return nil, domain.ErrPerformanceInvalidRequest
		}
		if _, exists := activeQuestions[questionCode]; !exists {
			return nil, domain.ErrPerformanceInvalidRequest
		}
		if score.Rating < 1 || score.Rating > 10 {
			return nil, domain.ErrPerformanceInvalidRequest
		}
		if _, exists := seen[questionCode]; exists {
			return nil, domain.ErrPerformanceInvalidRequest
		}
		seen[questionCode] = struct{}{}

		normalized.Scores[i].QuestionCode = questionCode
		normalized.Scores[i].Remarks = performanceTrimStringPtr(score.Remarks)
	}
	if len(seen) != len(activeQuestions) {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	assessment, err := s.repository.CreateAssessment(ctx, normalized)
	if err != nil {
		s.logger.LogError(
			ctx,
			"PerformanceService.CreateAssessment",
			"failed to create assessment",
			err,
			zap.String("employee_id", params.EmployeeID.String()),
		)
		return nil, fmt.Errorf("create assessment: %w", err)
	}

	return assessment, nil
}

func (s *PerformanceService) ListAssessments(
	ctx context.Context,
	params domain.ListPerformanceAssessmentsParams,
) (*domain.PerformanceAssessmentPage, error) {
	normalized, err := normalizeListAssessmentsParams(params)
	if err != nil {
		return nil, err
	}

	page, err := s.repository.ListAssessments(ctx, normalized)
	if err != nil {
		s.logger.LogError(
			ctx,
			"PerformanceService.ListAssessments",
			"failed to list assessments",
			err,
		)
		return nil, fmt.Errorf("list assessments: %w", err)
	}

	return page, nil
}

func (s *PerformanceService) GetAssessmentByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.PerformanceAssessment, error) {
	if id == uuid.Nil {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	item, err := s.repository.GetAssessmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *PerformanceService) DeleteAssessment(ctx context.Context, id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, domain.ErrPerformanceInvalidRequest
	}

	deleted, err := s.repository.DeleteAssessment(ctx, id)
	if err != nil {
		return false, err
	}
	return deleted, nil
}

func (s *PerformanceService) ListAssessmentScores(
	ctx context.Context,
	assessmentID uuid.UUID,
) ([]domain.PerformanceAssessmentScore, error) {
	if assessmentID == uuid.Nil {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	items, err := s.repository.ListAssessmentScores(ctx, assessmentID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PerformanceService) ListWorkAssignments(
	ctx context.Context,
	params domain.ListPerformanceWorkAssignmentsParams,
) (*domain.PerformanceWorkAssignmentPage, error) {
	normalized, err := normalizeListAssignmentsParams(params)
	if err != nil {
		return nil, err
	}

	page, err := s.repository.ListWorkAssignments(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("list work assignments: %w", err)
	}
	return page, nil
}

func (s *PerformanceService) GetWorkAssignmentByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.PerformanceWorkAssignment, error) {
	if id == uuid.Nil {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	item, err := s.repository.GetWorkAssignmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *PerformanceService) DecideWorkAssignment(
	ctx context.Context,
	id uuid.UUID,
	params domain.DecidePerformanceWorkAssignmentParams,
) (*domain.PerformanceWorkAssignment, error) {
	if id == uuid.Nil {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	decision := strings.TrimSpace(strings.ToLower(params.Decision))
	if decision != "approve" && decision != "request_revision" {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	normalized := domain.DecidePerformanceWorkAssignmentParams{
		Decision: decision,
		Feedback: performanceTrimStringPtr(params.Feedback),
	}

	item, err := s.repository.DecideWorkAssignment(ctx, id, normalized)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *PerformanceService) ListUpcoming(
	ctx context.Context,
	windowDays int,
) ([]domain.PerformanceUpcomingItem, error) {
	if windowDays <= 0 {
		windowDays = 14
	}
	if windowDays > 180 {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	items, err := s.repository.ListUpcoming(ctx, windowDays)
	if err != nil {
		return nil, fmt.Errorf("list upcoming: %w", err)
	}
	return items, nil
}

func (s *PerformanceService) SendUpcomingInvitations(
	ctx context.Context,
	employeeIDs []uuid.UUID,
	_ *string,
) (int, error) {
	if len(employeeIDs) == 0 {
		return 0, domain.ErrPerformanceInvalidRequest
	}
	for _, employeeID := range employeeIDs {
		if employeeID == uuid.Nil {
			return 0, domain.ErrPerformanceInvalidRequest
		}
	}

	// Placeholder implementation: accept the command and return count.
	// Hook into email/task queue in a later iteration.
	return len(employeeIDs), nil
}

func (s *PerformanceService) GetStats(ctx context.Context) (*domain.PerformanceStats, error) {
	stats, err := s.repository.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get performance stats: %w", err)
	}
	return stats, nil
}

func normalizeListAssessmentsParams(
	params domain.ListPerformanceAssessmentsParams,
) (domain.ListPerformanceAssessmentsParams, error) {
	if params.Limit < 5 || params.Limit > 100 {
		return domain.ListPerformanceAssessmentsParams{}, domain.ErrPerformanceInvalidRequest
	}
	if params.Offset < 0 {
		return domain.ListPerformanceAssessmentsParams{}, domain.ErrPerformanceInvalidRequest
	}

	if params.Status != nil {
		status := strings.TrimSpace(strings.ToLower(*params.Status))
		if status != domain.PerformanceAssessmentStatusDraft &&
			status != domain.PerformanceAssessmentStatusCompleted {
			return domain.ListPerformanceAssessmentsParams{}, domain.ErrPerformanceInvalidRequest
		}
		params.Status = &status
	}

	if params.FromDate != nil {
		normalized := performanceDateOnlyUTC(*params.FromDate)
		params.FromDate = &normalized
	}
	if params.ToDate != nil {
		normalized := performanceDateOnlyUTC(*params.ToDate)
		params.ToDate = &normalized
	}
	if params.FromDate != nil && params.ToDate != nil && params.FromDate.After(*params.ToDate) {
		return domain.ListPerformanceAssessmentsParams{}, domain.ErrPerformanceInvalidRequest
	}

	return params, nil
}

func normalizeListAssignmentsParams(
	params domain.ListPerformanceWorkAssignmentsParams,
) (domain.ListPerformanceWorkAssignmentsParams, error) {
	if params.Limit < 5 || params.Limit > 100 {
		return domain.ListPerformanceWorkAssignmentsParams{}, domain.ErrPerformanceInvalidRequest
	}
	if params.Offset < 0 {
		return domain.ListPerformanceWorkAssignmentsParams{}, domain.ErrPerformanceInvalidRequest
	}

	if params.Status != nil {
		status := strings.TrimSpace(strings.ToLower(*params.Status))
		switch status {
		case domain.PerformanceWorkAssignmentStatusOpen,
			domain.PerformanceWorkAssignmentStatusSubmitted,
			domain.PerformanceWorkAssignmentStatusApproved,
			domain.PerformanceWorkAssignmentStatusRevisionNeeded:
		default:
			return domain.ListPerformanceWorkAssignmentsParams{}, domain.ErrPerformanceInvalidRequest
		}
		params.Status = &status
	}

	if params.DueAfter != nil {
		normalized := performanceDateOnlyUTC(*params.DueAfter)
		params.DueAfter = &normalized
	}
	if params.DueBefore != nil {
		normalized := performanceDateOnlyUTC(*params.DueBefore)
		params.DueBefore = &normalized
	}
	if params.DueAfter != nil && params.DueBefore != nil && params.DueAfter.After(*params.DueBefore) {
		return domain.ListPerformanceWorkAssignmentsParams{}, domain.ErrPerformanceInvalidRequest
	}

	return params, nil
}

func performanceTrimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func performanceDateOnlyUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
