package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type HandbookRepository struct {
	store *db.Store
}

func NewHandbookRepository(store *db.Store) domain.HandbookRepository {
	return &HandbookRepository{
		store: store,
	}
}

func (r *HandbookRepository) WithTx(ctx context.Context, fn func(tx domain.HandbookRepository) error) error {
	if r.store == nil {
		return errors.New("handbook repository transaction store is not configured")
	}

	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&HandbookRepository{
			store: &db.Store{
				Queries:  q,
				ConnPool: r.store.ConnPool,
			},
		})
	})
}

func (r *HandbookRepository) GetActiveEmployeeHandbookByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*domain.MyActiveHandbook, error) {
	row, err := r.store.GetActiveEmployeeHandbookByEmployeeID(ctx, employeeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrActiveHandbookNotFound
		}
		return nil, err
	}
	return toDomainMyActiveHandbook(row), nil
}

func (r *HandbookRepository) ListEmployeeHandbookStepsByHandbookID(ctx context.Context, handbookID uuid.UUID) ([]domain.MyHandbookStep, error) {
	rows, err := r.store.ListEmployeeHandbookStepsByHandbookID(ctx, handbookID)
	if err != nil {
		return nil, err
	}

	steps := make([]domain.MyHandbookStep, 0, len(rows))
	for _, row := range rows {
		steps = append(steps, toDomainMyHandbookStep(row))
	}
	return steps, nil
}

func (r *HandbookRepository) MarkEmployeeHandbookStarted(ctx context.Context, handbookID uuid.UUID) (*domain.EmployeeHandbookAssignment, error) {
	row, err := r.store.MarkEmployeeHandbookStarted(ctx, handbookID)
	if err != nil {
		return nil, err
	}
	model := toDomainEmployeeHandbookAssignment(row)
	return &model, nil
}

func (r *HandbookRepository) CompleteEmployeeHandbookStep(ctx context.Context, params domain.CompleteHandbookStepParams) (*domain.CompletedHandbookStep, error) {
	row, err := r.store.CompleteEmployeeHandbookStep(ctx, db.CompleteEmployeeHandbookStepParams{
		Response:           params.Response,
		EmployeeHandbookID: params.EmployeeHandbookID,
		StepID:             params.StepID,
	})
	if err != nil {
		return nil, err
	}
	return &domain.CompletedHandbookStep{
		HandbookID:  row.EmployeeHandbookID,
		StepID:      row.StepID,
		StepStatus:  string(row.Status),
		CompletedAt: conv.TimeFromPgTimestamptz(row.CompletedAt),
	}, nil
}

func (r *HandbookRepository) CountRemainingRequiredHandbookSteps(ctx context.Context, handbookID uuid.UUID) (int32, error) {
	return r.store.CountRemainingRequiredHandbookSteps(ctx, handbookID)
}

func (r *HandbookRepository) MarkEmployeeHandbookCompleted(ctx context.Context, handbookID uuid.UUID) (*domain.EmployeeHandbookAssignment, error) {
	row, err := r.store.MarkEmployeeHandbookCompleted(ctx, handbookID)
	if err != nil {
		return nil, err
	}
	model := toDomainEmployeeHandbookAssignment(row)
	return &model, nil
}

func (r *HandbookRepository) CreateEmployeeHandbookAssignmentHistory(ctx context.Context, params domain.CreateAssignmentHistoryParams) error {
	_, err := r.store.CreateEmployeeHandbookAssignmentHistory(ctx, db.CreateEmployeeHandbookAssignmentHistoryParams{
		EmployeeHandbookID: params.EmployeeHandbookID,
		EmployeeID:         params.EmployeeID,
		TemplateID:         params.TemplateID,
		TemplateVersion:    params.TemplateVersion,
		Event:              db.HandbookAssignmentEventEnum(params.Event),
		ActorEmployeeID:    params.ActorEmployeeID,
		Metadata:           params.Metadata,
	})
	return err
}

