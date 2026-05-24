package handler

import (
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type qualificationTypeResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	AppContext *string   `json:"app_context"`
	CreatedAt time.Time `json:"created_at"`
}

func toQualificationTypeResponse(qt *domain.QualificationType) qualificationTypeResponse {
	return qualificationTypeResponse{
		ID:        qt.ID,
		Code:      qt.Code,
		Name:      qt.Name,
		AppContext: qt.AppContext,
		CreatedAt: qt.CreatedAt,
	}
}

type createQualificationTypeRequest struct {
	Name string `json:"name" binding:"required"`
}
