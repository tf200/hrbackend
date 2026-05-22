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
	profile        *domain.EmployeeProfile
	employeeDetail *domain.EmployeeDetail
	err            error
}

func (f *fakeEmployeeRepo) GetEmployeeByUserID(_ context.Context, _ uuid.UUID) (*domain.EmployeeProfile, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.profile, nil
}

func (f *fakeEmployeeRepo) GetEmployeeByID(_ context.Context, _ uuid.UUID) (*domain.EmployeeDetail, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.employeeDetail, nil
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
func (f *fakeEmployeeRepo) ListQualifications(_ context.Context, _ uuid.UUID) ([]domain.Qualification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) AddQualification(_ context.Context, _ uuid.UUID, _ domain.CreateQualificationParams) (*domain.Qualification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) UpdateQualification(_ context.Context, _ uuid.UUID, _ domain.UpdateQualificationParams) (*domain.Qualification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) DeleteQualification(_ context.Context, _ uuid.UUID) (*domain.Qualification, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) ListQualificationTypes(_ context.Context) ([]domain.QualificationType, error) {
	return nil, f.err
}
func (f *fakeEmployeeRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return f.err
}

// --- Fake task queue ---

type fakeTaskQueue struct {
	enqueuedPayload domain.EmailDeliveryTaskPayload
	enqueued        bool
	err             error
}

func (f *fakeTaskQueue) EnqueueEmailDelivery(_ context.Context, payload domain.EmailDeliveryTaskPayload, _ *domain.TaskEnqueueOptions) error {
	if f.err != nil {
		return f.err
	}
	f.enqueued = true
	f.enqueuedPayload = payload
	return nil
}
func (f *fakeTaskQueue) EnqueueIncident(_ context.Context, _ domain.IncidentTaskPayload, _ *domain.TaskEnqueueOptions) error { return nil }
func (f *fakeTaskQueue) EnqueueIncidentConfirmedEmail(_ context.Context, _ domain.IncidentConfirmedEmailTaskPayload, _ *domain.TaskEnqueueOptions) error { return nil }
func (f *fakeTaskQueue) EnqueueNotificationTask(_ context.Context, _ domain.NotificationTaskPayload, _ *domain.TaskEnqueueOptions) error { return nil }
func (f *fakeTaskQueue) Close() error { return nil }

func employeeDetailWithWorkEmail(email string) *domain.EmployeeDetail {
	return &domain.EmployeeDetail{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		FirstName:        "John",
		LastName:         "Doe",
		WorkEmailAddress: &email,
	}
}

func TestResetPassword_manualPassword(t *testing.T) {
	email := "john@example.com"
	password := "MySecurePass123"
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail(email)}
	svc := &EmployeeService{repo: repo}

	result, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: false,
		Password:  &password,
		SendEmail: false,
	})
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if result.TemporaryPassword != password {
		t.Fatalf("expected temporary_password %q, got %q", password, result.TemporaryPassword)
	}
}

func TestResetPassword_generatedPassword(t *testing.T) {
	email := "john@example.com"
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail(email)}
	svc := &EmployeeService{repo: repo}

	result, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: true,
		SendEmail: false,
	})
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if len(result.TemporaryPassword) == 0 {
		t.Fatal("expected non-empty generated password")
	}
}

func TestResetPassword_missingPasswordReturnsError(t *testing.T) {
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail("john@example.com")}
	svc := &EmployeeService{repo: repo}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: false,
		Password:  nil,
		SendEmail: false,
	})
	if !errors.Is(err, domain.ErrInvalidPasswordResetRequest) {
		t.Fatalf("expected ErrInvalidPasswordResetRequest, got %v", err)
	}
}

func TestResetPassword_whitespacePasswordReturnsError(t *testing.T) {
	whitespace := "   "
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail("john@example.com")}
	svc := &EmployeeService{repo: repo}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: false,
		Password:  &whitespace,
		SendEmail: false,
	})
	if !errors.Is(err, domain.ErrInvalidPasswordResetRequest) {
		t.Fatalf("expected ErrInvalidPasswordResetRequest, got %v", err)
	}
}

func TestResetPassword_sendEmailWithoutTaskQueueReturnsError(t *testing.T) {
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail("john@example.com")}
	svc := &EmployeeService{repo: repo, taskQueue: nil}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: true,
		SendEmail: true,
	})
	if !errors.Is(err, domain.ErrEmailDeliveryFailed) {
		t.Fatalf("expected ErrEmailDeliveryFailed, got %v", err)
	}
}

func TestResetPassword_sendEmailWithoutWorkEmailReturnsError(t *testing.T) {
	detail := employeeDetailWithWorkEmail("")
	detail.WorkEmailAddress = nil
	repo := &fakeEmployeeRepo{employeeDetail: detail}
	tq := &fakeTaskQueue{}
	svc := &EmployeeService{repo: repo, taskQueue: tq}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: true,
		SendEmail: true,
	})
	if !errors.Is(err, domain.ErrEmailDeliveryFailed) {
		t.Fatalf("expected ErrEmailDeliveryFailed, got %v", err)
	}
	if tq.enqueued {
		t.Fatal("expected email NOT to be enqueued")
	}
}

func TestResetPassword_sendEmailEnqueuesPayload(t *testing.T) {
	email := "john@example.com"
	password := "MySecurePass123"
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail(email)}
	tq := &fakeTaskQueue{}
	svc := &EmployeeService{repo: repo, taskQueue: tq}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: false,
		Password:  &password,
		SendEmail: true,
	})
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if !tq.enqueued {
		t.Fatal("expected email to be enqueued")
	}
	if tq.enqueuedPayload.To != email {
		t.Fatalf("expected To %q, got %q", email, tq.enqueuedPayload.To)
	}
	if tq.enqueuedPayload.UserEmail != email {
		t.Fatalf("expected UserEmail %q, got %q", email, tq.enqueuedPayload.UserEmail)
	}
	if tq.enqueuedPayload.UserPassword != password {
		t.Fatalf("expected UserPassword %q, got %q", password, tq.enqueuedPayload.UserPassword)
	}
}

func TestResetPassword_enqueueErrorReturnsError(t *testing.T) {
	email := "john@example.com"
	password := "MySecurePass123"
	repo := &fakeEmployeeRepo{employeeDetail: employeeDetailWithWorkEmail(email)}
	tq := &fakeTaskQueue{err: errors.New("redis down")}
	svc := &EmployeeService{repo: repo, taskQueue: tq}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: false,
		Password:  &password,
		SendEmail: true,
	})
	if !errors.Is(err, domain.ErrEmailDeliveryFailed) {
		t.Fatalf("expected ErrEmailDeliveryFailed, got %v", err)
	}
}

func TestResetPassword_employeeNotFoundError(t *testing.T) {
	repo := &fakeEmployeeRepo{err: domain.ErrEmployeeNotFound}
	svc := &EmployeeService{repo: repo}

	_, err := svc.ResetPassword(context.Background(), uuid.New(), domain.ResetPasswordParams{
		Generated: true,
		SendEmail: false,
	})
	if !errors.Is(err, domain.ErrEmployeeNotFound) {
		t.Fatalf("expected ErrEmployeeNotFound, got %v", err)
	}
}
