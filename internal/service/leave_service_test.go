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

func TestCalculateFullDayMinutesCountsWeekdays(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	fri := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		mon, fri,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int32(5 * 480)
	if minutes != expected {
		t.Fatalf("expected %d minutes, got %d", expected, minutes)
	}
}

func TestCalculateFullDayMinutesIncludesWeekendsInRange(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	wed := time.Date(2026, time.July, 8, 0, 0, 0, 0, time.UTC)
	nextWed := time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		wed, nextWed,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := int32(8 * 480)
	if minutes != expected {
		t.Fatalf("expected %d minutes, got %d", expected, minutes)
	}
}

func TestCalculateFullDayMinutesExcludesRosterFreeDay(t *testing.T) {
	repo := newFakeLeaveRepoWithRosterFreeDay()
	svc := &LeaveService{repository: repo}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	fri := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		mon, fri,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int32(4 * 480)
	if minutes != expected {
		t.Fatalf("expected %d minutes, got %d (friday is roster-free)", expected, minutes)
	}
}

func TestCalculateFullDayMinutesIncludesWeekendOnlyRangeWithRosterFreeConfig(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithRosterFreeDay()}

	sat := time.Date(2026, time.July, 11, 0, 0, 0, 0, time.UTC)
	sun := time.Date(2026, time.July, 12, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		sat, sun,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int32(2 * 480)
	if minutes != expected {
		t.Fatalf("expected %d minutes, got %d", expected, minutes)
	}
}

func TestCalculateFullDayMinutesExcludesRosterFreeFridayOnly(t *testing.T) {
	repo := newFakeLeaveRepoWithRosterFreeDay()
	svc := &LeaveService{repository: repo}

	tue := time.Date(2026, time.July, 7, 0, 0, 0, 0, time.UTC)
	fri := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		tue, fri,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int32(3 * 480)
	if minutes != expected {
		t.Fatalf("expected %d minutes, got %d", expected, minutes)
	}
}

func TestCalculateFullDayMinutesIncludesWeekendOnlyRange(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	sat := time.Date(2026, time.July, 11, 0, 0, 0, 0, time.UTC)
	sun := time.Date(2026, time.July, 12, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		sat, sun,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int32(2 * 480)
	if minutes != expected {
		t.Fatalf("expected %d minutes, got %d", expected, minutes)
	}
}

func TestCalculateFullDayMinutesSingleDay(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		mon, mon,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minutes != 480 {
		t.Fatalf("expected 480 minutes for single day, got %d", minutes)
	}
}

func TestCalculateFullDayMinutesRejectsRosterFreeDayOnly(t *testing.T) {
	repo := newFakeLeaveRepoWithRosterFreeDay()
	svc := &LeaveService{repository: repo}

	fri := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		fri, fri,
		nil, nil,
	)
	if err == nil {
		t.Fatal("expected error for roster-free day only")
	}
}

func TestCalculateHoursMinutesStandard(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		mon, mon,
		&startTime, &endTime,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minutes != 240 {
		t.Fatalf("expected 240 minutes, got %d", minutes)
	}
}

func TestCalculateHoursMinutesAllowsWeekend(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	sat := time.Date(2026, time.July, 11, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

	minutes, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		sat, sat,
		&startTime, &endTime,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minutes != 240 {
		t.Fatalf("expected 240 minutes, got %d", minutes)
	}
}

func TestCalculateHoursMinutesRejectsRosterFreeDay(t *testing.T) {
	repo := newFakeLeaveRepoWithRosterFreeDay()
	svc := &LeaveService{repository: repo}

	fri := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		fri, fri,
		&startTime, &endTime,
	)
	if err == nil {
		t.Fatal("expected error for hourly leave on roster-free day")
	}
}

func TestCalculateHoursMinutesRejectsMultiDay(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	tue := time.Date(2026, time.July, 7, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		mon, tue,
		&startTime, &endTime,
	)
	if err == nil {
		t.Fatal("expected error for multi-day hourly leave")
	}
}

