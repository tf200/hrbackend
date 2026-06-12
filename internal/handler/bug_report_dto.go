package handler

import (
	"encoding/json"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type createBugReportRequest struct {
	Subject     string          `json:"subject"     binding:"required"`
	Category    string          `json:"category"    binding:"required,oneof=bug feature improvement other"`
	Severity    string          `json:"severity"    binding:"required,oneof=low medium high critical"`
	Description string          `json:"description" binding:"required"`
	Steps       *string         `json:"steps"`
	DebugInfo   json.RawMessage `json:"debug_info"`
}

type bugReportResponse struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	Subject       string          `json:"subject"`
	Category      string          `json:"category"`
	Severity      string          `json:"severity"`
	Description   string          `json:"description"`
	Steps         *string         `json:"steps,omitempty"`
	DebugInfo     json.RawMessage `json:"debug_info"`
	TrelloCardID  *string         `json:"trello_card_id,omitempty"`
	TrelloCardURL *string         `json:"trello_card_url,omitempty"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
}

func toCreateBugReportParams(
	userID uuid.UUID,
	req createBugReportRequest,
) domain.CreateBugReportParams {
	return domain.CreateBugReportParams{
		UserID:      userID,
		Subject:     req.Subject,
		Category:    req.Category,
		Severity:    req.Severity,
		Description: req.Description,
		Steps:       req.Steps,
		DebugInfo:   req.DebugInfo,
	}
}

func toBugReportResponse(item *domain.BugReport) bugReportResponse {
	return bugReportResponse{
		ID:            item.ID,
		UserID:        item.UserID,
		Subject:       item.Subject,
		Category:      item.Category,
		Severity:      item.Severity,
		Description:   item.Description,
		Steps:         item.Steps,
		DebugInfo:     item.DebugInfo,
		TrelloCardID:  item.TrelloCardID,
		TrelloCardURL: item.TrelloCardURL,
		Status:        item.Status,
		CreatedAt:     item.CreatedAt,
	}
}
