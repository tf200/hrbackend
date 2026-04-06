package main

import (
	"testing"
)

func TestBuildGeneratedDatasetSeedsORTCoverageSamples(t *testing.T) {
	dataset := buildGeneratedDataset("testrun", 42)

	employeeByAlias := make(map[string]string, len(dataset.Employees))
	for _, item := range dataset.Employees {
		employeeByAlias[item.Alias] = item.IrregularHoursProfile
	}

	expectedProfiles := map[string]string{
		"finance_head":        "roster",
		"care_staff_01":       "non_roster",
		"operations_staff_01": "none",
		"planning_staff_02":   "none",
	}
	for alias, expected := range expectedProfiles {
		if employeeByAlias[alias] != expected {
			t.Fatalf("expected %s irregular_hours_profile %q, got %q", alias, expected, employeeByAlias[alias])
		}
	}

	assertApprovedEntry(t, dataset, "finance_head_head_wed_entry", "finance_head", "19:00", "23:00", "normal")
	assertApprovedEntry(t, dataset, "care_staff_01_ort_weekday_entry", "care_staff_01", "20:00", "23:00", "normal")
	assertApprovedEntry(t, dataset, "operations_staff_01_ort_saturday_entry", "operations_staff_01", "21:00", "23:30", "overtime")
	assertApprovedEntry(t, dataset, "planning_staff_02_ort_sunday_entry", "planning_staff_02", "12:00", "18:00", "travel")

	if !hasSchedule(dataset, "care_staff_01_ort_weekday") {
		t.Fatal("expected care_staff_01_ort_weekday schedule to exist")
	}
	if !hasSchedule(dataset, "operations_staff_01_ort_saturday") {
		t.Fatal("expected operations_staff_01_ort_saturday schedule to exist")
	}
	if !hasSchedule(dataset, "planning_staff_02_ort_sunday") {
		t.Fatal("expected planning_staff_02_ort_sunday schedule to exist")
	}

	var submittedCount int
	var rejectedCount int
	for _, item := range dataset.TimeEntries {
		switch item.Status {
		case "submitted":
			submittedCount++
		case "rejected":
			rejectedCount++
		}
	}
	if submittedCount == 0 {
		t.Fatal("expected at least one submitted time entry to remain in the dataset")
	}
	if rejectedCount == 0 {
		t.Fatal("expected at least one rejected time entry to remain in the dataset")
	}
}

func assertApprovedEntry(
	t *testing.T,
	dataset generatedDataset,
	alias, employeeAlias, startTime, endTime, hourType string,
) {
	t.Helper()

	for _, item := range dataset.TimeEntries {
		if item.Alias != alias {
			continue
		}
		if item.Status != "approved" {
			t.Fatalf("expected %s to be approved, got %s", alias, item.Status)
		}
		if item.EmployeeAlias != employeeAlias {
			t.Fatalf("expected %s employee alias %s, got %s", alias, employeeAlias, item.EmployeeAlias)
		}
		if item.StartTime != startTime || item.EndTime != endTime {
			t.Fatalf("expected %s time range %s-%s, got %s-%s", alias, startTime, endTime, item.StartTime, item.EndTime)
		}
		if item.HourType != hourType {
			t.Fatalf("expected %s hour type %s, got %s", alias, hourType, item.HourType)
		}
		return
	}

	t.Fatalf("expected time entry %s to exist", alias)
}

func hasSchedule(dataset generatedDataset, alias string) bool {
	for _, item := range dataset.Schedules {
		if item.Alias == alias {
			return true
		}
	}
	return false
}
