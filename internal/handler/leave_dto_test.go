package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindListLeaveCalendarRequestReadsRepeatedLeaveTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(
		http.MethodGet,
		"/leave-requests/calendar?month=2026-04&leave_types=vacation&leave_types=sick&employee_search=jane",
		nil,
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	got, err := bindListLeaveCalendarRequest(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Month != "2026-04" {
		t.Fatalf("unexpected month: %s", got.Month)
	}
	if len(got.LeaveTypes) != 2 || got.LeaveTypes[0] != "vacation" || got.LeaveTypes[1] != "sick" {
		t.Fatalf("unexpected leave types: %#v", got.LeaveTypes)
	}
}

func TestBindListLeaveCalendarRequestBindsDepartmentID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(
		http.MethodGet,
		"/leave-requests/calendar?month=2026-04&department_id=a5514673-7217-476b-bbe3-07db2a725e12",
		nil,
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	got, err := bindListLeaveCalendarRequest(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.DepartmentID == nil {
		t.Fatalf("expected department id to be set")
	}
	if got.DepartmentID.String() != "a5514673-7217-476b-bbe3-07db2a725e12" {
		t.Fatalf("unexpected department id: %s", got.DepartmentID.String())
	}
}

func TestBindListLeaveCalendarRequestRejectsInvalidLeaveType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(
		http.MethodGet,
		"/leave-requests/calendar?month=2026-04&leave_types=invalid",
		nil,
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	_, err := bindListLeaveCalendarRequest(ctx)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestBindListLeaveCalendarRequestRejectsInvalidMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(
		http.MethodGet,
		"/leave-requests/calendar?month=2026-4",
		nil,
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	_, err := bindListLeaveCalendarRequest(ctx)
	if err == nil {
		t.Fatalf("expected bind error")
	}
}
