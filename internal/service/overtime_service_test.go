package service

import (
	"context"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

type fakeNotificationService struct {
	lastRequest domain.NotificationRequest
	callsCount  int
}

func (f *fakeNotificationService) Notify(ctx context.Context, req domain.NotificationRequest) {
	f.lastRequest = req
	f.callsCount++
}

func (f *fakeNotificationService) ListNotifications(
	ctx context.Context,
	userID uuid.UUID,
	page, pageSize int32,
) ([]domain.Notification, int64, error) {
	return nil, 0, nil
}

func (f *fakeNotificationService) GetUnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	return 0, nil
}

func (f *fakeNotificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}

func (f *fakeNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}

type fakeEmployeeRepository struct {
	domain.EmployeeRepository
	employee *domain.EmployeeDetail
}

func (f *fakeEmployeeRepository) GetEmployeeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmployeeDetail, error) {
	if f.employee != nil {
		return f.employee, nil
	}
	return &domain.EmployeeDetail{
		ID:        id,
		FirstName: "Admin",
		LastName:  "User",
	}, nil
}

type fakeOvertimeRepository struct {
	lastParams domain.CreateOvertimeEntryParams
	mockEntry  *domain.OvertimeEntry
}

type fakeOvertimeTxRepository struct {
	repo *fakeOvertimeRepository
}

func (f *fakeOvertimeTxRepository) GetOvertimeEntryForUpdate(
	ctx context.Context,
	id uuid.UUID,
) (*domain.OvertimeEntry, error) {
	return f.repo.mockEntry, nil
}

func (f *fakeOvertimeTxRepository) ApproveOvertimeEntry(
	ctx context.Context,
	id, approvedByEmployeeID uuid.UUID,
) (*domain.OvertimeEntry, error) {
	entry := *f.repo.mockEntry
	entry.Status = domain.OvertimeStatusApproved
	approvedByName := "Admin User"
	entry.ApprovedByName = &approvedByName
	entry.ApprovedByEmployeeID = &approvedByEmployeeID
	return &entry, nil
}

func (f *fakeOvertimeTxRepository) RejectOvertimeEntry(
	ctx context.Context,
	id uuid.UUID,
	rejectionReason *string,
) (*domain.OvertimeEntry, error) {
	entry := *f.repo.mockEntry
	entry.Status = domain.OvertimeStatusRejected
	entry.RejectionReason = rejectionReason
	return &entry, nil
}

func (f *fakeOvertimeTxRepository) UpdateOvertimeEntryByAdmin(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	return nil, nil
}

func (f *fakeOvertimeRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.OvertimeTxRepository) error,
) error {
	tx := &fakeOvertimeTxRepository{repo: f}
	return fn(tx)
}

func (f *fakeOvertimeRepository) CreateOvertimeEntry(
	ctx context.Context,
	params domain.CreateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	f.lastParams = params
	return &domain.OvertimeEntry{
		ID:           uuid.New(),
		EmployeeID:   params.EmployeeID,
		EmployeeName: "John Doe",
		EntryDate:    params.EntryDate,
		Minutes:      params.Minutes,
		Reason:       params.Reason,
		Description:  params.Description,
		Status:       domain.OvertimeStatusSubmitted,
		SubmittedAt:  time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (f *fakeOvertimeRepository) GetOvertimeEntryByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.OvertimeEntry, error) {
	return nil, nil
}

func (f *fakeOvertimeRepository) ListOvertimeEntries(
	ctx context.Context,
	params domain.ListOvertimeEntriesParams,
) (*domain.OvertimeEntryPage, error) {
	return nil, nil
}

func (f *fakeOvertimeRepository) ListMyOvertimeEntries(
	ctx context.Context,
	params domain.ListMyOvertimeEntriesParams,
) (*domain.OvertimeEntryPage, error) {
	return nil, nil
}

func (f *fakeOvertimeRepository) GetCurrentMonthOvertimeStats(
	ctx context.Context,
) (*domain.OvertimeStats, error) {
	return nil, nil
}

func (f *fakeOvertimeRepository) GetMyCurrentMonthOvertimeStats(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.OvertimeStats, error) {
	return nil, nil
}

func TestOvertimeServiceCreateOvertimeEntryTriggersNotification(t *testing.T) {
	repo := &fakeOvertimeRepository{}
	ns := &fakeNotificationService{}
	empRepo := &fakeEmployeeRepository{}
	svc := NewOvertimeService(repo, empRepo, ns, nil)

	actorID := uuid.New()
	date := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	params := domain.CreateOvertimeEntryParams{
		EntryDate: date,
		Minutes:   120,
		Reason:    "emergency",
	}

	entry, err := svc.CreateOvertimeEntry(context.Background(), actorID, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry to be non-nil")
	}

	if ns.callsCount != 1 {
		t.Fatalf("expected Notify to be called 1 time, got %d", ns.callsCount)
	}

	req := ns.lastRequest
	if len(req.Recipients.Roles) != 1 || req.Recipients.Roles[0] != "admin" {
		t.Errorf("expected recipient role admin, got %v", req.Recipients.Roles)
	}

	expectedMsg := "John Doe has requested 120 minutes of overtime for the shift on 2026-06-10."
	if req.Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, req.Message)
	}

	data, ok := req.Data.(domain.OvertimeRequestCreatedNotificationData)
	if !ok {
		t.Fatalf(
			"expected NotificationData of type OvertimeRequestCreatedNotificationData, got %T",
			req.Data,
		)
	}

	if data.OvertimeEntryID != entry.ID {
		t.Errorf("expected OvertimeEntryID %s, got %s", entry.ID, data.OvertimeEntryID)
	}

	if data.EmployeeID != actorID {
		t.Errorf("expected EmployeeID %s, got %s", actorID, data.EmployeeID)
	}

	if data.EmployeeName != "John Doe" {
		t.Errorf("expected EmployeeName 'John Doe', got %q", data.EmployeeName)
	}

	if data.Minutes != 120 {
		t.Errorf("expected Minutes 120, got %d", data.Minutes)
	}

	if !data.EntryDate.Equal(date) {
		t.Errorf("expected EntryDate %v, got %v", date, data.EntryDate)
	}

	if data.Reason != "emergency" {
		t.Errorf("expected Reason 'emergency', got %q", data.Reason)
	}
}

