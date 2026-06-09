package domain

import (
	"testing"
	"time"
)

func TestResolvePayrollPeriod(t *testing.T) {
	tests := []struct {
		name      string
		date      time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name: "anchor start",
			date: time.Date(
				2025,
				time.December,
				29,
				15,
				30,
				0,
				0,
				time.FixedZone("test", 3600),
			),
			wantStart: time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "period crossing year",
			date:      time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
			wantStart: time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "next rolling period",
			date:      time.Date(2026, time.January, 26, 0, 0, 0, 0, time.UTC),
			wantStart: time.Date(2026, time.January, 26, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.February, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "previous rolling period",
			date:      time.Date(2025, time.December, 28, 0, 0, 0, 0, time.UTC),
			wantStart: time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2025, time.December, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := ResolvePayrollPeriod(tt.date)
			if !gotStart.Equal(tt.wantStart) {
				t.Fatalf("start = %s, want %s", gotStart, tt.wantStart)
			}
			if !gotEnd.Equal(tt.wantEnd) {
				t.Fatalf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestIsPayrollPeriodStart(t *testing.T) {
	if !IsPayrollPeriodStart(time.Date(2025, time.December, 29, 12, 0, 0, 0, time.UTC)) {
		t.Fatal("expected anchor date to be a period start")
	}
	if IsPayrollPeriodStart(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected non-anchor-aligned date not to be a period start")
	}
}

func TestPayrollPeriodOptionsThrough(t *testing.T) {
	options := PayrollPeriodOptionsThrough(time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC))
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
	if !options[0].PeriodStart.Equal(time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected first start: %s", options[0].PeriodStart)
	}
	if !options[0].PeriodEnd.Equal(time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected first end: %s", options[0].PeriodEnd)
	}
	if options[0].IsCurrent {
		t.Fatal("expected first option not to be current")
	}
	if !options[1].PeriodStart.Equal(time.Date(2026, time.January, 26, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected second start: %s", options[1].PeriodStart)
	}
	if !options[1].PeriodEnd.Equal(time.Date(2026, time.February, 22, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected second end: %s", options[1].PeriodEnd)
	}
	if !options[1].IsCurrent {
		t.Fatal("expected second option to be current")
	}
}

func TestPayrollPeriodOptionsThroughBeforeAnchor(t *testing.T) {
	options := PayrollPeriodOptionsThrough(time.Date(2025, time.December, 28, 0, 0, 0, 0, time.UTC))
	if len(options) != 0 {
		t.Fatalf("expected no options before anchor, got %d", len(options))
	}
}
