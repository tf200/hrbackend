package service

import (
	"context"
	"errors"
	"testing"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestRoleServiceListAllPermissionsGroupsResults(t *testing.T) {
	repo := &fakeRoleRepository{
		permissions: []domain.PermissionCatalogItem{
			{
				PermissionID:       uuid.New(),
				PermissionName:     "EMPLOYEE.VIEW",
				PermissionResource: "EMPLOYEE",
				GroupKey:           "employee",
				SectionKey:         "view",
				DisplayName:        "Employee View",
				SortOrder:          10,
			},
			{
				PermissionID:       uuid.New(),
				PermissionName:     "EMPLOYEE.UPDATE",
				PermissionResource: "EMPLOYEE",
				GroupKey:           "employee",
				SectionKey:         "update",
				DisplayName:        "Employee Update",
				SortOrder:          20,
			},
			{
				PermissionID:       uuid.New(),
				PermissionName:     "EMPLOYEE.VIEW_ALL",
				PermissionResource: "EMPLOYEE",
				GroupKey:           "employee",
				SectionKey:         "view",
				DisplayName:        "Employee View All",
				SortOrder:          30,
			},
			{
				PermissionID:       uuid.New(),
				PermissionName:     "ROLE.VIEW",
				PermissionResource: "ROLE",
				GroupKey:           "role_management",
				SectionKey:         "view",
				DisplayName:        "Role View",
				SortOrder:          40,
			},
		},
	}
	service := &RoleService{repository: repo}

	groups, err := service.ListAllPermissions(context.Background())
	if err != nil {
		t.Fatalf("ListAllPermissions returned error: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	if groups[0].GroupKey != "employee" || groups[0].GroupLabel != "Employee" {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	if len(groups[0].Sections) != 2 {
		t.Fatalf("expected employee group to have 2 sections, got %d", len(groups[0].Sections))
	}
	if groups[0].Sections[0].SectionKey != "view" || groups[0].Sections[0].SectionLabel != "View" {
		t.Fatalf("unexpected first section: %+v", groups[0].Sections[0])
	}
	if len(groups[0].Sections[0].Permissions) != 2 {
		t.Fatalf("expected view section to have 2 permissions, got %d", len(groups[0].Sections[0].Permissions))
	}
	if groups[0].Sections[0].Permissions[0].PermissionName != "EMPLOYEE.VIEW" {
		t.Fatalf("expected first permission to preserve source order, got %s", groups[0].Sections[0].Permissions[0].PermissionName)
	}
	if groups[1].GroupKey != "role_management" || groups[1].GroupLabel != "Role Management" {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
}

func TestRoleServiceListAllPermissionsReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("boom")
	repo := &fakeRoleRepository{permissionsErr: expectedErr}
	service := &RoleService{repository: repo}

	_, err := service.ListAllPermissions(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestRoleServiceListAllPermissionsHandlesEmptyResults(t *testing.T) {
	service := &RoleService{repository: &fakeRoleRepository{}}

	groups, err := service.ListAllPermissions(context.Background())
	if err != nil {
		t.Fatalf("ListAllPermissions returned error: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected no groups, got %d", len(groups))
	}
}

func TestRoleServiceListRolePermissions(t *testing.T) {
	expected := []domain.RolePermission{
		{
			PermissionID:       uuid.New(),
			PermissionName:     "ROLE.VIEW",
			PermissionResource: "ROLE",
			PermissionMethod:   "VIEW",
			GroupKey:           "role",
			SectionKey:         "view",
			DisplayName:        "Role View",
			SortOrder:          10,
		},
	}
	repo := &fakeRoleRepository{rolePermissions: expected}
	service := &RoleService{repository: repo}

	items, err := service.ListRolePermissions(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListRolePermissions returned error: %v", err)
	}
	if len(items) != len(expected) {
		t.Fatalf("expected %d permissions, got %d", len(expected), len(items))
	}
	if items[0].PermissionMethod != "VIEW" {
		t.Fatalf("expected permission method VIEW, got %s", items[0].PermissionMethod)
	}
}

func TestRoleServiceListRolePermissionsReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("boom")
	repo := &fakeRoleRepository{rolePermissionsErr: expectedErr}
	service := &RoleService{repository: repo}

	_, err := service.ListRolePermissions(context.Background(), uuid.New())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestRoleServiceListRolePermissionsReturnsRoleNotFound(t *testing.T) {
	repo := &fakeRoleRepository{rolePermissionsErr: domain.ErrRoleNotFound}
	service := &RoleService{repository: repo}

	_, err := service.ListRolePermissions(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestRoleServiceListRolePermissionsHandlesEmptyResults(t *testing.T) {
	service := &RoleService{repository: &fakeRoleRepository{}}

	items, err := service.ListRolePermissions(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListRolePermissions returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no permissions, got %d", len(items))
	}
}

func TestHumanizePermissionKey(t *testing.T) {
	tests := map[string]string{
		"employee":        "Employee",
		"role_management": "Role Management",
		"view-all":        "View All",
		"":                "",
	}

	for input, expected := range tests {
		if actual := humanizePermissionKey(input); actual != expected {
			t.Fatalf("humanizePermissionKey(%q) = %q, want %q", input, actual, expected)
		}
	}
}

type fakeRoleRepository struct {
	roles              []domain.RoleSummary
	rolesErr           error
	permissions        []domain.PermissionCatalogItem
	permissionsErr     error
	rolePermissions    []domain.RolePermission
	rolePermissionsErr error
}

func (f *fakeRoleRepository) ListRoles(_ context.Context) ([]domain.RoleSummary, error) {
	if f.rolesErr != nil {
		return nil, f.rolesErr
	}

	return f.roles, nil
}

func (f *fakeRoleRepository) ListAllPermissions(
	_ context.Context,
) ([]domain.PermissionCatalogItem, error) {
	if f.permissionsErr != nil {
		return nil, f.permissionsErr
	}

	return f.permissions, nil
}

func (f *fakeRoleRepository) ListRolePermissions(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.RolePermission, error) {
	if f.rolePermissionsErr != nil {
		return nil, f.rolePermissionsErr
	}

	return f.rolePermissions, nil
}
