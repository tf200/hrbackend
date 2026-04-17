package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
)

type RoleRepository struct {
	queries db.Querier
}

func NewRoleRepository(queries db.Querier) domain.RoleRepository {
	return &RoleRepository{queries: queries}
}

func (r *RoleRepository) ListRoles(ctx context.Context) ([]domain.RoleSummary, error) {
	rows, err := r.queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.RoleSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.RoleSummary{
			ID:              row.ID,
			Name:            row.Name,
			Description:     row.Description,
			PermissionCount: row.PermissionCount,
			EmployeeCount:   row.EmployeeCount,
		})
	}

	return items, nil
}

func (r *RoleRepository) ListAllPermissions(ctx context.Context) ([]domain.PermissionCatalogItem, error) {
	rows, err := r.queries.ListAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PermissionCatalogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.PermissionCatalogItem{
			PermissionID:       row.ID,
			PermissionName:     row.Name,
			PermissionResource: row.Resource,
			GroupKey:           row.GroupKey,
			SectionKey:         row.SectionKey,
			DisplayName:        row.DisplayName,
			Description:        row.Description,
			SortOrder:          row.SortOrder,
		})
	}

	return items, nil
}

var _ domain.RoleRepository = (*RoleRepository)(nil)
