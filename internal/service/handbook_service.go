package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/goccy/go-json"

	"hrbackend/internal/domain"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
)

type HandbookService struct {
	repository domain.HandbookRepository
	logger     domain.Logger
}

func NewHandbookService(
	repository domain.HandbookRepository,
	logger domain.Logger,
) domain.HandbookService {
	return &HandbookService{
		repository: repository,
		logger:     logger,
	}
}

func (s *HandbookService) GetMyActiveHandbook(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.MyActiveHandbook, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	handbook, err := s.repository.GetActiveEmployeeHandbookByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	steps, err := s.repository.ListEmployeeHandbookStepsByHandbookID(ctx, handbook.HandbookID)
	if err != nil {
		return nil, err
	}

	handbook.Steps = steps
	return handbook, nil
}

func (s *HandbookService) StartMyHandbook(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.StartedHandbook, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	handbook, err := s.repository.GetActiveEmployeeHandbookByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	var started *domain.EmployeeHandbookAssignment
	err = s.repository.WithTx(ctx, func(tx domain.HandbookRepository) error {
		started, err = tx.MarkEmployeeHandbookStarted(ctx, handbook.HandbookID)
		if err != nil {
			return err
		}

		if handbook.Status == "not_started" {
			metadata, err := marshalJSONPayload(map[string]any{
				"source": "employee_self_service",
			})
			if err != nil {
				return err
			}
			return tx.CreateEmployeeHandbookAssignmentHistory(
				ctx,
				domain.CreateAssignmentHistoryParams{
					EmployeeHandbookID: &handbook.HandbookID,
					EmployeeID:         employeeID,
					TemplateID:         handbook.TemplateID,
					TemplateVersion:    handbook.TemplateVersion,
					Event:              "started",
					ActorEmployeeID:    &employeeID,
					Metadata:           metadata,
				},
			)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.StartedHandbook{
		HandbookID: started.EmployeeHandbookID,
		Status:     started.Status,
		StartedAt:  started.StartedAt,
	}, nil
}

func (s *HandbookService) CompleteMyHandbookStep(
	ctx context.Context,
	employeeID, stepID uuid.UUID,
	response []byte,
) (*domain.CompletedHandbookStep, error) {
	if employeeID == uuid.Nil || stepID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	handbook, err := s.repository.GetActiveEmployeeHandbookByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		response = []byte("null")
	}

	if err := json.Unmarshal(response, new(any)); err != nil {
		return nil, fmt.Errorf("%w: invalid response payload", domain.ErrHandbookInvalidRequest)
	}

	var completed *domain.CompletedHandbookStep
	err = s.repository.WithTx(ctx, func(tx domain.HandbookRepository) error {
		completed, err = tx.CompleteEmployeeHandbookStep(ctx, domain.CompleteHandbookStepParams{
			EmployeeHandbookID: handbook.HandbookID,
			StepID:             stepID,
			Response:           response,
		})
		if err != nil {
			return err
		}

		remaining, err := tx.CountRemainingRequiredHandbookSteps(ctx, handbook.HandbookID)
		if err != nil {
			return err
		}

		if remaining == 0 {
			finished, err := tx.MarkEmployeeHandbookCompleted(ctx, handbook.HandbookID)
			if err != nil {
				return err
			}
			completed.HandbookStatus = finished.Status

			metadata, err := marshalJSONPayload(map[string]any{
				"source":            "employee_self_service",
				"completed_step_id": stepID.String(),
			})
			if err != nil {
				return err
			}
			return tx.CreateEmployeeHandbookAssignmentHistory(
				ctx,
				domain.CreateAssignmentHistoryParams{
					EmployeeHandbookID: &handbook.HandbookID,
					EmployeeID:         employeeID,
					TemplateID:         handbook.TemplateID,
					TemplateVersion:    handbook.TemplateVersion,
					Event:              "completed",
					ActorEmployeeID:    &employeeID,
					Metadata:           metadata,
				},
			)
		}

		completed.HandbookStatus = handbook.Status
		return nil
	})
	if err != nil {
		return nil, err
	}

	return completed, nil
}

func (s *HandbookService) CreateTemplateForDepartment(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CreateTemplateForDepartmentParams,
) (*domain.HandbookTemplate, error) {
	if params.DepartmentID == uuid.Nil || strings.TrimSpace(params.Title) == "" {
		return nil, domain.ErrHandbookInvalidRequest
	}
	params.Title = strings.TrimSpace(params.Title)
	return s.repository.CreateHandbookTemplateForDepartment(ctx, actorEmployeeID, params)
}

func (s *HandbookService) CloneTemplateToDraft(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.CloneTemplateToDraftParams,
) (*domain.HandbookTemplate, error) {
	if params.SourceTemplateID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}
	return s.repository.CloneHandbookTemplateToDraft(ctx, actorEmployeeID, params)
}

func (s *HandbookService) UpdateTemplate(
	ctx context.Context,
	params domain.UpdateTemplateParams,
) (*domain.HandbookTemplate, error) {
	if params.TemplateID == uuid.Nil || (!params.SetTitle && !params.SetDescription) {
		return nil, domain.ErrHandbookInvalidRequest
	}
	if params.SetTitle {
		if params.Title == nil || strings.TrimSpace(*params.Title) == "" {
			return nil, domain.ErrHandbookInvalidRequest
		}
		title := strings.TrimSpace(*params.Title)
		params.Title = &title
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if tmpl.Status != "draft" {
		return nil, domain.ErrTemplateNotDraft
	}

	return s.repository.UpdateHandbookTemplateMetadata(ctx, params)
}

func (s *HandbookService) PublishTemplate(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.PublishTemplateParams,
) (*domain.HandbookTemplate, error) {
	if params.TemplateID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if tmpl.Status != "draft" {
		return nil, domain.ErrTemplateNotDraft
	}

	stepCount, err := s.repository.CountHandbookStepsByTemplateID(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if stepCount == 0 {
		return nil, domain.ErrTemplateHasNoSteps
	}

	return s.repository.PublishHandbookTemplate(ctx, actorEmployeeID, params)
}

func (s *HandbookService) ListTemplatesByDepartment(
	ctx context.Context,
	departmentID uuid.UUID,
) ([]domain.HandbookTemplate, error) {
	if departmentID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}
	return s.repository.ListHandbookTemplatesByDepartment(ctx, departmentID)
}

func (s *HandbookService) CreateStep(
	ctx context.Context,
	params domain.CreateStepParams,
) (*domain.HandbookStep, error) {
	if params.TemplateID == uuid.Nil || strings.TrimSpace(params.Title) == "" ||
		params.SortOrder <= 0 {
		return nil, domain.ErrHandbookInvalidRequest
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if tmpl.Status != "draft" {
		return nil, domain.ErrTemplateNotDraft
	}

	params.Title = strings.TrimSpace(params.Title)
	params.Kind = normalizeHandbookKind(params.Kind)
	params.Content, err = normalizeAndValidateContent(params.Kind, params.Content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidStepContent, err)
	}

	return s.repository.CreateHandbookStep(ctx, params)
}

func (s *HandbookService) UpdateStep(
	ctx context.Context,
	params domain.UpdateStepParams,
) (*domain.HandbookStep, error) {
	if params.StepID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	step, err := s.repository.GetHandbookStepByID(ctx, params.StepID)
	if err != nil {
		return nil, err
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, step.TemplateID)
	if err != nil {
		return nil, err
	}
	if tmpl.Status != "draft" {
		return nil, domain.ErrTemplateNotDraft
	}

	if params.SetTitle {
		if params.Title == nil || strings.TrimSpace(*params.Title) == "" {
			return nil, domain.ErrHandbookInvalidRequest
		}
		title := strings.TrimSpace(*params.Title)
		params.Title = &title
	}

	if params.ContentProvided {
		params.Content, err = normalizeAndValidateContent(step.Kind, params.Content)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrInvalidStepContent, err)
		}
	}

	return s.repository.UpdateHandbookStepByID(ctx, params)
}

