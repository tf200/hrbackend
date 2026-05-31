package handler

import (
	"testing"
	"time"

	"hrbackend/internal/domain"
)

func TestPayoutRequestResponseIncludesSalaryMonth(t *testing.T) {
	salaryMonth := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	response := toPayoutRequestResponse(domain.PayoutRequest{SalaryMonth: &salaryMonth})

	if response.SalaryMonth == nil {
		t.Fatalf("expected salary month")
	}
	if *response.SalaryMonth != "2026-05" {
		t.Fatalf("unexpected salary month: %s", *response.SalaryMonth)
	}
}
