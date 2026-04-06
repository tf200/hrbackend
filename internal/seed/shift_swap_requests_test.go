package seed

import (
	"testing"
	"time"
)

func TestSameOptionalTime(t *testing.T) {
	now := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	same := now
	later := now.Add(time.Hour)

	if !sameOptionalTime(&now, &same) {
		t.Fatalf("expected identical timestamps to match")
	}
	if sameOptionalTime(&now, &later) {
		t.Fatalf("expected different timestamps to differ")
	}
	if !sameOptionalTime(nil, nil) {
		t.Fatalf("expected nil timestamps to match")
	}
}
