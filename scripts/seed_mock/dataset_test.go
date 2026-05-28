package main

import (
	"testing"
)

func TestBuildGeneratedDatasetSeedsORTCoverageSamples(t *testing.T) {
	dataset := buildGeneratedDataset("testrun", 42)

	if !hasSchedule(dataset, "care_staff_01_ort_weekday") {
		t.Fatal("expected care_staff_01_ort_weekday schedule to exist")
	}
	if !hasSchedule(dataset, "operations_staff_01_ort_saturday") {
		t.Fatal("expected operations_staff_01_ort_saturday schedule to exist")
	}
	if !hasSchedule(dataset, "planning_staff_02_ort_sunday") {
		t.Fatal("expected planning_staff_02_ort_sunday schedule to exist")
	}
}

func hasSchedule(dataset generatedDataset, alias string) bool {
	for _, item := range dataset.Schedules {
		if item.Alias == alias {
			return true
		}
	}
	return false
}