func TestCalculateHoursMinutesRejectsExceedsFullDay(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		mon, mon,
		&startTime, &endTime,
	)
	if err == nil {
		t.Fatal("expected error for hourly leave exceeding full day")
	}
}

func TestCalculateHoursMinutesRejectsEndBeforeStart(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		mon, mon,
		&startTime, &endTime,
	)
	if err == nil {
		t.Fatal("expected error for hourly leave with end before start")
	}
}

func TestCalculateHoursMinutesRejectsMissingTimes(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"hours",
		mon, mon,
		nil, nil,
	)
	if err == nil {
		t.Fatal("expected error for missing times")
	}
}

func TestCalculateFullDayMinutesWithTimesRejected(t *testing.T) {
	svc := &LeaveService{repository: newFakeLeaveRepoWithUnlimitedContract()}

	mon := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)

	_, err := svc.calculateRequestedMinutes(
		context.Background(),
		uuid.Nil,
		"full_day",
		mon, mon,
		&startTime, nil,
	)
	if err == nil {
		t.Fatal("expected error for full_day with start_time")
	}
}

func TestNormalizeUpdateParamsPopulatesAllFields(t *testing.T) {
	current := domain.LeaveRequest{
		LeaveType:    "vacation",
		DurationType: "full_day",
		StartDate:    time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC),
		StartTime:    nil,
		EndTime:      nil,
		Reason:       strPtr("family event"),
		Status:       "pending",
	}

	update := domain.UpdateLeaveRequestParams{
		Reason: strPtr("updated reason"),
	}

	out, err := normalizeUpdateParams(current, update, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *out.updateParams.LeaveType != "vacation" {
		t.Fatalf("expected leave_type vacation, got %s", *out.updateParams.LeaveType)
	}
	if *out.updateParams.DurationType != "full_day" {
		t.Fatalf("expected duration_type full_day, got %s", *out.updateParams.DurationType)
	}
	if !out.updateParams.StartDate.Equal(current.StartDate) {
		t.Fatal("start_date should have been populated from current")
	}
	if !out.updateParams.EndDate.Equal(current.EndDate) {
		t.Fatal("end_date should have been populated from current")
	}
	if out.updateParams.StartTime != nil {
		t.Fatal("start_time should be nil for full_day")
	}
	if out.updateParams.EndTime != nil {
		t.Fatal("end_time should be nil for full_day")
	}
	if *out.updateParams.Reason != "updated reason" {
		t.Fatalf("expected 'updated reason', got '%s'", *out.updateParams.Reason)
	}
}

func TestNormalizeUpdateParamsSwitchesToHoursClearsDates(t *testing.T) {
	current := domain.LeaveRequest{
		LeaveType:    "vacation",
		DurationType: "full_day",
		StartDate:    time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC),
		Reason:       nil,
		Status:       "pending",
	}

	startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)
	update := domain.UpdateLeaveRequestParams{
		DurationType: strPtr("hours"),
		StartTime:    &startTime,
		EndTime:      &endTime,
	}

	out, err := normalizeUpdateParams(current, update, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *out.updateParams.DurationType != "hours" {
		t.Fatalf("expected hours, got %s", *out.updateParams.DurationType)
	}
	if out.effectiveStartTime == nil || !out.effectiveStartTime.Equal(startTime) {
		t.Fatal("start_time should be set")
	}
	if out.effectiveEndTime == nil || !out.effectiveEndTime.Equal(endTime) {
		t.Fatal("end_time should be set")
	}
}

