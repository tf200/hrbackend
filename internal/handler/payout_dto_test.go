package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPreviewPayrollRequestBindsUUIDQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(
		http.MethodGet,
		"/payroll/preview?employee_id=a5514673-7217-476b-bbe3-07db2a725e12&period_start=2026-04-01&period_end=2026-04-30",
		nil,
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	var got previewPayrollRequest
	if err := ctx.ShouldBindQuery(&got); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.EmployeeID.String() != "a5514673-7217-476b-bbe3-07db2a725e12" {
		t.Fatalf("unexpected employee id: %s", got.EmployeeID.String())
	}
}