func (r *HandbookRepository) CreateHandbookTemplateForDepartment(ctx context.Context, actorEmployeeID uuid.UUID, params domain.CreateTemplateForDepartmentParams) (*domain.HandbookTemplate, error) {
	row, err := r.store.CreateHandbookTemplateForDepartment(ctx, db.CreateHandbookTemplateForDepartmentParams{
		DepartmentID:        params.DepartmentID,
		Title:               params.Title,
		Description:         params.Description,
		CreatedByEmployeeID: uuidPtrOrNil(actorEmployeeID),
	})
	if err != nil {
		if isDraftTemplateUniqueViolation(err) {
			return nil, domain.ErrDraftTemplateAlreadyExists
		}
		return nil, err
	}
	model := toDomainHandbookTemplate(row)
	return &model, nil
}

func (r *HandbookRepository) CloneHandbookTemplateToDraft(ctx context.Context, actorEmployeeID uuid.UUID, params domain.CloneTemplateToDraftParams) (*domain.HandbookTemplate, error) {
	row, err := r.store.CloneHandbookTemplateToDraft(ctx, db.CloneHandbookTemplateToDraftParams{
		SourceTemplateID:    params.SourceTemplateID,
		CreatedByEmployeeID: uuidPtrOrNil(actorEmployeeID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		if isDraftTemplateUniqueViolation(err) {
			return nil, domain.ErrDraftTemplateAlreadyExists
		}
		return nil, err
	}
	model := toDomainClonedHandbookTemplate(row)
	return &model, nil
}

func (r *HandbookRepository) GetHandbookTemplateByID(ctx context.Context, templateID uuid.UUID) (*domain.HandbookTemplate, error) {
	row, err := r.store.GetHandbookTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, err
	}
	model := toDomainHandbookTemplate(row)
	return &model, nil
}

func (r *HandbookRepository) UpdateHandbookTemplateMetadata(ctx context.Context, params domain.UpdateTemplateParams) (*domain.HandbookTemplate, error) {
	row, err := r.store.UpdateHandbookTemplateMetadata(ctx, db.UpdateHandbookTemplateMetadataParams{
		TemplateID:     params.TemplateID,
		Title:          params.Title,
		SetTitle:       params.SetTitle,
		Description:    params.Description,
		SetDescription: params.SetDescription,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, err
	}
	model := toDomainHandbookTemplate(row)
	return &model, nil
}

func (r *HandbookRepository) CountHandbookStepsByTemplateID(ctx context.Context, templateID uuid.UUID) (int32, error) {
	return r.store.CountHandbookStepsByTemplateID(ctx, templateID)
}

func (r *HandbookRepository) PublishHandbookTemplate(ctx context.Context, actorEmployeeID uuid.UUID, params domain.PublishTemplateParams) (*domain.HandbookTemplate, error) {
	row, err := r.store.PublishHandbookTemplate(ctx, db.PublishHandbookTemplateParams{
		TemplateID:            params.TemplateID,
		PublishedByEmployeeID: uuidPtrOrNil(actorEmployeeID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, err
	}
	model := toDomainHandbookTemplate(row)
	return &model, nil
}

func (r *HandbookRepository) ListHandbookTemplatesByDepartment(ctx context.Context, departmentID uuid.UUID) ([]domain.HandbookTemplate, error) {
	rows, err := r.store.ListHandbookTemplatesByDepartment(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.HandbookTemplate, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainHandbookTemplate(row))
	}
	return items, nil
}

func (r *HandbookRepository) CreateHandbookStep(ctx context.Context, params domain.CreateStepParams) (*domain.HandbookStep, error) {
	row, err := r.store.CreateHandbookStep(ctx, db.CreateHandbookStepParams{
		TemplateID: params.TemplateID,
		SortOrder:  params.SortOrder,
		Kind:       db.HandbookStepKindEnum(params.Kind),
		Title:      params.Title,
		Body:       params.Body,
		Content:    params.Content,
		IsRequired: params.IsRequired,
	})
	if err != nil {
		return nil, err
	}
	model := toDomainHandbookStep(row)
	return &model, nil
}

func (r *HandbookRepository) ListHandbookStepsByTemplate(ctx context.Context, templateID uuid.UUID) ([]domain.HandbookStep, error) {
	rows, err := r.store.ListHandbookStepsByTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.HandbookStep, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainHandbookStep(row))
	}
	return items, nil
}

func (r *HandbookRepository) GetHandbookStepByID(ctx context.Context, stepID uuid.UUID) (*domain.HandbookStep, error) {
	row, err := r.store.GetHandbookStepByID(ctx, stepID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStepNotFound
		}
		return nil, err
	}
	model := toDomainHandbookStep(row)
	return &model, nil
}

