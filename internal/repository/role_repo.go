package repository

import (
	"context"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"

	"github.com/google/uuid"
)

type RoleRepository struct {
	store *db.Store
}

func NewRoleRepository(store *db.Store) domain.RoleRepository {
	return &RoleRepository{store: store}
}

func (r *RoleRepository) ListRoles(ctx context.Context) ([]domain.RoleSummary, error) {
	rows, err := r.store.ListRoles(ctx)
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
	rows, err := r.store.ListAllPermissions(ctx)
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
	if _, err := r.store.GetRoleByID(ctx, roleID); err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}

	rows, err := r.store.ListRolePermissions(ctx, roleID)
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

func (r *RoleRepository) UpdateRolePermissions(
	ctx context.Context,
	roleID uuid.UUID,
	permissionIDs []uuid.UUID,
) error {
	// First check if role exists
	if _, err := r.store.GetRoleByID(ctx, roleID); err != nil {
		if isDBNotFound(err) {
			return domain.ErrRoleNotFound
		}
		return err
	}

	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		// 1. Delete all old permissions for this role (Bulk Delete)
		err := q.RemovePermissionsFromRole(ctx, roleID)
		if err != nil {
			return err
		}

		// 2. Insert new permissions if any (Bulk Insert)
		if len(permissionIDs) > 0 {
			err = q.AddPermissionsToRole(ctx, db.AddPermissionsToRoleParams{
				RoleID:        roleID,
				PermissionIds: permissionIDs,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

var _ domain.RoleRepository = (*RoleRepository)(nil)
