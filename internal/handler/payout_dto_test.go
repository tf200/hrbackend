package handler

import (
	"testing"
	"time"

	"hrbackend/internal/domain"
)

func TestPayoutRequestResponseIncludesPayPeriodStart(t *testing.T) {
	payPeriodStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	response := toPayoutRequestResponse(domain.PayoutRequest{PayPeriodStart: &payPeriodStart})

	if response.PayPeriodStart == nil {
		t.Fatalf("expected pay period start")
	}
	if *response.PayPeriodStart != "2026-05-01" {
		t.Fatalf("unexpected pay period start: %s", *response.PayPeriodStart)
	}
}
