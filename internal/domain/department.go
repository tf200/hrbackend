package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID                       uuid.UUID
	Name                     string
	Description              *string
	DepartmentHeadEmployeeID *uuid.UUID
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CreateDepartmentParams struct {
	Name                     string
	Description              *string
	DepartmentHeadEmployeeID *uuid.UUID
}

type UpdateDepartmentParams struct {
	Name                     *string
	Description              *string
	DepartmentHeadEmployeeID *uuid.UUID
}

type ListDepartmentsParams struct {
	Limit  int32
	Offset int32
	Search string
}

type DepartmentPage struct {
	Items      []Department
	TotalCount int64
}

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, params CreateDepartmentParams) (*Department, error)
	GetDepartmentByID(ctx context.Context, departmentID uuid.UUID) (*Department, error)
	UpdateDepartment(
		ctx context.Context,
		departmentID uuid.UUID,
		params UpdateDepartmentParams,
	) (*Department, error)
	DeleteDepartment(ctx context.Context, departmentID uuid.UUID) error
	ListDepartments(ctx context.Context, params ListDepartmentsParams) (*DepartmentPage, error)
}

type DepartmentService interface {
	CreateDepartment(ctx context.Context, params CreateDepartmentParams) (*Department, error)
	GetDepartmentByID(ctx context.Context, departmentID uuid.UUID) (*Department, error)
	UpdateDepartment(
		ctx context.Context,
		departmentID uuid.UUID,
		params UpdateDepartmentParams,
	) (*Department, error)
	DeleteDepartment(ctx context.Context, departmentID uuid.UUID) error
	ListDepartments(ctx context.Context, params ListDepartmentsParams) (*DepartmentPage, error)
}
