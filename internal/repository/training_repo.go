package repository

import (
	"context"
	"errors"
	"strings"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TrainingRepository struct {
	queries db.Querier
}

func NewTrainingRepository(queries db.Querier) domain.TrainingRepository {
	return &TrainingRepository{queries: queries}
}

func (r *TrainingRepository) AssignTrainingToEmployee(
	ctx context.Context,
	params domain.AssignTrainingToEmployeeParams,
) (*domain.EmployeeTrainingAssignment, error) {
	row, err := r.queries.AssignTrainingToEmployee(ctx, db.AssignTrainingToEmployeeParams{
		EmployeeID:           params.EmployeeID,
		TrainingID:           params.TrainingID,
		AssignedByEmployeeID: params.AssignedByEmployeeID,
		DueAt:                conv.PgTimestamptzFromTime(params.DueAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrainingInvalidRequest
		}
		if isTrainingAssignmentUniqueViolation(err) {
			return nil, domain.ErrTrainingAssignmentConflict
		}
		return nil, err
	}

	return toDomainEmployeeTrainingAssignment(row), nil
}

func (r *TrainingRepository) CancelTrainingAssignment(
	ctx context.Context,
	params domain.CancelTrainingAssignmentParams,
) (*domain.EmployeeTrainingAssignment, error) {
	row, err := r.queries.CancelTrainingAssignment(ctx, db.CancelTrainingAssignmentParams{
		ID:                 params.AssignmentID,
		CancellationReason: params.CancellationReason,
	})
	if err == nil {
		return toDomainEmployeeTrainingAssignment(row), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	existing, getErr := r.queries.GetTrainingAssignmentByID(ctx, params.AssignmentID)
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return nil, domain.ErrTrainingAssignmentNotFound
		}
		return nil, getErr
	}

	switch string(existing.Status) {
	case "assigned", "in_progress":
		return nil, err
	default:
		return nil, domain.ErrTrainingAssignmentNotCancellable
	}
}

func (r *TrainingRepository) ListTrainingAssignments(
	ctx context.Context,
	params domain.ListTrainingAssignmentsParams,
) (*domain.TrainingAssignmentPage, error) {
	rows, err := r.queries.ListTrainingAssignmentsPaginated(
		ctx,
		db.ListTrainingAssignmentsPaginatedParams{
			Limit:          params.Limit,
			Offset:         params.Offset,
			EmployeeSearch: params.EmployeeSearch,
			DepartmentID:   params.DepartmentID,
			TrainingID:     params.TrainingID,
			StatusFilter:   params.Status,
		},
	)
	if err != nil {
		return nil, err
	}

	page := &domain.TrainingAssignmentPage{
		Items: make([]domain.TrainingAssignmentListItem, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, domain.TrainingAssignmentListItem{
			AssignmentID:         row.AssignmentID,
			EmployeeID:           row.EmployeeID,
			EmployeeNumber:       row.EmployeeNumber,
			EmploymentNumber:     row.EmploymentNumber,
			FirstName:            row.FirstName,
			LastName:             row.LastName,
			DepartmentID:         row.DepartmentID,
			DepartmentName:       row.DepartmentName,
			TrainingID:           row.TrainingID,
			TrainingTitle:        row.TrainingTitle,
			TrainingCategory:     row.TrainingCategory,
			Status:               row.Status,
			AssignedAt:           conv.TimeFromPgTimestamptz(row.AssignedAt),
			DueAt:                timePtrFromPgTimestamptz(row.DueAt),
			StartedAt:            timePtrFromPgTimestamptz(row.StartedAt),
			CompletedAt:          timePtrFromPgTimestamptz(row.CompletedAt),
			AssignedByEmployeeID: row.AssignedByEmployeeID,
			AssignedByName:       trimStringPtr(&row.AssignedByName),
			IsOverdue:            row.IsOverdue,
		})
	}

	return page, nil
}

func (r *TrainingRepository) CreateTrainingCatalogItem(
	ctx context.Context,
	params domain.CreateTrainingCatalogItemParams,
) (*domain.TrainingCatalogItem, error) {
	item, err := r.queries.CreateTrainingCatalogItem(ctx, db.CreateTrainingCatalogItemParams{
		Title:                    params.Title,
		Description:              params.Description,
		Category:                 params.Category,
		EstimatedDurationMinutes: params.EstimatedDurationMinutes,
		CreatedByEmployeeID:      params.CreatedByEmployeeID,
	})
	if err != nil {
		return nil, err
	}

	return toDomainTrainingCatalogItem(item), nil
}

func (r *TrainingRepository) ListTrainingCatalogItems(
	ctx context.Context,
	params domain.ListTrainingCatalogItemsParams,
) (*domain.TrainingCatalogItemPage, error) {
	rows, err := r.queries.ListTrainingCatalogItemsPaginated(
		ctx,
		db.ListTrainingCatalogItemsPaginatedParams{
			Search:   params.Search,
			IsActive: params.IsActive,
			Offset:   params.Offset,
			Limit:    params.Limit,
		},
	)
	if err != nil {
		return nil, err
	}

	page := &domain.TrainingCatalogItemPage{
		Items: make([]domain.TrainingCatalogItem, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, domain.TrainingCatalogItem{
			ID:                       row.ID,
			Title:                    row.Title,
			Description:              row.Description,
			Category:                 row.Category,
			EstimatedDurationMinutes: row.EstimatedDurationMinutes,
			IsActive:                 row.IsActive,
			CreatedByEmployeeID:      row.CreatedByEmployeeID,
			CreatedAt:                conv.TimeFromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                conv.TimeFromPgTimestamptz(row.UpdatedAt),
		})
	}

	return page, nil
}

func toDomainTrainingCatalogItem(item db.TrainingCatalogItem) *domain.TrainingCatalogItem {
	return &domain.TrainingCatalogItem{
		ID:                       item.ID,
		Title:                    item.Title,
		Description:              item.Description,
		Category:                 item.Category,
		EstimatedDurationMinutes: item.EstimatedDurationMinutes,
		IsActive:                 item.IsActive,
		CreatedByEmployeeID:      item.CreatedByEmployeeID,
		CreatedAt:                conv.TimeFromPgTimestamptz(item.CreatedAt),
		UpdatedAt:                conv.TimeFromPgTimestamptz(item.UpdatedAt),
	}
}

func toDomainEmployeeTrainingAssignment(item db.EmployeeTrainingAssignment) *domain.EmployeeTrainingAssignment {
	return &domain.EmployeeTrainingAssignment{
		ID:                   item.ID,
		EmployeeID:           item.EmployeeID,
		TrainingID:           item.TrainingID,
		AssignedByEmployeeID: item.AssignedByEmployeeID,
		Status:               string(item.Status),
		AssignedAt:           conv.TimeFromPgTimestamptz(item.AssignedAt),
		DueAt:                conv.TimeFromPgTimestamptz(item.DueAt),
		StartedAt:            timePtrFromPgTimestamptz(item.StartedAt),
		CompletedAt:          timePtrFromPgTimestamptz(item.CompletedAt),
		CancelledAt:          timePtrFromPgTimestamptz(item.CancelledAt),
		CancellationReason:   item.CancellationReason,
		CompletionNotes:      item.CompletionNotes,
		CreatedAt:            conv.TimeFromPgTimestamptz(item.CreatedAt),
		UpdatedAt:            conv.TimeFromPgTimestamptz(item.UpdatedAt),
	}
}

func isTrainingAssignmentUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		strings.Contains(pgErr.ConstraintName, "uq_employee_training_one_non_cancelled")
}

var _ domain.TrainingRepository = (*TrainingRepository)(nil)