func TestNormalizeUpdateParamsPartialUpdateKeepsExistingFields(t *testing.T) {
	current := domain.LeaveRequest{
		LeaveType:    "vacation",
		DurationType: "hours",
		StartDate:    time.Date(2026, time.July, 8, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, time.July, 8, 0, 0, 0, 0, time.UTC),
		StartTime:    timePtr(time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)),
		EndTime:      timePtr(time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)),
		Reason:       strPtr("appointment"),
		Status:       "pending",
	}

	newStart := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	update := domain.UpdateLeaveRequestParams{
		StartDate: &newStart,
		EndDate:   &newStart,
	}

	out, err := normalizeUpdateParams(current, update, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *out.updateParams.LeaveType != "vacation" {
		t.Fatalf("expected vacation, got %s", *out.updateParams.LeaveType)
	}
	if *out.updateParams.DurationType != "hours" {
		t.Fatalf("expected hours, got %s", *out.updateParams.DurationType)
	}
	if out.effectiveStartTime == nil {
		t.Fatal("start_time should be preserved from current")
	}
	if out.effectiveEndTime == nil {
		t.Fatal("end_time should be preserved from current")
	}
	if *out.updateParams.Reason != "appointment" {
		t.Fatalf("reason should be preserved from current")
	}
}

