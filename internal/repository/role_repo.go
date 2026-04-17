package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"

	"github.com/google/uuid"
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

func (r *RoleRepository) ListRolePermissions(
	ctx context.Context,
	roleID uuid.UUID,
) ([]domain.RolePermission, error) {
	if _, err := r.queries.GetRoleByID(ctx, roleID); err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}

	rows, err := r.queries.ListRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.RolePermission, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.RolePermission{
			PermissionID:       row.PermissionID,
			PermissionName:     row.PermissionName,
			PermissionResource: row.PermissionResource,
			PermissionMethod:   row.PermissionMethod,
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