func TestOvertimeServiceDecideOvertimeEntryApproveTriggersNotification(t *testing.T) {
	actorID := uuid.New()
	adminID := uuid.New()
	date := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)

	entry := &domain.OvertimeEntry{
		ID:           uuid.New(),
		EmployeeID:   actorID,
		EmployeeName: "John Doe",
		EntryDate:    date,
		Minutes:      120,
		Reason:       "emergency",
		Status:       domain.OvertimeStatusSubmitted,
	}

	repo := &fakeOvertimeRepository{mockEntry: entry}
	ns := &fakeNotificationService{}
	empRepo := &fakeEmployeeRepository{}
	svc := NewOvertimeService(repo, empRepo, ns, nil)

	updated, err := svc.DecideOvertimeEntryByAdmin(
		context.Background(),
		adminID,
		entry.ID,
		domain.DecideOvertimeEntryParams{
			Decision: "approve",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Status != domain.OvertimeStatusApproved {
		t.Errorf("expected status %s, got %s", domain.OvertimeStatusApproved, updated.Status)
	}

	if ns.callsCount != 1 {
		t.Fatalf("expected Notify to be called 1 time, got %d", ns.callsCount)
	}

	req := ns.lastRequest
	if len(req.Recipients.EmployeeIDs) != 1 || req.Recipients.EmployeeIDs[0] != actorID {
		t.Errorf(
			"expected recipient EmployeeIDs to contain %s, got %v",
			actorID,
			req.Recipients.EmployeeIDs,
		)
	}

	expectedMsg := "Your overtime request for 120 minutes on 2026-06-10 has been approved by Admin User."
	if req.Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, req.Message)
	}

	data, ok := req.Data.(domain.OvertimeRequestDecidedNotificationData)
	if !ok {
		t.Fatalf(
			"expected NotificationData of type OvertimeRequestDecidedNotificationData, got %T",
			req.Data,
		)
	}

	if data.Status != domain.OvertimeStatusApproved {
		t.Errorf("expected payload status %s, got %s", domain.OvertimeStatusApproved, data.Status)
	}

	if data.DecidedByEmployeeID != adminID {
		t.Errorf("expected DecidedByEmployeeID %s, got %s", adminID, data.DecidedByEmployeeID)
	}

	if data.DecidedByName != "Admin User" {
		t.Errorf("expected DecidedByName 'Admin User', got %q", data.DecidedByName)
	}
}

func TestOvertimeServiceDecideOvertimeEntryRejectTriggersNotification(t *testing.T) {
	actorID := uuid.New()
	adminID := uuid.New()
	date := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)

	entry := &domain.OvertimeEntry{
		ID:           uuid.New(),
		EmployeeID:   actorID,
		EmployeeName: "John Doe",
		EntryDate:    date,
		Minutes:      120,
		Reason:       "emergency",
		Status:       domain.OvertimeStatusSubmitted,
	}

	repo := &fakeOvertimeRepository{mockEntry: entry}
	ns := &fakeNotificationService{}
	empRepo := &fakeEmployeeRepository{}
	svc := NewOvertimeService(repo, empRepo, ns, nil)

	reason := "Not needed"
	updated, err := svc.DecideOvertimeEntryByAdmin(
		context.Background(),
		adminID,
		entry.ID,
		domain.DecideOvertimeEntryParams{
			Decision:        "reject",
			RejectionReason: &reason,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Status != domain.OvertimeStatusRejected {
		t.Errorf("expected status %s, got %s", domain.OvertimeStatusRejected, updated.Status)
	}

	if ns.callsCount != 1 {
		t.Fatalf("expected Notify to be called 1 time, got %d", ns.callsCount)
	}

	req := ns.lastRequest
	if len(req.Recipients.EmployeeIDs) != 1 || req.Recipients.EmployeeIDs[0] != actorID {
		t.Errorf(
			"expected recipient EmployeeIDs to contain %s, got %v",
			actorID,
			req.Recipients.EmployeeIDs,
		)
	}

	expectedMsg := "Your overtime request for 120 minutes on 2026-06-10 has been rejected by Admin User. Reason: Not needed"
	if req.Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, req.Message)
	}

	data, ok := req.Data.(domain.OvertimeRequestDecidedNotificationData)
	if !ok {
		t.Fatalf(
			"expected NotificationData of type OvertimeRequestDecidedNotificationData, got %T",
			req.Data,
		)
	}

	if data.Status != domain.OvertimeStatusRejected {
		t.Errorf("expected payload status %s, got %s", domain.OvertimeStatusRejected, data.Status)
	}

	if data.RejectionReason != "Not needed" {
		t.Errorf("expected payload RejectionReason 'Not needed', got %q", data.RejectionReason)
	}
}
