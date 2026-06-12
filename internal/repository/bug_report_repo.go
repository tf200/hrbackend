package repository

import (
	"context"
	"encoding/json"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
)

type BugReportRepository struct {
	queries db.Querier
}

func NewBugReportRepository(queries db.Querier) domain.BugReportRepository {
	return &BugReportRepository{queries: queries}
}

func (r *BugReportRepository) CreateBugReport(
	ctx context.Context,
	params domain.CreateBugReportParams,
) (*domain.BugReport, error) {
	row, err := r.queries.CreateBugReport(ctx, db.CreateBugReportParams{
		UserID:      params.UserID,
		Subject:     params.Subject,
		Category:    db.BugReportCategoryEnum(params.Category),
		Severity:    db.BugReportSeverityEnum(params.Severity),
		Description: params.Description,
		Steps:       params.Steps,
		DebugInfo:   []byte(params.DebugInfo),
	})
	if err != nil {
		return nil, err
	}

	model := toDomainBugReport(row)
	return &model, nil
}

func (r *BugReportRepository) UpdateBugReportTrelloCard(
	ctx context.Context,
	bugReportID uuid.UUID,
	card domain.BugReportCard,
) (*domain.BugReport, error) {
	row, err := r.queries.UpdateBugReportTrelloCard(ctx, db.UpdateBugReportTrelloCardParams{
		ID:            bugReportID,
		TrelloCardID:  &card.ID,
		TrelloCardUrl: &card.URL,
	})
	if err != nil {
		return nil, err
	}

	model := toDomainBugReport(row)
	return &model, nil
}

func toDomainBugReport(row db.BugReport) domain.BugReport {
	debugInfo := json.RawMessage(row.DebugInfo)
	return domain.BugReport{
		ID:            row.ID,
		UserID:        row.UserID,
		Subject:       row.Subject,
		Category:      string(row.Category),
		Severity:      string(row.Severity),
		Description:   row.Description,
		Steps:         row.Steps,
		DebugInfo:     debugInfo,
		TrelloCardID:  row.TrelloCardID,
		TrelloCardURL: row.TrelloCardUrl,
		Status:        string(row.Status),
		CreatedAt:     conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:     conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}