func (s *HandbookService) DeleteStep(ctx context.Context, params domain.DeleteStepParams) error {
	if params.StepID == uuid.Nil {
		return domain.ErrHandbookInvalidRequest
	}

	step, err := s.repository.GetHandbookStepByID(ctx, params.StepID)
	if err != nil {
		return err
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, step.TemplateID)
	if err != nil {
		return err
	}
	if tmpl.Status != "draft" {
		return domain.ErrTemplateNotDraft
	}

	return s.repository.WithTx(ctx, func(tx domain.HandbookRepository) error {
		if err := tx.DeleteHandbookStepByID(ctx, params.StepID); err != nil {
			return err
		}

		remaining, err := tx.ListHandbookStepsByTemplate(ctx, step.TemplateID)
		if err != nil {
			return err
		}
		for i, item := range remaining {
			if err := tx.UpdateHandbookStepSortOrder(ctx, item.ID, int32(i+1)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *HandbookService) ReorderTemplateSteps(
	ctx context.Context,
	params domain.ReorderStepsParams,
) ([]domain.HandbookStep, error) {
	if params.TemplateID == uuid.Nil || len(params.OrderedStepIDs) == 0 {
		return nil, domain.ErrHandbookInvalidRequest
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if tmpl.Status != "draft" {
		return nil, domain.ErrTemplateNotDraft
	}

	currentSteps, err := s.repository.ListHandbookStepsByTemplate(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if len(currentSteps) != len(params.OrderedStepIDs) {
		return nil, domain.ErrInvalidStepReorder
	}

	allowed := make(map[uuid.UUID]struct{}, len(currentSteps))
	for _, item := range currentSteps {
		allowed[item.ID] = struct{}{}
	}

	seen := make(map[uuid.UUID]struct{}, len(params.OrderedStepIDs))
	for _, id := range params.OrderedStepIDs {
		if _, ok := allowed[id]; !ok {
			return nil, domain.ErrInvalidStepReorder
		}
		if _, ok := seen[id]; ok {
			return nil, domain.ErrInvalidStepReorder
		}
		seen[id] = struct{}{}
	}

	err = s.repository.WithTx(ctx, func(tx domain.HandbookRepository) error {
		for i, stepID := range params.OrderedStepIDs {
			if err := tx.UpdateHandbookStepSortOrder(ctx, stepID, -int32(i+1)); err != nil {
				return err
			}
		}
		for i, stepID := range params.OrderedStepIDs {
			if err := tx.UpdateHandbookStepSortOrder(ctx, stepID, int32(i+1)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repository.ListHandbookStepsByTemplate(ctx, params.TemplateID)
}

func (s *HandbookService) ListStepsByTemplate(
	ctx context.Context,
	templateID uuid.UUID,
) ([]domain.HandbookStep, error) {
	if templateID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}
	return s.repository.ListHandbookStepsByTemplate(ctx, templateID)
}

func (s *HandbookService) AssignTemplateToEmployee(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.AssignTemplateToEmployeeParams,
) (*domain.EmployeeHandbookAssignment, error) {
	if params.EmployeeID == uuid.Nil || params.TemplateID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	tmpl, err := s.repository.GetHandbookTemplateByID(ctx, params.TemplateID)
	if err != nil {
		return nil, err
	}
	if tmpl.Status != "published" {
		return nil, domain.ErrTemplateNotPublished
	}

	var previousActive *domain.MyActiveHandbook
	current, err := s.repository.GetActiveEmployeeHandbookByEmployeeID(ctx, params.EmployeeID)
	if err == nil {
		previousActive = current
	} else if !isActiveHandbookNotFound(err) {
		return nil, err
	}

	var assigned *domain.EmployeeHandbookAssignment
	err = s.repository.WithTx(ctx, func(tx domain.HandbookRepository) error {
		if err := tx.WaiveActiveEmployeeHandbooksByEmployeeID(ctx, params.EmployeeID); err != nil {
			return err
		}

		assigned, err = tx.CreateEmployeeHandbookFromTemplate(ctx, actorEmployeeID, params)
		if err != nil {
			return err
		}

		if previousActive != nil {
			metadata, err := marshalJSONPayload(map[string]any{
				"source":                  "manual_assignment",
				"replaced_by_handbook_id": assigned.EmployeeHandbookID.String(),
				"new_template_id":         assigned.TemplateID.String(),
			})
			if err != nil {
				return err
			}
			if err := tx.CreateEmployeeHandbookAssignmentHistory(
				ctx,
				domain.CreateAssignmentHistoryParams{
					EmployeeHandbookID: &previousActive.HandbookID,
					EmployeeID:         previousActive.EmployeeID,
					TemplateID:         previousActive.TemplateID,
					TemplateVersion:    previousActive.TemplateVersion,
					Event:              "reassigned",
					ActorEmployeeID:    ptr.UUIDOrNil(actorEmployeeID),
					Metadata:           metadata,
				},
			); err != nil {
				return err
			}
		}

		metadata, err := marshalJSONPayload(map[string]any{
			"source":               "manual_assignment",
			"previous_handbook_id": activeHandbookIDString(previousActive),
			"previous_template_id": activeTemplateIDString(previousActive),
		})
		if err != nil {
			return err
		}

		return tx.CreateEmployeeHandbookAssignmentHistory(ctx, domain.CreateAssignmentHistoryParams{
			EmployeeHandbookID: &assigned.EmployeeHandbookID,
			EmployeeID:         assigned.EmployeeID,
			TemplateID:         assigned.TemplateID,
			TemplateVersion:    assigned.TemplateVersion,
			Event:              "assigned",
			ActorEmployeeID:    ptr.UUIDOrNil(actorEmployeeID),
			Metadata:           metadata,
		})
	})
	if err != nil {
		return nil, err
	}

	return assigned, nil
}

func (s *HandbookService) WaiveEmployeeHandbook(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.WaiveEmployeeHandbookParams,
) (*domain.WaivedEmployeeHandbook, error) {
	if params.EmployeeHandbookID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	handbook, err := s.repository.GetEmployeeHandbookByID(ctx, params.EmployeeHandbookID)
	if err != nil {
		return nil, err
	}
	if handbook.Status != "not_started" && handbook.Status != "in_progress" {
		return nil, domain.ErrEmployeeHandbookNotActive
	}

	var waived *domain.WaivedEmployeeHandbook
	err = s.repository.WithTx(ctx, func(tx domain.HandbookRepository) error {
		waived, err = tx.WaiveEmployeeHandbookByID(ctx, params.EmployeeHandbookID)
		if err != nil {
			return err
		}

		metadata := map[string]any{"source": "manual_waive"}
		if params.Reason != nil && strings.TrimSpace(*params.Reason) != "" {
			metadata["reason"] = strings.TrimSpace(*params.Reason)
		}

		payload, err := marshalJSONPayload(metadata)
		if err != nil {
			return err
		}

		return tx.CreateEmployeeHandbookAssignmentHistory(ctx, domain.CreateAssignmentHistoryParams{
			EmployeeHandbookID: &waived.EmployeeHandbookID,
			EmployeeID:         waived.EmployeeID,
			TemplateID:         handbook.TemplateID,
			TemplateVersion:    handbook.TemplateVersion,
			Event:              "waived",
			ActorEmployeeID:    ptr.UUIDOrNil(actorEmployeeID),
			Metadata:           payload,
		})
	})
	if err != nil {
		return nil, err
	}

	return waived, nil
}

func (s *HandbookService) ListEmployeeHandbookHistory(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.HandbookAssignmentHistoryEntry, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}
	return s.repository.ListEmployeeHandbookAssignmentHistoryByEmployeeID(ctx, employeeID, 50, 0)
}

func (s *HandbookService) ListEligibleEmployees(
	ctx context.Context,
	actorEmployeeID uuid.UUID,
	params domain.ListEligibleEmployeesParams,
) (*domain.EligibleEmployeePage, error) {
	if actorEmployeeID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}

	userID, err := s.repository.GetUserIDByEmployeeID(ctx, actorEmployeeID)
	if err != nil {
		return nil, err
	}

	hasViewAll, err := s.repository.CheckUserPermission(
		ctx,
		userID,
		"HANDBOOK.ELIGIBLE_EMPLOYEES.VIEW_ALL",
	)
	if err != nil {
		return nil, domain.ErrEligibleEmployeePermissionCheck
	}

	if !hasViewAll {
		profile, err := s.repository.GetEmployeeProfileByID(ctx, actorEmployeeID)
		if err != nil {
			return nil, err
		}
		if profile.DepartmentID == nil {
			return &domain.EligibleEmployeePage{
				Items:      []domain.EligibleEmployee{},
				TotalCount: 0,
			}, nil
		}
		params.DepartmentID = profile.DepartmentID
	}

	params.Search = ptr.TrimString(params.Search)
	return s.repository.ListEligibleEmployeesForHandbookAssignment(ctx, params)
}

func (s *HandbookService) ListEmployeeHandbookAssignments(
	ctx context.Context,
	params domain.ListEmployeeHandbookAssignmentsParams,
) (*domain.EmployeeHandbookAssignmentPage, error) {
	if params.Status != nil {
		normalized := strings.TrimSpace(strings.ToLower(*params.Status))
		switch normalized {
		case "", "unassigned", "not_started", "in_progress", "completed", "waived":
		default:
			return nil, domain.ErrInvalidAssignmentStatusFilter
		}
		params.Status = &normalized
	}
	params.Search = ptr.TrimString(params.Search)
	return s.repository.ListEmployeeHandbookAssignments(ctx, params)
}

func (s *HandbookService) GetEmployeeHandbookDetails(
	ctx context.Context,
	handbookID uuid.UUID,
) (*domain.EmployeeHandbookDetails, error) {
	if handbookID == uuid.Nil {
		return nil, domain.ErrHandbookInvalidRequest
	}
	return s.repository.GetEmployeeHandbookDetailsByID(ctx, handbookID)
}

func normalizeHandbookKind(kind string) string {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "rich_text":
		return "content"
	case "content", "ack", "link", "quiz":
		return strings.TrimSpace(strings.ToLower(kind))
	default:
		return strings.TrimSpace(strings.ToLower(kind))
	}
}

func normalizeAndValidateContent(kind string, raw []byte) ([]byte, error) {
	switch kind {
	case "link":
		if len(raw) == 0 || string(raw) == "null" {
			return nil, fmt.Errorf("link content is required")
		}

		var content struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(raw, &content); err != nil {
			return nil, fmt.Errorf("invalid link content: %w", err)
		}

		urlStr := strings.TrimSpace(content.URL)
		if urlStr == "" {
			return nil, fmt.Errorf("link URL is required")
		}

		parsed, err := url.Parse(urlStr)
		if err != nil || parsed.Host == "" {
			return nil, fmt.Errorf("invalid link URL")
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return nil, fmt.Errorf("URL must use http or https")
		}

		return json.Marshal(map[string]any{"url": urlStr})

	case "quiz":
		if len(raw) == 0 || string(raw) == "null" {
			return nil, fmt.Errorf("quiz content is required")
		}

		var content struct {
			Question           string   `json:"question"`
			Options            []string `json:"options"`
			CorrectOptionIndex int      `json:"correct_option_index"`
		}
		if err := json.Unmarshal(raw, &content); err != nil {
			return nil, fmt.Errorf("invalid quiz content: %w", err)
		}
		if strings.TrimSpace(content.Question) == "" {
			return nil, fmt.Errorf("quiz question is required")
		}
		if len(content.Options) < 2 {
			return nil, fmt.Errorf("quiz must have at least 2 options")
		}
		if content.CorrectOptionIndex < 0 || content.CorrectOptionIndex >= len(content.Options) {
			return nil, fmt.Errorf("correct_option_index out of range")
		}

		options := make([]string, 0, len(content.Options))
		for _, option := range content.Options {
			trimmed := strings.TrimSpace(option)
			if trimmed == "" {
				return nil, fmt.Errorf("quiz options must be non-empty")
			}
			options = append(options, trimmed)
		}

		return json.Marshal(map[string]any{
			"question":             strings.TrimSpace(content.Question),
			"options":              options,
			"correct_option_index": content.CorrectOptionIndex,
		})

	default:
		if len(raw) == 0 || string(raw) == "null" {
			return []byte("{}"), nil
		}
		return []byte("{}"), nil
	}
}

func marshalJSONPayload(v any) ([]byte, error) {
	if v == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(v)
}

func activeHandbookIDString(item *domain.MyActiveHandbook) string {
	if item == nil {
		return ""
	}
	return item.HandbookID.String()
}

func activeTemplateIDString(item *domain.MyActiveHandbook) string {
	if item == nil {
		return ""
	}
	return item.TemplateID.String()
}

func isActiveHandbookNotFound(err error) bool {
	return err == domain.ErrActiveHandbookNotFound
}
