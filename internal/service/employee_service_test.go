package service

import (
	"context"
	"errors"
	"testing"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestComputePortalAccess_adminOnly(t *testing.T) {
	perms := []domain.Permission{
		{Name: domain.PortalPermissionAdmin},
	}
	got := computePortalAccess(perms)
	if got != domain.PortalAccessAdmin {
		t.Fatalf("expected %q, got %q", domain.PortalAccessAdmin, got)
	}
}

func TestComputePortalAccess_employeeOnly(t *testing.T) {
	perms := []domain.Permission{
		{Name: domain.PortalPermissionEmployee},
	}
	got := computePortalAccess(perms)
	if got != domain.PortalAccessEmployee {
		t.Fatalf("expected %q, got %q", domain.PortalAccessEmployee, got)
	}
}

func TestComputePortalAccess_both(t *testing.T) {
	perms := []domain.Permission{
		{Name: domain.PortalPermissionAdmin},
		{Name: domain.PortalPermissionEmployee},
	}
	got := computePortalAccess(perms)
	if got != domain.PortalAccessBoth {
		t.Fatalf("expected %q, got %q", domain.PortalAccessBoth, got)
	}
}

func TestComputePortalAccess_neitherDefaultsToEmployee(t *testing.T) {
	perms := []domain.Permission{
		{Name: "SOME.OTHER.PERMISSION"},
	}
	got := computePortalAccess(perms)
	if got != domain.PortalAccessEmployee {
		t.Fatalf("expected default %q, got %q", domain.PortalAccessEmployee, got)
	}
}

func TestComputePortalAccess_emptyPermissionsDefaultsToEmployee(t *testing.T) {
	got := computePortalAccess(nil)
	if got != domain.PortalAccessEmployee {
		t.Fatalf("expected default %q, got %q", domain.PortalAccessEmployee, got)
	}
}

func TestComputePortalAccess_adminAmongOtherPermissions(t *testing.T) {
	perms := []domain.Permission{
		{Name: "EMPLOYEE.VIEW"},
		{Name: "LEAVE.REQUEST.CREATE"},
		{Name: domain.PortalPermissionAdmin},
		{Name: "SCHEDULE.VIEW"},
	}
	got := computePortalAccess(perms)
	if got != domain.PortalAccessAdmin {
		t.Fatalf("expected %q, got %q", domain.PortalAccessAdmin, got)
	}
}

// --- Service-level test for GetEmployeeProfile ---

func TestEmployeeServiceGetEmployeeProfile_setsPortalAccess(t *testing.T) {
	userID := uuid.New()
	perms := []domain.Permission{
		{Name: domain.PortalPermissionAdmin},
	}

	repo := &fakeEmployeeRepo{
		profile: &domain.EmployeeProfile{
			UserID:      userID,
			Permissions: perms,
		},
	}
	svc := &EmployeeService{repo: repo}

	profile, err := svc.GetEmployeeProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetEmployeeProfile returned error: %v", err)
	}
	if profile.PortalAccess != domain.PortalAccessAdmin {
		t.Fatalf("expected portal_access %q, got %q", domain.PortalAccessAdmin, profile.PortalAccess)
	}
}

func TestEmployeeServiceGetEmployeeProfile_repoError(t *testing.T) {
	expectedErr := errors.New("db error")
	repo := &fakeEmployeeRepo{err: expectedErr}
	svc := &EmployeeService{repo: repo}

	_, err := svc.GetEmployeeProfile(context.Background(), uuid.New())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

// --- Fake repository ---

type fakeEmployeeRepo struct {
	profile *domain.EmployeeProfile
	err     error
}

func (f *fakeEmployeeRepo) GetEmployeeByUserID(_ context.Context, _ uuid.UUID) (*domain.EmployeeProfile, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.profile, nil
}

func (f *fakeEmployeeRepo) GetEmployeeByID(_ context.Context, _ uuid.UUID) (*domain.EmployeeDetail, error) {
	return nil, nil
}
func (f *fakeEmployeeRepo) ListEmployees(_ context.Context, _ domain.ListEmployeesParams) (*domain.EmployeePage, error) {
	return nil, nil
}
func (f *fakeEmployeeRepo) CountEmployees(_ context.Context, _ domain.ListEmployeesParams) (int64, error) {
	return 0, nil
}
func (f *fakeEmployeeRepo) CreateEmployee(_ context.Context, _ domain.CreateEmployeeParams) (*domain.EmployeeDetail, error) {
	return nil, nil
}
func (f *fakeEmployeeRepo) UpdateEmployee(_ context.Context, _ uuid.UUID, _ domain.UpdateEmployeeParams) (*domain.EmployeeDetail, error) {
	return nil, nil
}
func (f *fakeEmployeeRepo) GetEmployeeCounts(_ context.Context) (*domain.EmployeeCounts, error) {
	return nil, f.err
}

func (f *fakeEmployeeRepo) SearchEmployeesByNameOrEmail(_ context.Context, _ *string) ([]domain.EmployeeSearchResult, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) GetContractDetails(_ context.Context, _ uuid.UUID) (*domain.ContractDetails, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) AddContractDetails(_ context.Context, _ uuid.UUID, _ domain.AddContractDetailsParams) (*domain.EmployeeDetail, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) UpdateIsSubcontractor(_ context.Context, _ uuid.UUID, _ string) (*domain.EmployeeDetail, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) ListContractChanges(_ context.Context, _ uuid.UUID) ([]domain.EmployeeContractChange, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) CreateContractChange(_ context.Context, _, _ uuid.UUID, _ domain.CreateEmployeeContractChangeParams) (*domain.CreateEmployeeContractChangeResult, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) ListEducation(_ context.Context, _ uuid.UUID) ([]domain.Education, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) AddEducation(_ context.Context, _ uuid.UUID, _ domain.CreateEducationParams) (*domain.Education, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) UpdateEducation(_ context.Context, _ uuid.UUID, _ domain.UpdateEducationParams) (*domain.Education, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) DeleteEducation(_ context.Context, _ uuid.UUID) (*domain.Education, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) ListExperience(_ context.Context, _ uuid.UUID) ([]domain.Experience, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) AddExperience(_ context.Context, _ uuid.UUID, _ domain.CreateExperienceParams) (*domain.Experience, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) UpdateExperience(_ context.Context, _ uuid.UUID, _ domain.UpdateExperienceParams) (*domain.Experience, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) DeleteExperience(_ context.Context, _ uuid.UUID) (*domain.Experience, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) ListCertification(_ context.Context, _ uuid.UUID) ([]domain.Certification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) AddCertification(_ context.Context, _ uuid.UUID, _ domain.CreateCertificationParams) (*domain.Certification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) UpdateCertification(_ context.Context, _ uuid.UUID, _ domain.UpdateCertificationParams) (*domain.Certification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) DeleteCertification(_ context.Context, _ uuid.UUID) (*domain.Certification, error) {
	return nil, f.err
}