func (r *HandbookRepository) UpdateHandbookStepByID(ctx context.Context, params domain.UpdateStepParams) (*domain.HandbookStep, error) {
	row, err := r.store.UpdateHandbookStepByID(ctx, db.UpdateHandbookStepByIDParams{
		StepID:        params.StepID,
		Title:         params.Title,
		SetTitle:      params.SetTitle,
		Body:          params.Body,
		SetBody:       params.SetBody,
		Content:       params.Content,
		SetContent:    params.ContentProvided,
		IsRequired:    params.IsRequired,
		SetIsRequired: params.SetIsRequired,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStepNotFound
		}
		return nil, err
	}
	model := toDomainHandbookStep(row)
	return &model, nil
}

func (r *HandbookRepository) DeleteHandbookStepByID(ctx context.Context, stepID uuid.UUID) error {
	return r.store.DeleteHandbookStepByID(ctx, stepID)
}

func (r *HandbookRepository) UpdateHandbookStepSortOrder(ctx context.Context, stepID uuid.UUID, sortOrder int32) error {
	return r.store.UpdateHandbookStepSortOrder(ctx, db.UpdateHandbookStepSortOrderParams{
		StepID:    stepID,
		SortOrder: sortOrder,
	})
}

func (r *HandbookRepository) WaiveActiveEmployeeHandbooksByEmployeeID(ctx context.Context, employeeID uuid.UUID) error {
	return r.store.WaiveActiveEmployeeHandbooksByEmployeeID(ctx, employeeID)
}

func (r *HandbookRepository) CreateEmployeeHandbookFromTemplate(ctx context.Context, actorEmployeeID uuid.UUID, params domain.AssignTemplateToEmployeeParams) (*domain.EmployeeHandbookAssignment, error) {
	row, err := r.store.CreateEmployeeHandbookFromTemplate(ctx, db.CreateEmployeeHandbookFromTemplateParams{
		EmployeeID:           params.EmployeeID,
		TemplateID:           params.TemplateID,
		AssignedByEmployeeID: uuidPtrOrNil(actorEmployeeID),
	})
	if err != nil {
		return nil, err
	}
	model := toDomainCreateEmployeeHandbookFromTemplate(row)
	return &model, nil
}

func (r *HandbookRepository) GetEmployeeHandbookByID(ctx context.Context, handbookID uuid.UUID) (*domain.EmployeeHandbookAssignment, error) {
	row, err := r.store.GetEmployeeHandbookByID(ctx, handbookID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeHandbookNotFound
		}
		return nil, err
	}
	model := toDomainEmployeeHandbookAssignment(row)
	return &model, nil
}

func (r *HandbookRepository) WaiveEmployeeHandbookByID(ctx context.Context, handbookID uuid.UUID) (*domain.WaivedEmployeeHandbook, error) {
	row, err := r.store.WaiveEmployeeHandbookByID(ctx, handbookID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeHandbookNotFound
		}
		return nil, err
	}
	return &domain.WaivedEmployeeHandbook{
		EmployeeHandbookID: row.ID,
		EmployeeID:         row.EmployeeID,
		Status:             string(row.Status),
		CompletedAt:        handbookTimePtrFromPgTimestamptz(row.CompletedAt),
	}, nil
}

