package repository

import (
	"testing"
	"time"

	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
)

func TestToDomainLeaveCalendarGroupsRowsByEmployee(t *testing.T) {
	employeeID := uuid.New()
	firstLeaveID := uuid.New()
	secondLeaveID := uuid.New()
	departmentName := "Operations"

	rows := []db.ListLeaveCalendarRowsRow{
		{
			EmployeeID:        employeeID,
			EmployeeFirstName: "Jane",
			EmployeeLastName:  "Doe",
			DepartmentName:    &departmentName,
			LeaveRequestID:    secondLeaveID,
			LeaveType:         db.LeaveRequestTypeEnumSick,
			Status:            db.LeaveRequestStatusEnumApproved,
			StartDate:         conv.PgDateFromTime(time.Date(2026, time.April, 8, 0, 0, 0, 0, time.UTC)),
			EndDate:           conv.PgDateFromTime(time.Date(2026, time.April, 9, 0, 0, 0, 0, time.UTC)),
		},
		{
			EmployeeID:        employeeID,
			EmployeeFirstName: "Jane",
			EmployeeLastName:  "Doe",
			DepartmentName:    &departmentName,
			LeaveRequestID:    firstLeaveID,
			LeaveType:         db.LeaveRequestTypeEnumVacation,
			Status:            db.LeaveRequestStatusEnumPending,
			StartDate:         conv.PgDateFromTime(time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)),
			EndDate:           conv.PgDateFromTime(time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC)),
		},
	}

	got := toDomainLeaveCalendar(rows)
	if len(got) != 1 {
		t.Fatalf("expected 1 employee group, got %d", len(got))
	}
	if got[0].EmployeeName != "Jane Doe" {
		t.Fatalf("unexpected employee name: %s", got[0].EmployeeName)
	}
	if got[0].DepartmentName == nil || *got[0].DepartmentName != departmentName {
		t.Fatalf("unexpected department name: %#v", got[0].DepartmentName)
	}
	if len(got[0].LeaveRecords) != 2 {
		t.Fatalf("expected 2 leave records, got %d", len(got[0].LeaveRecords))
	}
	if got[0].LeaveRecords[0].LeaveRequestID != secondLeaveID {
		t.Fatalf("expected records to preserve query order")
	}
	if got[0].LeaveRecords[1].LeaveRequestID != firstLeaveID {
		t.Fatalf("expected second record id %s, got %s", firstLeaveID, got[0].LeaveRecords[1].LeaveRequestID)
	}
}

func TestToDBLeaveTypesDropsInvalidValues(t *testing.T) {
	got := toDBLeaveTypes([]string{"vacation", "invalid", " sick "})
	if len(got) != 2 {
		t.Fatalf("expected 2 valid leave types, got %d", len(got))
	}
	if got[0] != db.LeaveRequestTypeEnumVacation || got[1] != db.LeaveRequestTypeEnumSick {
		t.Fatalf("unexpected leave types: %#v", got)
	}
}
