package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrBugReportInvalidRequest = errors.New("invalid bug report")

const (
	BugReportCategoryBug         = "bug"
	BugReportCategoryFeature     = "feature"
	BugReportCategoryImprovement = "improvement"
	BugReportCategoryOther       = "other"

	BugReportSeverityLow      = "low"
	BugReportSeverityMedium   = "medium"
	BugReportSeverityHigh     = "high"
	BugReportSeverityCritical = "critical"

	BugReportStatusOpen       = "open"
	BugReportStatusInProgress = "in_progress"
	BugReportStatusResolved   = "resolved"
	BugReportStatusClosed     = "closed"
)

type BugReport struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Subject       string
	Category      string
	Severity      string
	Description   string
	Steps         *string
	DebugInfo     json.RawMessage
	TrelloCardID  *string
	TrelloCardURL *string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateBugReportParams struct {
	UserID      uuid.UUID
	Subject     string
	Category    string
	Severity    string
	Description string
	Steps       *string
	DebugInfo   json.RawMessage
}

type BugReportRepository interface {
	CreateBugReport(ctx context.Context, params CreateBugReportParams) (*BugReport, error)
	UpdateBugReportTrelloCard(
		ctx context.Context,
		bugReportID uuid.UUID,
		card BugReportCard,
	) (*BugReport, error)
}

type BugReportCard struct {
	ID  string
	URL string
}

type BugReportCardPublisher interface {
	CreateBugReportCard(ctx context.Context, report BugReport) (*BugReportCard, error)
}

type BugReportService interface {
	CreateBugReport(ctx context.Context, params CreateBugReportParams) (*BugReport, error)
}
