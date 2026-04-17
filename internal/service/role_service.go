package service

import (
	"context"
	"strings"

	"hrbackend/internal/domain"

	"go.uber.org/zap"
)

type RoleService struct {
	repository domain.RoleRepository
	logger     domain.Logger
}

func NewRoleService(repository domain.RoleRepository, logger domain.Logger) domain.RoleService {
	return &RoleService{
		repository: repository,
		logger:     logger,
	}
}

func (s *RoleService) ListRoles(ctx context.Context) ([]domain.RoleSummary, error) {
	items, err := s.repository.ListRoles(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"RoleService.ListRoles",
				"failed to list roles",
				err,
				zap.Int("result_count", len(items)),
			)
		}
		return nil, err
	}

	return items, nil
}

func (s *RoleService) ListAllPermissions(ctx context.Context) ([]domain.PermissionCatalogGroup, error) {
	permissions, err := s.repository.ListAllPermissions(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"RoleService.ListAllPermissions",
				"failed to list permissions",
				err,
			)
		}
		return nil, err
	}

	groupOrder := make([]string, 0, 16)
	grouped := make(map[string]*domain.PermissionCatalogGroup, 16)

	for _, permission := range permissions {
		groupKey := permission.GroupKey
		group, exists := grouped[groupKey]
		if !exists {
			groupOrder = append(groupOrder, groupKey)
			group = &domain.PermissionCatalogGroup{
				GroupKey:   groupKey,
				GroupLabel: humanizePermissionKey(groupKey),
				Sections:   make([]domain.PermissionCatalogSection, 0, 4),
			}
			grouped[groupKey] = group
		}

		sectionKey := permission.SectionKey
		sectionIndex := -1
		for i := range group.Sections {
			if group.Sections[i].SectionKey == sectionKey {
				sectionIndex = i
				break
			}
		}

		if sectionIndex == -1 {
			group.Sections = append(group.Sections, domain.PermissionCatalogSection{
				SectionKey:   sectionKey,
				SectionLabel: humanizePermissionKey(sectionKey),
				Permissions:  make([]domain.PermissionCatalogItem, 0, 8),
			})
			sectionIndex = len(group.Sections) - 1
		}

		group.Sections[sectionIndex].Permissions = append(
			group.Sections[sectionIndex].Permissions,
			permission,
		)
	}

	response := make([]domain.PermissionCatalogGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		response = append(response, *grouped[key])
	}

	return response, nil
}

func humanizePermissionKey(key string) string {
	normalized := strings.TrimSpace(strings.ToLower(key))
	if normalized == "" {
		return ""
	}

	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '.' || r == '_' || r == '-'
	})
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}

	return strings.Join(parts, " ")
}

var _ domain.RoleService = (*RoleService)(nil)
