package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrRoleNotFound = errors.New("role not found")

type RoleSummary struct {
	ID              uuid.UUID
	Name            string
	Description     *string
	PermissionCount int64
	EmployeeCount   int64
}

type PermissionCatalogItem struct {
	PermissionID       uuid.UUID
	PermissionName     string
	PermissionResource string
	GroupKey           string
	SectionKey         string
	DisplayName        string
	Description        *string
	SortOrder          int32
}

type PermissionCatalogSection struct {
	SectionKey   string
	SectionLabel string
	Permissions  []PermissionCatalogItem
}

type PermissionCatalogGroup struct {
	GroupKey   string
	GroupLabel string
	Sections   []PermissionCatalogSection
}

type RolePermission struct {
	PermissionID       uuid.UUID
	PermissionName     string
	PermissionResource string
	PermissionMethod   string
	GroupKey           string
	SectionKey         string
	DisplayName        string
	Description        *string
	SortOrder          int32
}

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]RoleSummary, error)
	ListAllPermissions(ctx context.Context) ([]PermissionCatalogItem, error)
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]RolePermission, error)
}

type RoleService interface {
	ListRoles(ctx context.Context) ([]RoleSummary, error)
	ListAllPermissions(ctx context.Context) ([]PermissionCatalogGroup, error)
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]RolePermission, error)
}
