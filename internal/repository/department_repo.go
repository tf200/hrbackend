package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
)

type DepartmentRepository struct {
	queries db.Querier
}

func NewDepartmentRepository(queries db.Querier) domain.DepartmentRepository {
	return &DepartmentRepository{queries: queries}
}

func (r *DepartmentRepository) CreateDepartment(
	ctx context.Context,
	params domain.CreateDepartmentParams,
) (*domain.Department, error) {
	department, err := r.queries.CreateDepartment(ctx, db.CreateDepartmentParams{
		Name:                     params.Name,
		Description:              params.Description,
		DepartmentHeadEmployeeID: params.DepartmentHeadEmployeeID,
	})
	if err != nil {
		return nil, err
	}

	return toDomainDepartment(department), nil
}

func (r *DepartmentRepository) GetDepartmentByID(
	ctx context.Context,
	departmentID uuid.UUID,
) (*domain.Department, error) {
	department, err := r.queries.GetDepartment(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	return toDomainDepartment(department), nil
}

func (r *DepartmentRepository) UpdateDepartment(
	ctx context.Context,
	departmentID uuid.UUID,
	params domain.UpdateDepartmentParams,
) (*domain.Department, error) {
	department, err := r.queries.UpdateDepartment(ctx, db.UpdateDepartmentParams{
		ID:                       departmentID,
		Name:                     params.Name,
		Description:              params.Description,
		DepartmentHeadEmployeeID: params.DepartmentHeadEmployeeID,
	})
	if err != nil {
		return nil, err
	}

	return toDomainDepartment(department), nil
}

func (r *DepartmentRepository) DeleteDepartment(ctx context.Context, departmentID uuid.UUID) error {
	_, err := r.queries.DeleteDepartment(ctx, departmentID)
	return err
}

func (r *DepartmentRepository) ListDepartments(
	ctx context.Context,
	params domain.ListDepartmentsParams,
) (*domain.DepartmentPage, error) {
	rows, err := r.queries.ListDepartmentsPaginated(ctx, db.ListDepartmentsPaginatedParams{
		Limit:   params.Limit,
		Offset:  params.Offset,
		Column3: params.Search,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.DepartmentPage{
		Items: make([]domain.Department, 0, len(rows)),
	}

	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, domain.Department{
			ID:                       row.ID,
			Name:                     row.Name,
			Description:              row.Description,
			DepartmentHeadEmployeeID: row.DepartmentHeadEmployeeID,
			CreatedAt:                conv.TimeFromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                conv.TimeFromPgTimestamptz(row.UpdatedAt),
		})
	}

	return page, nil
}

func toDomainDepartment(d db.Department) *domain.Department {
	return &domain.Department{
		ID:                       d.ID,
		Name:                     d.Name,
		Description:              d.Description,
		DepartmentHeadEmployeeID: d.DepartmentHeadEmployeeID,
		CreatedAt:                conv.TimeFromPgTimestamptz(d.CreatedAt),
		UpdatedAt:                conv.TimeFromPgTimestamptz(d.UpdatedAt),
	}
}

var _ domain.DepartmentRepository = (*DepartmentRepository)(nil)
