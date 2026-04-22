package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTrainingInvalidRequest           = errors.New("invalid training request")
	ErrTrainingNotFound                 = errors.New("training not found")
	ErrTrainingAssignmentConflict       = errors.New("training assignment already exists")
	ErrTrainingAssignmentNotFound       = errors.New("training assignment not found")
	ErrTrainingAssignmentNotCancellable = errors.New("training assignment is not cancellable")
)

type TrainingCatalogItem struct {
	ID                       uuid.UUID
	Title                    string
	Description              *string
	Category                 *string
	EstimatedDurationMinutes *int32
	IsActive                 bool
	CreatedByEmployeeID      *uuid.UUID
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CreateTrainingCatalogItemParams struct {
	Title                    string
	Description              *string
	Category                 *string
	EstimatedDurationMinutes *int32
	CreatedByEmployeeID      *uuid.UUID
}

type AssignTrainingToEmployeeParams struct {
	EmployeeID           uuid.UUID
	TrainingID           uuid.UUID
	AssignedByEmployeeID *uuid.UUID
	DueAt                time.Time
}

type CancelTrainingAssignmentParams struct {
	AssignmentID       uuid.UUID
	CancellationReason *string
}

type EmployeeTrainingAssignment struct {
	ID                   uuid.UUID
	EmployeeID           uuid.UUID
	TrainingID           uuid.UUID
	AssignedByEmployeeID *uuid.UUID
	Status               string
	AssignedAt           time.Time
	DueAt                time.Time
	StartedAt            *time.Time
	CompletedAt          *time.Time
	CancelledAt          *time.Time
	CancellationReason   *string
	CompletionNotes      *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ListTrainingAssignmentsParams struct {
	Limit          int32
	Offset         int32
	EmployeeSearch *string
	DepartmentID   *uuid.UUID
	TrainingID     *uuid.UUID
	Status         *string
}

type ListTrainingCatalogItemsParams struct {
	Limit    int32
	Offset   int32
	Search   *string
	IsActive *bool
}

type TrainingAssignmentListItem struct {
	AssignmentID         uuid.UUID
	EmployeeID           uuid.UUID
	EmployeeNumber       *string
	EmploymentNumber     *string
	FirstName            string
	LastName             string
	DepartmentID         *uuid.UUID
	DepartmentName       *string
	TrainingID           uuid.UUID
	TrainingTitle        string
	TrainingCategory     *string
	Status               string
	AssignedAt           time.Time
	DueAt                *time.Time
	StartedAt            *time.Time
	CompletedAt          *time.Time
	AssignedByEmployeeID *uuid.UUID
	AssignedByName       *string
	IsOverdue            bool
}

type TrainingCatalogItemPage struct {
	Items      []TrainingCatalogItem
	TotalCount int64
}

type TrainingAssignmentPage struct {
	Items      []TrainingAssignmentListItem
	TotalCount int64
}

type TrainingRepository interface {
	AssignTrainingToEmployee(
		ctx context.Context,
		params AssignTrainingToEmployeeParams,
	) (*EmployeeTrainingAssignment, error)
	CancelTrainingAssignment(
		ctx context.Context,
		params CancelTrainingAssignmentParams,
	) (*EmployeeTrainingAssignment, error)
	ListTrainingAssignments(
		ctx context.Context,
		params ListTrainingAssignmentsParams,
	) (*TrainingAssignmentPage, error)
	CreateTrainingCatalogItem(
		ctx context.Context,
		params CreateTrainingCatalogItemParams,
	) (*TrainingCatalogItem, error)
	ListTrainingCatalogItems(
		ctx context.Context,
		params ListTrainingCatalogItemsParams,
	) (*TrainingCatalogItemPage, error)
}

type TrainingService interface {
	AssignTrainingToEmployee(
		ctx context.Context,
		params AssignTrainingToEmployeeParams,
	) (*EmployeeTrainingAssignment, error)
	CancelTrainingAssignment(
		ctx context.Context,
		params CancelTrainingAssignmentParams,
	) (*EmployeeTrainingAssignment, error)
	ListTrainingAssignments(
		ctx context.Context,
		params ListTrainingAssignmentsParams,
	) (*TrainingAssignmentPage, error)
	CreateTrainingCatalogItem(
		ctx context.Context,
		params CreateTrainingCatalogItemParams,
	) (*TrainingCatalogItem, error)
	ListTrainingCatalogItems(
		ctx context.Context,
		params ListTrainingCatalogItemsParams,
	) (*TrainingCatalogItemPage, error)
}
