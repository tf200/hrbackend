package handler

import (
	"time"

	"github.com/goccy/go-json"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type completeMyHandbookStepRequest struct {
	Response json.RawMessage `json:"response"`
}

type createHandbookTemplateRequest struct {
	DepartmentID uuid.UUID `json:"department_id" binding:"required"`
	Title        string    `json:"title"         binding:"required"`
	Description  *string   `json:"description"`
}

type cloneHandbookTemplateRequest struct {
	SourceTemplateID uuid.UUID `json:"source_template_id" binding:"required"`
}

type updateHandbookTemplateRequest struct {
	Title       *json.RawMessage `json:"title"`
	Description *json.RawMessage `json:"description"`
}

type publishHandbookTemplateRequest struct {
	TemplateID uuid.UUID `json:"template_id" binding:"required"`
}

type createHandbookStepRequest struct {
	TemplateID uuid.UUID       `json:"template_id" binding:"required"`
	SortOrder  int32           `json:"sort_order"  binding:"required"`
	Kind       string          `json:"kind"        binding:"required,oneof=content ack link quiz rich_text"`
	Title      string          `json:"title"       binding:"required"`
	Body       *string         `json:"body"`
	Content    json.RawMessage `json:"content"`
	IsRequired *bool           `json:"is_required"`
}

type updateHandbookStepRequest struct {
	Title      *json.RawMessage `json:"title"`
	Body       *json.RawMessage `json:"body"`
	Content    *json.RawMessage `json:"content"`
	IsRequired *json.RawMessage `json:"is_required"`
}

type reorderHandbookStepsRequest struct {
	OrderedStepIDs []uuid.UUID `json:"ordered_step_ids" binding:"required,min=1"`
}

type assignHandbookTemplateRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	TemplateID uuid.UUID `json:"template_id" binding:"required"`
}

type waiveEmployeeHandbookRequest struct {
	Reason *string `json:"reason"`
}

type listEmployeeHandbookAssignmentsRequest struct {
	httpapi.PageRequest
	DepartmentID *uuid.UUID `form:"department_id"`
	Search       *string    `form:"search"`
	Status       *string    `form:"status"`
}

type listEligibleEmployeesRequest struct {
	httpapi.PageRequest
	DepartmentID *uuid.UUID `form:"department_id"`
	Search       *string    `form:"search"`
}