func (r *HandbookRepository) ListEmployeeHandbookAssignmentHistoryByEmployeeID(ctx context.Context, employeeID uuid.UUID, limit, offset int32) ([]domain.HandbookAssignmentHistoryEntry, error) {
	rows, err := r.store.ListEmployeeHandbookAssignmentHistoryByEmployeeID(ctx, db.ListEmployeeHandbookAssignmentHistoryByEmployeeIDParams{
		EmployeeID: employeeID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.HandbookAssignmentHistoryEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.HandbookAssignmentHistoryEntry{
			ID:                 row.ID,
			EmployeeHandbookID: row.EmployeeHandbookID,
			EmployeeID:         row.EmployeeID,
			TemplateID:         row.TemplateID,
			TemplateVersion:    row.TemplateVersion,
			Event:              string(row.Event),
			ActorEmployeeID:    row.ActorEmployeeID,
			Metadata:           unmarshalJSON(row.Metadata),
			CreatedAt:          conv.TimeFromPgTimestamptz(row.CreatedAt),
		})
	}
	return items, nil
}

func (r *HandbookRepository) ListEmployeeHandbookAssignments(ctx context.Context, params domain.ListEmployeeHandbookAssignmentsParams) (*domain.EmployeeHandbookAssignmentPage, error) {
	rows, err := r.store.ListEmployeeHandbookAssignments(ctx, db.ListEmployeeHandbookAssignmentsParams{
		Limit:        params.Limit,
		Offset:       params.Offset,
		DepartmentID: params.DepartmentID,
		StatusFilter: normalizeAssignmentStatus(params.Status),
		Search:       handbookTrimStringPtr(params.Search),
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := r.store.CountEmployeeHandbookAssignments(ctx, db.CountEmployeeHandbookAssignmentsParams{
		DepartmentID: params.DepartmentID,
		StatusFilter: normalizeAssignmentStatus(params.Status),
		Search:       handbookTrimStringPtr(params.Search),
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.EmployeeHandbookAssignmentSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.EmployeeHandbookAssignmentSummary{
			EmployeeID:             row.EmployeeID,
			FirstName:              row.FirstName,
			LastName:               row.LastName,
			DepartmentID:           row.EmployeeDepartmentID,
			DepartmentName:         row.DepartmentName,
			EmployeeHandbookID:     row.EmployeeHandbookID,
			TemplateID:             row.HandbookTemplateID,
			TemplateTitle:          row.TemplateTitle,
			TemplateVersion:        row.TemplateVersion,
			HandbookStatus:         interfaceString(row.EmployeeHandbookStatus),
			AssignedAt:             handbookTimePtrFromPgTimestamptz(row.AssignedAt),
			StartedAt:              handbookTimePtrFromPgTimestamptz(row.StartedAt),
			CompletedAt:            handbookTimePtrFromPgTimestamptz(row.CompletedAt),
			DueAt:                  handbookTimePtrFromPgTimestamptz(row.DueAt),
			RequiredStepsTotal:     row.RequiredStepsTotal,
			RequiredStepsCompleted: row.RequiredStepsCompleted,
		})
	}

	return &domain.EmployeeHandbookAssignmentPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (r *HandbookRepository) GetEmployeeHandbookDetailsByID(ctx context.Context, handbookID uuid.UUID) (*domain.EmployeeHandbookDetails, error) {
	hb, err := r.store.GetEmployeeHandbookDetailsByID(ctx, handbookID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeHandbookNotFound
		}
		return nil, err
	}

	steps, err := r.ListEmployeeHandbookStepsByHandbookID(ctx, handbookID)
	if err != nil {
		return nil, err
	}

	return &domain.EmployeeHandbookDetails{
		EmployeeHandbookID: hb.ID,
		EmployeeID:         hb.EmployeeID,
		FirstName:          hb.FirstName,
		LastName:           hb.LastName,
		Status:             string(hb.Status),
		AssignedAt:         conv.TimeFromPgTimestamptz(hb.AssignedAt),
		StartedAt:          handbookTimePtrFromPgTimestamptz(hb.StartedAt),
		CompletedAt:        handbookTimePtrFromPgTimestamptz(hb.CompletedAt),
		DueAt:              handbookTimePtrFromPgTimestamptz(hb.DueAt),
		TemplateID:         hb.TemplateID,
		TemplateTitle:      hb.TemplateTitle,
		TemplateDesc:       hb.TemplateDescription,
		TemplateVersion:    hb.TemplateVersion,
		DepartmentID:       hb.DepartmentID,
		DepartmentName:     hb.DepartmentName,
		Steps:              steps,
	}, nil
}

func (r *HandbookRepository) GetUserIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error) {
	return r.store.GetUserIDByEmployeeID(ctx, employeeID)
}

func (r *HandbookRepository) CheckUserPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	return r.store.CheckUserPermission(ctx, db.CheckUserPermissionParams{
		UserID: userID,
		Name:   permission,
	})
}