func TestDecideLeaveRequestUsesLiveLegalBalance(t *testing.T) {
	employeeID := uuid.New()
	adminID := uuid.New()
	leaveRequestID := uuid.New()
	start := currentUTCDate().AddDate(0, 0, 1)
	tx := &fakeLeaveTxRepository{
		request: &domain.LeaveRequest{
			ID:               leaveRequestID,
			EmployeeID:       employeeID,
			LeaveType:        "vacation",
			Status:           "pending",
			RequestedMinutes: 90,
			StartDate:        start,
			EndDate:          start,
		},
		policy:                 &domain.LeavePolicy{LeaveType: "vacation", DeductsBalance: true},
		legalCalculatedMinutes: 1000,
		legalUsedMinutes:       910,
	}
	svc := &LeaveService{repository: &fakeLeaveRepository{tx: tx}}

	_, err := svc.DecideLeaveRequestByAdmin(
		context.Background(),
		adminID,
		leaveRequestID,
		domain.DecideLeaveRequestParams{Decision: "approve"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tx.lockedEmployeeForBalance {
		t.Fatal("expected employee balance lock")
	}
}

func TestDecideLeaveRequestRejectsInsufficientLiveBalance(t *testing.T) {
	employeeID := uuid.New()
	adminID := uuid.New()
	leaveRequestID := uuid.New()
	start := currentUTCDate().AddDate(0, 0, 1)
	tx := &fakeLeaveTxRepository{
		request: &domain.LeaveRequest{
			ID:               leaveRequestID,
			EmployeeID:       employeeID,
			LeaveType:        "vacation",
			Status:           "pending",
			RequestedMinutes: 100,
			StartDate:        start,
			EndDate:          start,
		},
		policy:                 &domain.LeavePolicy{LeaveType: "vacation", DeductsBalance: true},
		legalCalculatedMinutes: 30,
	}
	svc := &LeaveService{repository: &fakeLeaveRepository{tx: tx}}

	_, err := svc.DecideLeaveRequestByAdmin(
		context.Background(),
		adminID,
		leaveRequestID,
		domain.DecideLeaveRequestParams{Decision: "approve"},
	)
	if err != domain.ErrLeaveBalanceInsufficient {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
}

func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

type fakeLeaveRepository struct {
	lastListLeaveCalendarParams domain.ListLeaveCalendarParams
	contractFunc                func(employeeID uuid.UUID, date time.Time) (*domain.LeaveContractAtDate, error)
	tx                          domain.LeaveTxRepository
}

func newFakeLeaveRepoWithUnlimitedContract() *fakeLeaveRepository {
	return &fakeLeaveRepository{
		contractFunc: func(_ uuid.UUID, _ time.Time) (*domain.LeaveContractAtDate, error) {
			return &domain.LeaveContractAtDate{RosterFreeDay: ""}, nil
		},
	}
}

func newFakeLeaveRepoWithRosterFreeDay() *fakeLeaveRepository {
	return &fakeLeaveRepository{
		contractFunc: func(_ uuid.UUID, date time.Time) (*domain.LeaveContractAtDate, error) {
			if date.Weekday() == time.Friday {
				return &domain.LeaveContractAtDate{RosterFreeDay: "friday"}, nil
			}
			return &domain.LeaveContractAtDate{RosterFreeDay: ""}, nil
		},
	}
}

func (f *fakeLeaveRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.LeaveTxRepository) error,
) error {
	if f.tx != nil {
		return fn(f.tx)
	}
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

func (f *fakeLeaveRepository) GetEmployeeContractAtDate(
	_ context.Context,
	_ uuid.UUID,
	date time.Time,
) (*domain.LeaveContractAtDate, error) {
	if f.contractFunc == nil {
		return &domain.LeaveContractAtDate{RosterFreeDay: ""}, nil
	}
	return f.contractFunc(uuid.Nil, date)
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

func (f *fakeLeaveRepository) GetLeaveBalanceDetails(
	_ context.Context,
	_ domain.GetLeaveBalanceDetailsParams,
) (*domain.LeaveBalanceDetails, error) {
	return nil, nil
}

var _ domain.LeaveRepository = (*fakeLeaveRepository)(nil)

type fakeLeaveTxRepository struct {
	request                  *domain.LeaveRequest
	policy                   *domain.LeavePolicy
	legalCalculatedMinutes   int32
	legalUsedMinutes         int32
	lockedEmployeeForBalance bool
}

func (f *fakeLeaveTxRepository) GetLeaveRequestForUpdate(
	_ context.Context,
	_ uuid.UUID,
) (*domain.LeaveRequest, error) {
	return f.request, nil
}

func (f *fakeLeaveTxRepository) UpdateLeaveRequestEditableFields(
	_ context.Context,
	_ uuid.UUID,
	_ domain.UpdateLeaveRequestParams,
) (*domain.LeaveRequest, error) {
	return f.request, nil
}

func (f *fakeLeaveTxRepository) UpdateLeaveRequestDecision(
	_ context.Context,
	_ uuid.UUID,
	status string,
	decisionNote *string,
	decidedByEmployeeID uuid.UUID,
) (*domain.LeaveRequest, error) {
	updated := *f.request
	updated.Status = status
	updated.DecisionNote = decisionNote
	updated.DecidedByEmployeeID = &decidedByEmployeeID
	return &updated, nil
}

func (f *fakeLeaveTxRepository) GetActiveLeavePolicyByType(
	_ context.Context,
	_ string,
) (*domain.LeavePolicy, error) {
	return f.policy, nil
}

func (f *fakeLeaveTxRepository) LockEmployeeForLeaveBalance(_ context.Context, _ uuid.UUID) error {
	f.lockedEmployeeForBalance = true
	return nil
}

func (f *fakeLeaveTxRepository) GetLeaveHoursPerDay(_ context.Context, _ uuid.UUID) (int32, error) {
	return 8, nil
}

func (f *fakeLeaveTxRepository) GetEmployeeContractAtDate(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
) (*domain.LeaveContractAtDate, error) {
	return &domain.LeaveContractAtDate{}, nil
}

func (f *fakeLeaveTxRepository) ComputeLegalLeaveTotalForYear(
	_ context.Context,
	_ uuid.UUID,
	_ int32,
	_ time.Time,
) (int32, error) {
	return f.legalCalculatedMinutes, nil
}

func (f *fakeLeaveTxRepository) ComputeLegalLeaveUsedForYear(
	_ context.Context,
	_ uuid.UUID,
	_ int32,
) (int32, error) {
	return f.legalUsedMinutes, nil
}

func TestAllocateLeaveUsageToContractAccrualsNoUsage(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 800},
		{ContractID: uuid.New(), GainedMinutes: 1200},
	}

	allocateLeaveUsageToContractAccruals(accruals, 0)

	if accruals[0].DeductedMinutes != 0 || accruals[0].RemainingMinutes != 800 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
	if accruals[1].DeductedMinutes != 0 || accruals[1].RemainingMinutes != 1200 {
		t.Fatalf("unexpected second accrual: %#v", accruals[1])
	}
}

func TestAllocateLeaveUsageToContractAccrualsWithinFirstSegment(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 800},
		{ContractID: uuid.New(), GainedMinutes: 1200},
	}

	allocateLeaveUsageToContractAccruals(accruals, 200)

	if accruals[0].DeductedMinutes != 200 || accruals[0].RemainingMinutes != 600 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
	if accruals[1].DeductedMinutes != 0 || accruals[1].RemainingMinutes != 1200 {
		t.Fatalf("unexpected second accrual: %#v", accruals[1])
	}
}