type handbookTemplateResponse struct {
	ID           uuid.UUID  `json:"id"`
	DepartmentID uuid.UUID  `json:"department_id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Version      int32      `json:"version"`
	Status       string     `json:"status"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type handbookStepResponse struct {
	ID         uuid.UUID  `json:"id"`
	TemplateID uuid.UUID  `json:"template_id"`
	SortOrder  int32      `json:"sort_order"`
	Kind       string     `json:"kind"`
	Title      string     `json:"title"`
	Body       *string    `json:"body"`
	Content    any        `json:"content"`
	IsRequired bool       `json:"is_required"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type myHandbookStepResponse struct {
	StepID      uuid.UUID  `json:"step_id"`
	SortOrder   int32      `json:"sort_order"`
	Kind        string     `json:"kind"`
	Title       string     `json:"title"`
	Body        *string    `json:"body"`
	Content     any        `json:"content"`
	IsRequired  bool       `json:"is_required"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Response    any        `json:"response"`
}

type myActiveHandbookResponse struct {
	HandbookID      uuid.UUID                `json:"handbook_id"`
	Status          string                   `json:"status"`
	AssignedAt      time.Time                `json:"assigned_at"`
	StartedAt       *time.Time               `json:"started_at"`
	CompletedAt     *time.Time               `json:"completed_at"`
	DueAt           *time.Time               `json:"due_at"`
	TemplateID      uuid.UUID                `json:"template_id"`
	TemplateTitle   string                   `json:"template_title"`
	TemplateDesc    *string                  `json:"template_description"`
	TemplateVersion int32                    `json:"template_version"`
	DepartmentID    uuid.UUID                `json:"department_id"`
	DepartmentName  string                   `json:"department_name"`
	Steps           []myHandbookStepResponse `json:"steps"`
}

type startedHandbookResponse struct {
	HandbookID uuid.UUID  `json:"handbook_id"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at"`
}

type completedHandbookStepResponse struct {
	HandbookID     uuid.UUID `json:"handbook_id"`
	StepID         uuid.UUID `json:"step_id"`
	StepStatus     string    `json:"step_status"`
	CompletedAt    time.Time `json:"completed_at"`
	HandbookStatus string    `json:"handbook_status"`
}

type employeeHandbookAssignmentResponse struct {
	EmployeeHandbookID uuid.UUID  `json:"employee_handbook_id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	TemplateID         uuid.UUID  `json:"template_id"`
	TemplateVersion    int32      `json:"template_version"`
	AssignedAt         time.Time  `json:"assigned_at"`
	StartedAt          *time.Time `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	DueAt              *time.Time `json:"due_at"`
	Status             string     `json:"status"`
}

type waivedEmployeeHandbookResponse struct {
	EmployeeHandbookID uuid.UUID  `json:"employee_handbook_id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	Status             string     `json:"status"`
	CompletedAt        *time.Time `json:"completed_at"`
}

type handbookAssignmentHistoryEntryResponse struct {
	ID                 uuid.UUID  `json:"id"`
	EmployeeHandbookID *uuid.UUID `json:"employee_handbook_id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	TemplateID         uuid.UUID  `json:"template_id"`
	TemplateVersion    int32      `json:"template_version"`
	Event              string     `json:"event"`
	ActorEmployeeID    *uuid.UUID `json:"actor_employee_id"`
	Metadata           any        `json:"metadata"`
	CreatedAt          time.Time  `json:"created_at"`
}

type employeeHandbookAssignmentSummaryResponse struct {
	EmployeeID             uuid.UUID  `json:"employee_id"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	DepartmentID           *uuid.UUID `json:"department_id"`
	DepartmentName         *string    `json:"department_name"`
	EmployeeHandbookID     *uuid.UUID `json:"employee_handbook_id"`
	TemplateID             *uuid.UUID `json:"template_id"`
	TemplateTitle          *string    `json:"template_title"`
	TemplateVersion        *int32     `json:"template_version"`
	HandbookStatus         string     `json:"handbook_status"`
	AssignedAt             *time.Time `json:"assigned_at"`
	StartedAt              *time.Time `json:"started_at"`
	CompletedAt            *time.Time `json:"completed_at"`
	DueAt                  *time.Time `json:"due_at"`
	RequiredStepsTotal     int32      `json:"required_steps_total"`
	RequiredStepsCompleted int32      `json:"required_steps_completed"`
}

type eligibleEmployeeResponse struct {
	EmployeeID     uuid.UUID  `json:"employee_id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	DepartmentID   *uuid.UUID `json:"department_id"`
	DepartmentName *string    `json:"department_name"`
}

type employeeHandbookDetailsResponse struct {
	EmployeeHandbookID uuid.UUID                `json:"employee_handbook_id"`
	EmployeeID         uuid.UUID                `json:"employee_id"`
	FirstName          string                   `json:"first_name"`
	LastName           string                   `json:"last_name"`
	Status             string                   `json:"status"`
	AssignedAt         time.Time                `json:"assigned_at"`
	StartedAt          *time.Time               `json:"started_at"`
	CompletedAt        *time.Time               `json:"completed_at"`
	DueAt              *time.Time               `json:"due_at"`
	TemplateID         uuid.UUID                `json:"template_id"`
	TemplateTitle      string                   `json:"template_title"`
	TemplateDesc       *string                  `json:"template_description"`
	TemplateVersion    int32                    `json:"template_version"`
	DepartmentID       uuid.UUID                `json:"department_id"`
	DepartmentName     string                   `json:"department_name"`
	Steps              []myHandbookStepResponse `json:"steps"`
}

func toCreateTemplateForDepartmentParams(
	req createHandbookTemplateRequest,
) domain.CreateTemplateForDepartmentParams {
	return domain.CreateTemplateForDepartmentParams{
		DepartmentID: req.DepartmentID,
		Title:        req.Title,
		Description:  req.Description,
	}
}

func toCloneTemplateToDraftParams(
	req cloneHandbookTemplateRequest,
) domain.CloneTemplateToDraftParams {
	return domain.CloneTemplateToDraftParams{SourceTemplateID: req.SourceTemplateID}
}

func toPublishTemplateParams(req publishHandbookTemplateRequest) domain.PublishTemplateParams {
	return domain.PublishTemplateParams{TemplateID: req.TemplateID}
}

func toCreateHandbookStepParams(req createHandbookStepRequest) domain.CreateStepParams {
	return domain.CreateStepParams{
		TemplateID: req.TemplateID,
		SortOrder:  req.SortOrder,
		Kind:       req.Kind,
		Title:      req.Title,
		Body:       req.Body,
		Content:    req.Content,
		IsRequired: req.IsRequired,
	}
}

func toAssignHandbookTemplateParams(
	req assignHandbookTemplateRequest,
) domain.AssignTemplateToEmployeeParams {
	return domain.AssignTemplateToEmployeeParams{
		EmployeeID: req.EmployeeID,
		TemplateID: req.TemplateID,
	}
}

func toListEmployeeHandbookAssignmentsParams(
	req listEmployeeHandbookAssignmentsRequest,
) domain.ListEmployeeHandbookAssignmentsParams {
	params := req.Params()
	return domain.ListEmployeeHandbookAssignmentsParams{
		Limit:        params.Limit,
		Offset:       params.Offset,
		DepartmentID: req.DepartmentID,
		Search:       req.Search,
		Status:       req.Status,
	}
}

func toListEligibleEmployeesParams(
	req listEligibleEmployeesRequest,
) domain.ListEligibleEmployeesParams {
	params := req.Params()
	return domain.ListEligibleEmployeesParams{
		Limit:        params.Limit,
		Offset:       params.Offset,
		DepartmentID: req.DepartmentID,
		Search:       req.Search,
	}
}

func toHandbookTemplateResponse(item *domain.HandbookTemplate) handbookTemplateResponse {
	return handbookTemplateResponse{
		ID:           item.ID,
		DepartmentID: item.DepartmentID,
		Title:        item.Title,
		Description:  item.Description,
		Version:      item.Version,
		Status:       item.Status,
		PublishedAt:  item.PublishedAt,
		ArchivedAt:   item.ArchivedAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func toHandbookStepResponse(item *domain.HandbookStep) handbookStepResponse {
	var updatedAt *time.Time
	if !item.UpdatedAt.IsZero() {
		updatedAt = &item.UpdatedAt
	}
	return handbookStepResponse{
		ID:         item.ID,
		TemplateID: item.TemplateID,
		SortOrder:  item.SortOrder,
		Kind:       item.Kind,
		Title:      item.Title,
		Body:       item.Body,
		Content:    item.Content,
		IsRequired: item.IsRequired,
		UpdatedAt:  updatedAt,
	}
}

func toMyHandbookStepResponse(item domain.MyHandbookStep) myHandbookStepResponse {
	return myHandbookStepResponse{
		StepID:      item.StepID,
		SortOrder:   item.SortOrder,
		Kind:        item.Kind,
		Title:       item.Title,
		Body:        item.Body,
		Content:     item.Content,
		IsRequired:  item.IsRequired,
		Status:      item.Status,
		StartedAt:   item.StartedAt,
		CompletedAt: item.CompletedAt,
		Response:    item.Response,
	}
}

func toMyActiveHandbookResponse(item *domain.MyActiveHandbook) myActiveHandbookResponse {
	steps := make([]myHandbookStepResponse, 0, len(item.Steps))
	for _, step := range item.Steps {
		steps = append(steps, toMyHandbookStepResponse(step))
	}
	return myActiveHandbookResponse{
		HandbookID:      item.HandbookID,
		Status:          item.Status,
		AssignedAt:      item.AssignedAt,
		StartedAt:       item.StartedAt,
		CompletedAt:     item.CompletedAt,
		DueAt:           item.DueAt,
		TemplateID:      item.TemplateID,
		TemplateTitle:   item.TemplateTitle,
		TemplateDesc:    item.TemplateDesc,
		TemplateVersion: item.TemplateVersion,
		DepartmentID:    item.DepartmentID,
		DepartmentName:  item.DepartmentName,
		Steps:           steps,
	}
}

func toStartedHandbookResponse(item *domain.StartedHandbook) startedHandbookResponse {
	return startedHandbookResponse{
		HandbookID: item.HandbookID,
		Status:     item.Status,
		StartedAt:  item.StartedAt,
	}
}

func toCompletedHandbookStepResponse(
	item *domain.CompletedHandbookStep,
) completedHandbookStepResponse {
	return completedHandbookStepResponse{
		HandbookID:     item.HandbookID,
		StepID:         item.StepID,
		StepStatus:     item.StepStatus,
		CompletedAt:    item.CompletedAt,
		HandbookStatus: item.HandbookStatus,
	}
}

func toEmployeeHandbookAssignmentResponse(
	item *domain.EmployeeHandbookAssignment,
) employeeHandbookAssignmentResponse {
	return employeeHandbookAssignmentResponse{
		EmployeeHandbookID: item.EmployeeHandbookID,
		EmployeeID:         item.EmployeeID,
		TemplateID:         item.TemplateID,
		TemplateVersion:    item.TemplateVersion,
		AssignedAt:         item.AssignedAt,
		StartedAt:          item.StartedAt,
		CompletedAt:        item.CompletedAt,
		DueAt:              item.DueAt,
		Status:             item.Status,
	}
}

func toWaivedEmployeeHandbookResponse(
	item *domain.WaivedEmployeeHandbook,
) waivedEmployeeHandbookResponse {
	return waivedEmployeeHandbookResponse{
		EmployeeHandbookID: item.EmployeeHandbookID,
		EmployeeID:         item.EmployeeID,
		Status:             item.Status,
		CompletedAt:        item.CompletedAt,
	}
}

func toHandbookAssignmentHistoryEntryResponse(
	item domain.HandbookAssignmentHistoryEntry,
) handbookAssignmentHistoryEntryResponse {
	return handbookAssignmentHistoryEntryResponse{
		ID:                 item.ID,
		EmployeeHandbookID: item.EmployeeHandbookID,
		EmployeeID:         item.EmployeeID,
		TemplateID:         item.TemplateID,
		TemplateVersion:    item.TemplateVersion,
		Event:              item.Event,
		ActorEmployeeID:    item.ActorEmployeeID,
		Metadata:           item.Metadata,
		CreatedAt:          item.CreatedAt,
	}
}

func toEmployeeHandbookAssignmentSummaryResponse(
	item domain.EmployeeHandbookAssignmentSummary,
) employeeHandbookAssignmentSummaryResponse {
	return employeeHandbookAssignmentSummaryResponse{
		EmployeeID:             item.EmployeeID,
		FirstName:              item.FirstName,
		LastName:               item.LastName,
		DepartmentID:           item.DepartmentID,
		DepartmentName:         item.DepartmentName,
		EmployeeHandbookID:     item.EmployeeHandbookID,
		TemplateID:             item.TemplateID,
		TemplateTitle:          item.TemplateTitle,
		TemplateVersion:        item.TemplateVersion,
		HandbookStatus:         item.HandbookStatus,
		AssignedAt:             item.AssignedAt,
		StartedAt:              item.StartedAt,
		CompletedAt:            item.CompletedAt,
		DueAt:                  item.DueAt,
		RequiredStepsTotal:     item.RequiredStepsTotal,
		RequiredStepsCompleted: item.RequiredStepsCompleted,
	}
}

func toEligibleEmployeeResponse(item domain.EligibleEmployee) eligibleEmployeeResponse {
	return eligibleEmployeeResponse{
		EmployeeID:     item.EmployeeID,
		FirstName:      item.FirstName,
		LastName:       item.LastName,
		DepartmentID:   item.DepartmentID,
		DepartmentName: item.DepartmentName,
	}
}

func toEmployeeHandbookDetailsResponse(
	item *domain.EmployeeHandbookDetails,
) employeeHandbookDetailsResponse {
	steps := make([]myHandbookStepResponse, 0, len(item.Steps))
	for _, step := range item.Steps {
		steps = append(steps, toMyHandbookStepResponse(step))
	}
	return employeeHandbookDetailsResponse{
		EmployeeHandbookID: item.EmployeeHandbookID,
		EmployeeID:         item.EmployeeID,
		FirstName:          item.FirstName,
		LastName:           item.LastName,
		Status:             item.Status,
		AssignedAt:         item.AssignedAt,
		StartedAt:          item.StartedAt,
		CompletedAt:        item.CompletedAt,
		DueAt:              item.DueAt,
		TemplateID:         item.TemplateID,
		TemplateTitle:      item.TemplateTitle,
		TemplateDesc:       item.TemplateDesc,
		TemplateVersion:    item.TemplateVersion,
		DepartmentID:       item.DepartmentID,
		DepartmentName:     item.DepartmentName,
		Steps:              steps,
	}
}
