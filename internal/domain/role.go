package domain

import (
	"context"

	"github.com/google/uuid"
)

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

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]RoleSummary, error)
	ListAllPermissions(ctx context.Context) ([]PermissionCatalogItem, error)
}

type RoleService interface {
	ListRoles(ctx context.Context) ([]RoleSummary, error)
	ListAllPermissions(ctx context.Context) ([]PermissionCatalogGroup, error)
}
