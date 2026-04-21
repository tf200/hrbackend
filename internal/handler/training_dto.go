package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type listTrainingCatalogItemsRequest struct {
	httpapi.PageRequest
	Search   *string `form:"search"`
	IsActive *bool   `form:"is_active"`
}

type createTrainingCatalogItemRequest struct {
	Title                    string  `json:"title" binding:"required"`
	Description              *string `json:"description"`
	Category                 *string `json:"category"`
	EstimatedDurationMinutes *int32  `json:"estimated_duration_minutes" binding:"omitempty,min=1"`
}

type trainingCatalogItemResponse struct {
	ID                       uuid.UUID  `json:"id"`
	Title                    string     `json:"title"`
	Description              *string    `json:"description"`
	Category                 *string    `json:"category"`
	EstimatedDurationMinutes *int32     `json:"estimated_duration_minutes"`
	IsActive                 bool       `json:"is_active"`
	CreatedByEmployeeID      *uuid.UUID `json:"created_by_employee_id"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

func toCreateTrainingCatalogItemParams(
	req createTrainingCatalogItemRequest,
	employeeID uuid.UUID,
) domain.CreateTrainingCatalogItemParams {
	var createdByEmployeeID *uuid.UUID
	if employeeID != uuid.Nil {
		createdByEmployeeID = &employeeID
	}

	return domain.CreateTrainingCatalogItemParams{
		Title:                    req.Title,
		Description:              req.Description,
		Category:                 req.Category,
		EstimatedDurationMinutes: req.EstimatedDurationMinutes,
		CreatedByEmployeeID:      createdByEmployeeID,
	}
}

func toListTrainingCatalogItemsParams(
	req listTrainingCatalogItemsRequest,
) domain.ListTrainingCatalogItemsParams {
	return domain.ListTrainingCatalogItemsParams{
		Limit:    req.PageSize,
		Offset:   (req.Page - 1) * req.PageSize,
		Search:   req.Search,
		IsActive: req.IsActive,
	}
}

func toTrainingCatalogItemResponse(item *domain.TrainingCatalogItem) trainingCatalogItemResponse {
	return trainingCatalogItemResponse{
		ID:                       item.ID,
		Title:                    item.Title,
		Description:              item.Description,
		Category:                 item.Category,
		EstimatedDurationMinutes: item.EstimatedDurationMinutes,
		IsActive:                 item.IsActive,
		CreatedByEmployeeID:      item.CreatedByEmployeeID,
		CreatedAt:                item.CreatedAt,
		UpdatedAt:                item.UpdatedAt,
	}
}

func toTrainingCatalogItemResponses(items []domain.TrainingCatalogItem) []trainingCatalogItemResponse {
	results := make([]trainingCatalogItemResponse, len(items))
	for i := range items {
		results[i] = toTrainingCatalogItemResponse(&items[i])
	}
	return results
}
