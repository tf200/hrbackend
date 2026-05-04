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

	activeQuestions := make(map[string]domain.PerformanceQuestion)
	for _, domainItem := range catalog {
		for _, question := range domainItem.Questions {
			activeQuestions[question.Code] = question
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

	var assessment *domain.PerformanceAssessment
	err = s.repository.WithTx(ctx, func(tx domain.PerformanceTxRepository) error {
		employeeName, err := tx.GetActiveEmployeeName(ctx, normalized.EmployeeID)
		if err != nil {
			return err
		}

		created, err := tx.CreateAssessment(ctx, domain.CreatePerformanceAssessmentRecordParams{
			EmployeeID:         normalized.EmployeeID,
			ReviewerEmployeeID: normalized.ReviewerEmployeeID,
			AssessmentDate:     normalized.AssessmentDate,
			TotalScore:         averagePerformanceScore(normalized.Scores),
			Notes:              normalized.Notes,
		}, *employeeName)
		if err != nil {
			return err
		}

		for _, score := range normalized.Scores {
			if err := tx.CreateAssessmentScore(ctx, created.ID, score); err != nil {
				return err
			}
		}

		assessment = created
		return nil
	})
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

	err := s.repository.WithTx(ctx, func(tx domain.PerformanceTxRepository) error {
		currentStatus, err := tx.GetWorkAssignmentStatusForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if currentStatus != domain.PerformanceWorkAssignmentStatusSubmitted {
			return domain.ErrPerformanceStateInvalid
		}

		nextStatus := domain.PerformanceWorkAssignmentStatusApproved
		if normalized.Decision == "request_revision" {
			nextStatus = domain.PerformanceWorkAssignmentStatusRevisionNeeded
		}

		return tx.UpdateWorkAssignmentDecision(ctx, id, nextStatus, normalized.Feedback)
	})
	if err != nil {
		return nil, err
	}

	return s.repository.GetWorkAssignmentByID(ctx, id)
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

func (s *PerformanceService) GetMine(
	ctx context.Context,
	params domain.PerformanceMineParams,
) (*domain.PerformanceMine, error) {
	if params.EmployeeID == uuid.Nil {
		return nil, domain.ErrPerformanceInvalidRequest
	}
	if params.Limit <= 0 {
		params.Limit = 12
	}
	if params.Limit > 100 {
		return nil, domain.ErrPerformanceInvalidRequest
	}

	reviewContext, err := s.repository.GetMineReviewContext(ctx, params.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("get performance mine review context: %w", err)
	}

	completedStatus := domain.PerformanceAssessmentStatusCompleted
	assessmentPage, err := s.repository.ListAssessments(ctx, domain.ListPerformanceAssessmentsParams{
		Limit:      params.Limit,
		Offset:     0,
		EmployeeID: &params.EmployeeID,
		Status:     &completedStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("list performance mine assessments: %w", err)
	}

	assignments := make([]domain.PerformanceWorkAssignment, 0)
	if params.IncludeAssignments {
		assignmentPage, err := s.repository.ListWorkAssignments(ctx, domain.ListPerformanceWorkAssignmentsParams{
			Limit:      100,
			Offset:     0,
			EmployeeID: &params.EmployeeID,
		})
		if err != nil {
			return nil, fmt.Errorf("list performance mine work assignments: %w", err)
		}
		assignments = assignmentPage.Items
	}

	scoresByAssessment := make(map[uuid.UUID][]domain.PerformanceAssessmentScore)
	if params.IncludeScores {
		for _, assessment := range assessmentPage.Items {
			scores, err := s.repository.ListAssessmentScores(ctx, assessment.ID)
			if err != nil {
				return nil, fmt.Errorf("list performance mine scores: %w", err)
			}
			scoresByAssessment[assessment.ID] = scores
		}
	}

	mine := &domain.PerformanceMine{
		Employee: domain.PerformanceMineEmployee{
			ID:   reviewContext.EmployeeID,
			Name: reviewContext.EmployeeName,
		},
		ReviewIntervalDays: domain.PerformanceReviewIntervalDays,
		NextReview: domain.PerformanceMineNextReview{
			LastAssessmentDate: reviewContext.LastAssessmentDate,
			NextAssessmentDate: reviewContext.NextAssessmentDate,
			DaysUntilDue:       reviewContext.DaysUntilDue,
			IsOverdue:          reviewContext.IsOverdue,
			IsDueSoon:          reviewContext.IsDueSoon,
			IsFirstReview:      reviewContext.IsFirstReview,
		},
		Assessments:     make([]domain.PerformanceMineAssessment, 0, len(assessmentPage.Items)),
		WorkAssignments: assignments,
	}

	var total float64
	var scoreCount int
	for i, assessment := range assessmentPage.Items {
		cycleNumber := len(assessmentPage.Items) - i
		var scoreDelta *float64
		if assessment.TotalScore != nil && i+1 < len(assessmentPage.Items) && assessmentPage.Items[i+1].TotalScore != nil {
			delta := *assessment.TotalScore - *assessmentPage.Items[i+1].TotalScore
			scoreDelta = &delta
		}

		mine.Assessments = append(mine.Assessments, domain.PerformanceMineAssessment{
			PerformanceAssessment: assessment,
			Title:                 assessment.AssessmentDate.Format("Assessment — January 2006"),
			CycleNumber:           cycleNumber,
			ScoreDelta:            scoreDelta,
			Scores:                scoresByAssessment[assessment.ID],
		})

		if assessment.TotalScore != nil {
			total += *assessment.TotalScore
			scoreCount++
		}
	}

	mine.Summary.AssessmentCount = len(assessmentPage.Items)
	if len(assessmentPage.Items) > 0 {
		mine.Summary.LatestScore = assessmentPage.Items[0].TotalScore
		mine.Summary.FirstScore = assessmentPage.Items[len(assessmentPage.Items)-1].TotalScore
		if mine.Summary.LatestScore != nil && mine.Summary.FirstScore != nil && len(assessmentPage.Items) > 1 {
			growth := *mine.Summary.LatestScore - *mine.Summary.FirstScore
			mine.Summary.ScoreGrowth = &growth
		}
	}
	if scoreCount > 0 {
		average := total / float64(scoreCount)
		mine.Summary.AverageScore = &average
	}

	for _, assignment := range assignments {
		switch assignment.Status {
		case domain.PerformanceWorkAssignmentStatusOpen:
			mine.Summary.OpenAssignmentCount++
		case domain.PerformanceWorkAssignmentStatusSubmitted:
			mine.Summary.SubmittedAssignmentCount++
		case domain.PerformanceWorkAssignmentStatusApproved:
			mine.Summary.ApprovedAssignmentCount++
		case domain.PerformanceWorkAssignmentStatusRevisionNeeded:
			mine.Summary.RevisionNeededAssignmentCount++
		}
	}

	if params.IncludeScores {
		for _, assessment := range assessmentPage.Items {
			if assessment.Status != domain.PerformanceAssessmentStatusCompleted {
				continue
			}
			for i := range scoresByAssessment[assessment.ID] {
				score := scoresByAssessment[assessment.ID][i]
				if mine.Highlighted.StrongestScore == nil || score.Rating > mine.Highlighted.StrongestScore.Rating {
					mine.Highlighted.StrongestScore = &score
				}
				if mine.Highlighted.FocusScore == nil || score.Rating < mine.Highlighted.FocusScore.Rating {
					mine.Highlighted.FocusScore = &score
				}
			}
			break
		}
	}

	return mine, nil
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

func averagePerformanceScore(scores []domain.CreatePerformanceAssessmentScoreParams) *float64 {
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
