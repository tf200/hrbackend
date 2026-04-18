package service

import (
	"context"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestLeaveServiceListLeaveCalendarRejectsZeroMonth(t *testing.T) {
	svc := &LeaveService{repository: &fakeLeaveRepository{}}

	_, err := svc.ListLeaveCalendar(context.Background(), domain.ListLeaveCalendarParams{})
	if err != domain.ErrLeaveRequestInvalidRequest {
		t.Fatalf("expected %v, got %v", domain.ErrLeaveRequestInvalidRequest, err)
	}
}

func TestLeaveServiceListLeaveCalendarRejectsInvalidLeaveType(t *testing.T) {
	svc := &LeaveService{repository: &fakeLeaveRepository{}}

	_, err := svc.ListLeaveCalendar(context.Background(), domain.ListLeaveCalendarParams{
		Month:      time.Date(2026, time.April, 15, 12, 0, 0, 0, time.UTC),
		LeaveTypes: []string{"vacation", "invalid"},
	})
	if err != domain.ErrLeaveRequestInvalidRequest {
		t.Fatalf("expected %v, got %v", domain.ErrLeaveRequestInvalidRequest, err)
	}
}

func TestLeaveServiceListLeaveCalendarNormalizesMonthAndLeaveTypes(t *testing.T) {
	repo := &fakeLeaveRepository{}
	svc := &LeaveService{repository: repo}

	_, err := svc.ListLeaveCalendar(context.Background(), domain.ListLeaveCalendarParams{
		Month:      time.Date(2026, time.April, 15, 12, 30, 0, 0, time.FixedZone("X", 3*3600)),
		LeaveTypes: []string{" vacation ", "sick"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	got := repo.lastListLeaveCalendarParams
	wantMonth := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	if !got.Month.Equal(wantMonth) {
		t.Fatalf("expected normalized month %v, got %v", wantMonth, got.Month)
	}
	if len(got.LeaveTypes) != 2 || got.LeaveTypes[0] != "vacation" || got.LeaveTypes[1] != "sick" {
		t.Fatalf("unexpected leave types: %#v", got.LeaveTypes)
	}
}

type fakeLeaveRepository struct {
	lastListLeaveCalendarParams domain.ListLeaveCalendarParams
}

func (f *fakeLeaveRepository) WithTx(
	_ context.Context,
	_ func(tx domain.LeaveTxRepository) error,
) error {
	return nil
}

func (f *fakeLeaveRepository) CreateLeaveRequest(
	_ context.Context,
	_ domain.CreateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) GetActiveLeavePolicyByType(
	_ context.Context,
	_ string,
) (*domain.LeavePolicy, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) ListMyLeaveRequests(
	_ context.Context,
	_ domain.ListMyLeaveRequestsParams,
) (*domain.LeaveRequestPage, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) ListLeaveRequests(
	_ context.Context,
	_ domain.ListLeaveRequestsParams,
) (*domain.LeaveRequestPage, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) ListLeaveCalendar(
	_ context.Context,
	params domain.ListLeaveCalendarParams,
) ([]domain.LeaveCalendarEmployee, error) {
	f.lastListLeaveCalendarParams = params
	return []domain.LeaveCalendarEmployee{}, nil
}

func (f *fakeLeaveRepository) GetMyLeaveRequestStats(
	_ context.Context,
	_ uuid.UUID,
) (*domain.LeaveRequestStats, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) GetLeaveRequestStats(
	_ context.Context,
) (*domain.LeaveRequestStats, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) ListLeaveBalances(
	_ context.Context,
	_ domain.ListLeaveBalancesParams,
) (*domain.LeaveBalancePage, error) {
	return nil, nil
}

func (f *fakeLeaveRepository) ListMyLeaveBalances(
	_ context.Context,
	_ domain.ListMyLeaveBalancesParams,
) (*domain.LeaveBalancePage, error) {
	return nil, nil
}

var _ domain.LeaveRepository = (*fakeLeaveRepository)(nil)
