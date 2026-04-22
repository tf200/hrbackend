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

type listTrainingAssignmentsRequest struct {
	httpapi.PageRequest
	EmployeeSearch *string    `form:"employee_search"`
	DepartmentID   *uuid.UUID `form:"department_id,parser=encoding.TextUnmarshaler"`
	TrainingID     *uuid.UUID `form:"training_id,parser=encoding.TextUnmarshaler"`
	Status         *string    `form:"status" binding:"omitempty,oneof=assigned in_progress completed cancelled"`
}

type createTrainingCatalogItemRequest struct {
	Title                    string  `json:"title" binding:"required"`
	Description              *string `json:"description"`
	Category                 *string `json:"category"`
	EstimatedDurationMinutes *int32  `json:"estimated_duration_minutes" binding:"omitempty,min=1"`
}

type assignTrainingRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	TrainingID uuid.UUID `json:"training_id" binding:"required"`
	DueAt      time.Time `json:"due_at" binding:"required"`
}

type cancelTrainingAssignmentRequest struct {
	Reason *string `json:"reason"`
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

type trainingAssignmentResponse struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	TrainingID           uuid.UUID  `json:"training_id"`
	AssignedByEmployeeID *uuid.UUID `json:"assigned_by_employee_id"`
	Status               string     `json:"status"`
	AssignedAt           time.Time  `json:"assigned_at"`
	DueAt                time.Time  `json:"due_at"`
	StartedAt            *time.Time `json:"started_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	CancelledAt          *time.Time `json:"cancelled_at"`
	CancellationReason   *string    `json:"cancellation_reason"`
	CompletionNotes      *string    `json:"completion_notes"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type trainingAssignmentListItemResponse struct {
	AssignmentID         uuid.UUID  `json:"assignment_id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	EmployeeNumber       *string    `json:"employee_number"`
	EmploymentNumber     *string    `json:"employment_number"`
	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	DepartmentID         *uuid.UUID `json:"department_id"`
	DepartmentName       *string    `json:"department_name"`
	TrainingID           uuid.UUID  `json:"training_id"`
	TrainingTitle        string     `json:"training_title"`
	TrainingCategory     *string    `json:"training_category"`
	Status               string     `json:"status"`
	AssignedAt           time.Time  `json:"assigned_at"`
	DueAt                *time.Time `json:"due_at"`
	StartedAt            *time.Time `json:"started_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	AssignedByEmployeeID *uuid.UUID `json:"assigned_by_employee_id"`
	AssignedByName       *string    `json:"assigned_by_name"`
	IsOverdue            bool       `json:"is_overdue"`
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

func toAssignTrainingToEmployeeParams(
	req assignTrainingRequest,
	employeeID uuid.UUID,
) domain.AssignTrainingToEmployeeParams {
	var assignedByEmployeeID *uuid.UUID
	if employeeID != uuid.Nil {
		assignedByEmployeeID = &employeeID
	}

	return domain.AssignTrainingToEmployeeParams{
		EmployeeID:           req.EmployeeID,
		TrainingID:           req.TrainingID,
		AssignedByEmployeeID: assignedByEmployeeID,
		DueAt:                req.DueAt,
	}
}

func toCancelTrainingAssignmentParams(
	assignmentID uuid.UUID,
	req cancelTrainingAssignmentRequest,
) domain.CancelTrainingAssignmentParams {
	return domain.CancelTrainingAssignmentParams{
		AssignmentID:       assignmentID,
		CancellationReason: req.Reason,
	}
}

func toListTrainingCatalogItemsParams(
	req listTrainingCatalogItemsRequest,
) domain.ListTrainingCatalogItemsParams {
	params := req.Params()
	return domain.ListTrainingCatalogItemsParams{
		Limit:    params.Limit,
		Offset:   params.Offset,
		Search:   req.Search,
		IsActive: req.IsActive,
	}
}

func toListTrainingAssignmentsParams(
	req listTrainingAssignmentsRequest,
) domain.ListTrainingAssignmentsParams {
	params := req.Params()
	return domain.ListTrainingAssignmentsParams{
		Limit:          params.Limit,
		Offset:         params.Offset,
		EmployeeSearch: req.EmployeeSearch,
		DepartmentID:   req.DepartmentID,
		TrainingID:     req.TrainingID,
		Status:         req.Status,
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

func toTrainingAssignmentResponse(item *domain.EmployeeTrainingAssignment) trainingAssignmentResponse {
	return trainingAssignmentResponse{
		ID:                   item.ID,
		EmployeeID:           item.EmployeeID,
		TrainingID:           item.TrainingID,
		AssignedByEmployeeID: item.AssignedByEmployeeID,
		Status:               item.Status,
		AssignedAt:           item.AssignedAt,
		DueAt:                item.DueAt,
		StartedAt:            item.StartedAt,
		CompletedAt:          item.CompletedAt,
		CancelledAt:          item.CancelledAt,
		CancellationReason:   item.CancellationReason,
		CompletionNotes:      item.CompletionNotes,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}

func toTrainingAssignmentListItemResponse(
	item domain.TrainingAssignmentListItem,
) trainingAssignmentListItemResponse {
	return trainingAssignmentListItemResponse{
		AssignmentID:         item.AssignmentID,
		EmployeeID:           item.EmployeeID,
		EmployeeNumber:       item.EmployeeNumber,
		EmploymentNumber:     item.EmploymentNumber,
		FirstName:            item.FirstName,
		LastName:             item.LastName,
		DepartmentID:         item.DepartmentID,
		DepartmentName:       item.DepartmentName,
		TrainingID:           item.TrainingID,
		TrainingTitle:        item.TrainingTitle,
		TrainingCategory:     item.TrainingCategory,
		Status:               item.Status,
		AssignedAt:           item.AssignedAt,
		DueAt:                item.DueAt,
		StartedAt:            item.StartedAt,
		CompletedAt:          item.CompletedAt,
		AssignedByEmployeeID: item.AssignedByEmployeeID,
		AssignedByName:       item.AssignedByName,
		IsOverdue:            item.IsOverdue,
	}
}
