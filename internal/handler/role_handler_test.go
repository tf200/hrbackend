package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hrbackend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRoleHandlerListRolePermissionsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewRoleHandler(&fakeRoleService{})
	router.GET("/roles/:id/permissions", handler.ListRolePermissions)

	req := httptest.NewRequest(http.MethodGet, "/roles/not-a-uuid/permissions", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestRoleHandlerListRolePermissionsRoleNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewRoleHandler(&fakeRoleService{rolePermissionsErr: domain.ErrRoleNotFound})
	router.GET("/roles/:id/permissions", handler.ListRolePermissions)

	req := httptest.NewRequest(
		http.MethodGet,
		"/roles/"+uuid.New().String()+"/permissions",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestRoleHandlerListRolePermissionsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	description := "Can view roles"
	expected := []domain.RolePermission{
		{
			PermissionID:       uuid.New(),
			PermissionName:     "ROLE.VIEW",
			PermissionResource: "ROLE",
			PermissionMethod:   "VIEW",
			GroupKey:           "role",
			SectionKey:         "view",
			DisplayName:        "Role View",
			Description:        &description,
			SortOrder:          10,
		},
	}

	router := gin.New()
	handler := NewRoleHandler(&fakeRoleService{rolePermissions: expected})
	router.GET("/roles/:id/permissions", handler.ListRolePermissions)

	req := httptest.NewRequest(
		http.MethodGet,
		"/roles/"+uuid.New().String()+"/permissions",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Success bool                     `json:"success"`
		Message string                   `json:"message"`
		Data    []rolePermissionResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !response.Success {
		t.Fatalf("expected success response")
	}
	if response.Message != "Role permissions retrieved successfully" {
		t.Fatalf("unexpected message: %s", response.Message)
	}
	if len(response.Data) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(response.Data))
	}
	if response.Data[0].PermissionMethod != "VIEW" {
		t.Fatalf("expected permission method VIEW, got %s", response.Data[0].PermissionMethod)
	}
}

func TestRoleHandlerListRolePermissionsEmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewRoleHandler(&fakeRoleService{})
	router.GET("/roles/:id/permissions", handler.ListRolePermissions)

	req := httptest.NewRequest(
		http.MethodGet,
		"/roles/"+uuid.New().String()+"/permissions",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Data []rolePermissionResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(response.Data) != 0 {
		t.Fatalf("expected empty permissions, got %d", len(response.Data))
	}
}

type fakeRoleService struct {
	roles              []domain.RoleSummary
	rolesErr           error
	allPermissions     []domain.PermissionCatalogGroup
	allPermissionsErr  error
	rolePermissions    []domain.RolePermission
	rolePermissionsErr error
}

func (f *fakeRoleService) ListRoles(_ context.Context) ([]domain.RoleSummary, error) {
	if f.rolesErr != nil {
		return nil, f.rolesErr
	}
	return f.roles, nil
}

func (f *fakeRoleService) ListAllPermissions(
	_ context.Context,
) ([]domain.PermissionCatalogGroup, error) {
	if f.allPermissionsErr != nil {
		return nil, f.allPermissionsErr
	}
	return f.allPermissions, nil
}

func (f *fakeRoleService) ListRolePermissions(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.RolePermission, error) {
	if f.rolePermissionsErr != nil {
		return nil, f.rolePermissionsErr
	}
	return f.rolePermissions, nil
}

var _ domain.RoleService = (*fakeRoleService)(nil)