func (r *HandbookRepository) GetEmployeeProfileByID(ctx context.Context, employeeID uuid.UUID) (*domain.HandbookEmployeeProfile, error) {
	row, err := r.store.GetEmployeeProfileByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return &domain.HandbookEmployeeProfile{
		ID:           row.ID,
		DepartmentID: row.DepartmentID,
	}, nil
}

func (r *HandbookRepository) ListEligibleEmployeesForHandbookAssignment(ctx context.Context, params domain.ListEligibleEmployeesParams) (*domain.EligibleEmployeePage, error) {
	search := handbookTrimStringPtr(params.Search)
	rows, err := r.store.ListEligibleEmployeesForHandbookAssignment(ctx, db.ListEligibleEmployeesForHandbookAssignmentParams{
		Limit:        params.Limit,
		Offset:       params.Offset,
		DepartmentID: params.DepartmentID,
		Search:       search,
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := r.store.CountEligibleEmployeesForHandbookAssignment(ctx, db.CountEligibleEmployeesForHandbookAssignmentParams{
		DepartmentID: params.DepartmentID,
		Search:       search,
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.EligibleEmployee, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.EligibleEmployee{
			EmployeeID:     row.EmployeeID,
			FirstName:      row.FirstName,
			LastName:       row.LastName,
			DepartmentID:   row.DepartmentID,
			DepartmentName: row.DepartmentName,
		})
	}

	return &domain.EligibleEmployeePage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func toDomainHandbookTemplate(t db.HandbookTemplate) domain.HandbookTemplate {
	return domain.HandbookTemplate{
		ID:           t.ID,
		DepartmentID: t.DepartmentID,
		Title:        t.Title,
		Description:  t.Description,
		Version:      t.Version,
		Status:       string(t.Status),
		PublishedAt:  handbookTimePtrFromPgTimestamptz(t.PublishedAt),
		ArchivedAt:   handbookTimePtrFromPgTimestamptz(t.ArchivedAt),
		CreatedAt:    conv.TimeFromPgTimestamptz(t.CreatedAt),
		UpdatedAt:    conv.TimeFromPgTimestamptz(t.UpdatedAt),
	}
}

func toDomainClonedHandbookTemplate(t db.CloneHandbookTemplateToDraftRow) domain.HandbookTemplate {
	return domain.HandbookTemplate{
		ID:           t.ID,
		DepartmentID: t.DepartmentID,
		Title:        t.Title,
		Description:  t.Description,
		Version:      t.Version,
		Status:       string(t.Status),
		PublishedAt:  handbookTimePtrFromPgTimestamptz(t.PublishedAt),
		ArchivedAt:   handbookTimePtrFromPgTimestamptz(t.ArchivedAt),
		CreatedAt:    conv.TimeFromPgTimestamptz(t.CreatedAt),
		UpdatedAt:    conv.TimeFromPgTimestamptz(t.UpdatedAt),
	}
}

func toDomainHandbookStep(step db.HandbookStep) domain.HandbookStep {
	return domain.HandbookStep{
		ID:         step.ID,
		TemplateID: step.TemplateID,
		SortOrder:  step.SortOrder,
		Kind:       string(step.Kind),
		Title:      step.Title,
		Body:       step.Body,
		Content:    unmarshalJSON(step.Content),
		IsRequired: step.IsRequired,
		UpdatedAt:  conv.TimeFromPgTimestamptz(step.UpdatedAt),
	}
}

func toDomainMyActiveHandbook(row db.GetActiveEmployeeHandbookByEmployeeIDRow) *domain.MyActiveHandbook {
	return &domain.MyActiveHandbook{
		HandbookID:      row.ID,
		EmployeeID:      row.EmployeeID,
		Status:          string(row.Status),
		AssignedAt:      conv.TimeFromPgTimestamptz(row.AssignedAt),
		StartedAt:       handbookTimePtrFromPgTimestamptz(row.StartedAt),
		CompletedAt:     handbookTimePtrFromPgTimestamptz(row.CompletedAt),
		DueAt:           handbookTimePtrFromPgTimestamptz(row.DueAt),
		TemplateID:      row.TemplateID,
		TemplateTitle:   row.TemplateTitle,
		TemplateDesc:    row.TemplateDescription,
		TemplateVersion: row.TemplateVersion,
		DepartmentID:    row.DepartmentID,
		DepartmentName:  row.DepartmentName,
	}
}

func toDomainMyHandbookStep(row db.ListEmployeeHandbookStepsByHandbookIDRow) domain.MyHandbookStep {
	return domain.MyHandbookStep{
		StepID:      row.StepID,
		SortOrder:   row.SortOrder,
		Kind:        string(row.Kind),
		Title:       row.Title,
		Body:        row.Body,
		Content:     unmarshalJSON(row.Content),
		IsRequired:  row.IsRequired,
		Status:      string(row.ProgressStatus),
		StartedAt:   handbookTimePtrFromPgTimestamptz(row.ProgressStartedAt),
		CompletedAt: handbookTimePtrFromPgTimestamptz(row.ProgressCompletedAt),
		Response:    unmarshalJSON(row.ProgressResponse),
	}
}

func toDomainEmployeeHandbookAssignment(row db.EmployeeHandbook) domain.EmployeeHandbookAssignment {
	return domain.EmployeeHandbookAssignment{
		EmployeeHandbookID: row.ID,
		EmployeeID:         row.EmployeeID,
		TemplateID:         row.TemplateID,
		TemplateVersion:    row.TemplateVersion,
		AssignedAt:         conv.TimeFromPgTimestamptz(row.AssignedAt),
		StartedAt:          handbookTimePtrFromPgTimestamptz(row.StartedAt),
		CompletedAt:        handbookTimePtrFromPgTimestamptz(row.CompletedAt),
		DueAt:              handbookTimePtrFromPgTimestamptz(row.DueAt),
		Status:             string(row.Status),
	}
}

func toDomainCreateEmployeeHandbookFromTemplate(row db.CreateEmployeeHandbookFromTemplateRow) domain.EmployeeHandbookAssignment {
	return domain.EmployeeHandbookAssignment{
		EmployeeHandbookID: row.ID,
		EmployeeID:         row.EmployeeID,
		TemplateID:         row.TemplateID,
		TemplateVersion:    row.TemplateVersion,
		AssignedAt:         conv.TimeFromPgTimestamptz(row.AssignedAt),
		StartedAt:          handbookTimePtrFromPgTimestamptz(row.StartedAt),
		CompletedAt:        handbookTimePtrFromPgTimestamptz(row.CompletedAt),
		DueAt:              handbookTimePtrFromPgTimestamptz(row.DueAt),
		Status:             string(row.Status),
	}
}

func handbookTimePtrFromPgTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func uuidPtrOrNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func normalizeAssignmentStatus(status *string) *string {
	if status == nil {
		return nil
	}
	trimmed := strings.TrimSpace(strings.ToLower(*status))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func handbookTrimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func interfaceString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(v)
	}
}

func unmarshalJSON(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func isDraftTemplateUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "handbook_templates_one_draft_per_department"
}