func TestAllocateLeaveUsageToContractAccrualsExhaustsFirstSegment(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 800},
		{ContractID: uuid.New(), GainedMinutes: 1200},
	}

	allocateLeaveUsageToContractAccruals(accruals, 800)

	if accruals[0].DeductedMinutes != 800 || accruals[0].RemainingMinutes != 0 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
	if accruals[1].DeductedMinutes != 0 || accruals[1].RemainingMinutes != 1200 {
		t.Fatalf("unexpected second accrual: %#v", accruals[1])
	}
}

func TestAllocateLeaveUsageToContractAccrualsSpansMultipleSegments(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 800},
		{ContractID: uuid.New(), GainedMinutes: 1200},
		{ContractID: uuid.New(), GainedMinutes: 600},
	}

	allocateLeaveUsageToContractAccruals(accruals, 1000)

	if accruals[0].DeductedMinutes != 800 || accruals[0].RemainingMinutes != 0 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
	if accruals[1].DeductedMinutes != 200 || accruals[1].RemainingMinutes != 1000 {
		t.Fatalf("unexpected second accrual: %#v", accruals[1])
	}
	if accruals[2].DeductedMinutes != 0 || accruals[2].RemainingMinutes != 600 {
		t.Fatalf("unexpected third accrual: %#v", accruals[2])
	}
}

func TestAllocateLeaveUsageToContractAccrualsExceedsTotalGained(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 800},
		{ContractID: uuid.New(), GainedMinutes: 1200},
	}

	allocateLeaveUsageToContractAccruals(accruals, 3000)

	if accruals[0].DeductedMinutes != 800 || accruals[0].RemainingMinutes != 0 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
	if accruals[1].DeductedMinutes != 1200 || accruals[1].RemainingMinutes != 0 {
		t.Fatalf("unexpected second accrual: %#v", accruals[1])
	}
}

func TestAllocateLeaveUsageToContractAccrualsZeroGainedSegment(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 0},
		{ContractID: uuid.New(), GainedMinutes: 1000},
	}

	allocateLeaveUsageToContractAccruals(accruals, 500)

	if accruals[0].DeductedMinutes != 0 || accruals[0].RemainingMinutes != 0 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
	if accruals[1].DeductedMinutes != 500 || accruals[1].RemainingMinutes != 500 {
		t.Fatalf("unexpected second accrual: %#v", accruals[1])
	}
}

func TestAllocateLeaveUsageToContractAccrualsEmptySlice(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{}
	allocateLeaveUsageToContractAccruals(accruals, 100)
}

func TestAllocateLeaveUsageToContractAccrualsNegativeUsageClampedToZero(t *testing.T) {
	accruals := []domain.LeaveContractAccrual{
		{ContractID: uuid.New(), GainedMinutes: 800},
	}

	allocateLeaveUsageToContractAccruals(accruals, -50)

	if accruals[0].DeductedMinutes != 0 || accruals[0].RemainingMinutes != 800 {
		t.Fatalf("unexpected first accrual: %#v", accruals[0])
	}
}

var _ domain.LeaveTxRepository = (*fakeLeaveTxRepository)(nil)
