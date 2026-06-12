package service

import (
	"context"
	"encoding/json"
	"strings"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BugReportService struct {
	repository    domain.BugReportRepository
	cardPublisher domain.BugReportCardPublisher
	logger        domain.Logger
}

func NewBugReportService(
	repository domain.BugReportRepository,
	cardPublisher domain.BugReportCardPublisher,
	logger domain.Logger,
) domain.BugReportService {
	return &BugReportService{
		repository:    repository,
		cardPublisher: cardPublisher,
		logger:        logger,
	}
}

func (s *BugReportService) CreateBugReport(
	ctx context.Context,
	params domain.CreateBugReportParams,
) (*domain.BugReport, error) {
	normalized, err := normalizeCreateBugReportParams(params)
	if err != nil {
		return nil, err
	}

	report, err := s.repository.CreateBugReport(ctx, normalized)
	if err != nil {
		return nil, err
	}

	if s.cardPublisher == nil {
		return report, nil
	}

	card, err := s.cardPublisher.CreateBugReportCard(ctx, *report)
	if err != nil {
		s.logTrelloWarning(ctx, "create Trello card", err, report.ID)
		return report, nil
	}
	if card == nil || strings.TrimSpace(card.ID) == "" {
		return report, nil
	}

	updated, err := s.repository.UpdateBugReportTrelloCard(ctx, report.ID, *card)
	if err != nil {
		s.logTrelloWarning(ctx, "save Trello card details", err, report.ID)
		return report, nil
	}

	return updated, nil
}

func (s *BugReportService) logTrelloWarning(
	ctx context.Context,
	operation string,
	err error,
	bugReportID uuid.UUID,
) {
	if s.logger == nil {
		return
	}
	s.logger.LogWarn(
		ctx,
		"BugReportService.CreateBugReport",
		operation,
		zap.Error(err),
		zap.String("bug_report_id", bugReportID.String()),
	)
}

func normalizeCreateBugReportParams(
	params domain.CreateBugReportParams,
) (domain.CreateBugReportParams, error) {
	params.Subject = strings.TrimSpace(params.Subject)
	params.Category = strings.ToLower(strings.TrimSpace(params.Category))
	params.Severity = strings.ToLower(strings.TrimSpace(params.Severity))
	params.Description = strings.TrimSpace(params.Description)
	params.Steps = trimBugReportStringPtr(params.Steps)

	if params.UserID == uuid.Nil || params.Subject == "" || params.Description == "" {
		return domain.CreateBugReportParams{}, domain.ErrBugReportInvalidRequest
	}
	if !isValidBugReportCategory(params.Category) || !isValidBugReportSeverity(params.Severity) {
		return domain.CreateBugReportParams{}, domain.ErrBugReportInvalidRequest
	}

	if len(params.DebugInfo) == 0 {
		params.DebugInfo = json.RawMessage(`{}`)
	} else if !json.Valid(params.DebugInfo) {
		return domain.CreateBugReportParams{}, domain.ErrBugReportInvalidRequest
	}

	return params, nil
}

func isValidBugReportCategory(category string) bool {
	switch category {
	case domain.BugReportCategoryBug,
		domain.BugReportCategoryFeature,
		domain.BugReportCategoryImprovement,
		domain.BugReportCategoryOther:
		return true
	default:
		return false
	}
}

func isValidBugReportSeverity(severity string) bool {
	switch severity {
	case domain.BugReportSeverityLow,
		domain.BugReportSeverityMedium,
		domain.BugReportSeverityHigh,
		domain.BugReportSeverityCritical:
		return true
	default:
		return false
	}
}

func trimBugReportStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
