package handler

import (
	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type roleResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description"`
	PermissionCount int64     `json:"permission_count"`
	EmployeeCount   int64     `json:"employee_count"`
}

type permissionCatalogItemResponse struct {
	PermissionID       uuid.UUID `json:"permission_id"`
	PermissionName     string    `json:"permission_name"`
	PermissionResource string    `json:"permission_resource"`
	DisplayName        string    `json:"display_name"`
	Description        *string   `json:"description"`
	SortOrder          int32     `json:"sort_order"`
}

type rolePermissionResponse struct {
	PermissionID       uuid.UUID `json:"permission_id"`
	PermissionName     string    `json:"permission_name"`
	PermissionResource string    `json:"permission_resource"`
	PermissionMethod   string    `json:"permission_method"`
	GroupKey           string    `json:"group_key"`
	SectionKey         string    `json:"section_key"`
	DisplayName        string    `json:"display_name"`
	Description        *string   `json:"description"`
	SortOrder          int32     `json:"sort_order"`
}

type permissionCatalogSectionResponse struct {
	SectionKey   string                          `json:"section_key"`
	SectionLabel string                          `json:"section_label"`
	Permissions  []permissionCatalogItemResponse `json:"permissions"`
}

type permissionCatalogGroupResponse struct {
	GroupKey   string                             `json:"group_key"`
	GroupLabel string                             `json:"group_label"`
	Sections   []permissionCatalogSectionResponse `json:"sections"`
}

func toRoleResponse(item domain.RoleSummary) roleResponse {
	return roleResponse{
		ID:              item.ID,
		Name:            item.Name,
		Description:     item.Description,
		PermissionCount: item.PermissionCount,
		EmployeeCount:   item.EmployeeCount,
	}
}

func toRoleResponses(items []domain.RoleSummary) []roleResponse {
	results := make([]roleResponse, len(items))
	for i, item := range items {
		results[i] = toRoleResponse(item)
	}
	return results
}

func toPermissionCatalogResponses(
	items []domain.PermissionCatalogGroup,
) []permissionCatalogGroupResponse {
	results := make([]permissionCatalogGroupResponse, len(items))
	for i, item := range items {
		sections := make([]permissionCatalogSectionResponse, len(item.Sections))
		for sectionIndex, section := range item.Sections {
			permissions := make([]permissionCatalogItemResponse, len(section.Permissions))
			for permissionIndex, permission := range section.Permissions {
				permissions[permissionIndex] = permissionCatalogItemResponse{
					PermissionID:       permission.PermissionID,
					PermissionName:     permission.PermissionName,
					PermissionResource: permission.PermissionResource,
					DisplayName:        permission.DisplayName,
					Description:        permission.Description,
					SortOrder:          permission.SortOrder,
				}
			}

			sections[sectionIndex] = permissionCatalogSectionResponse{
				SectionKey:   section.SectionKey,
				SectionLabel: section.SectionLabel,
				Permissions:  permissions,
			}
		}

		results[i] = permissionCatalogGroupResponse{
			GroupKey:   item.GroupKey,
			GroupLabel: item.GroupLabel,
			Sections:   sections,
		}
	}

	return results
}

func toRolePermissionResponses(items []domain.RolePermission) []rolePermissionResponse {
	results := make([]rolePermissionResponse, len(items))
	for i, item := range items {
		results[i] = rolePermissionResponse{
			PermissionID:       item.PermissionID,
			PermissionName:     item.PermissionName,
			PermissionResource: item.PermissionResource,
			PermissionMethod:   item.PermissionMethod,
			GroupKey:           item.GroupKey,
			SectionKey:         item.SectionKey,
			DisplayName:        item.DisplayName,
			Description:        item.Description,
			SortOrder:          item.SortOrder,
		}
	}

	return results
}
