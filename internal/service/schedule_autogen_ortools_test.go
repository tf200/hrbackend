//go:build ortools

package service

import (
	"context"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestGenerateSchedulesWithORToolsFillsSecondStaffUpToTargets(t *testing.T) {
	service := &ScheduleService{}
	employees := []domain.ScheduleEmployeeContractHours{
		newAutoGenEmployee("Ava", 40),
		newAutoGenEmployee("Ben", 40),
	}
	shifts := []domain.ScheduleLocationShift{
		newAutoGenShift("Day", 9, 17),
	}

	resp, err := service.generateSchedulesWithORTools(
		context.Background(),
		uuid.New(),
		"UTC",
		time.UTC,
		employees,
		shifts,
		2,
		2026,
	)
	if err != nil {
		t.Fatalf("generateSchedulesWithORTools returned error: %v", err)
	}

	if resp.Status != "optimal" && resp.Status != "feasible" {
		t.Fatalf("expected feasible solution, got %q", resp.Status)
	}
	if resp.Constraints.MaxStaffPerShift != 2 {
		t.Fatalf("expected max staff per shift 2, got %d", resp.Constraints.MaxStaffPerShift)
	}
	if resp.Constraints.AllowEmptyShift {
		t.Fatalf("expected allow_empty_shift to be false")
	}

	totalAssignedMinutes := int64(0)
	doubleStaffedSlots := 0
	for _, slot := range resp.Slots {
		if len(slot.EmployeeIDs) < 1 || len(slot.EmployeeIDs) > 2 {
			t.Fatalf("expected slot staffing between 1 and 2, got %d", len(slot.EmployeeIDs))
		}
		totalAssignedMinutes += int64(len(slot.EmployeeIDs)) * 8 * 60
		if len(slot.EmployeeIDs) == 2 {
			doubleStaffedSlots++
		}
	}

	const expectedAssignedMinutes = int64(80 * 60)
	if totalAssignedMinutes != expectedAssignedMinutes {
		t.Fatalf("expected %d assigned minutes, got %d", expectedAssignedMinutes, totalAssignedMinutes)
	}
	if doubleStaffedSlots != 3 {
		t.Fatalf("expected 3 double-staffed slots, got %d", doubleStaffedSlots)
	}

	assertGeneratedScheduleRespectsDailyAndRestRules(t, resp, shifts)
}

func TestGenerateSchedulesWithORToolsAvoidsExtraOvertime(t *testing.T) {
	service := &ScheduleService{}
	employees := []domain.ScheduleEmployeeContractHours{
		newAutoGenEmployee("Ava", 28),
		newAutoGenEmployee("Ben", 28),
	}
	shifts := []domain.ScheduleLocationShift{
		newAutoGenShift("Day", 9, 17),
	}

	resp, err := service.generateSchedulesWithORTools(
		context.Background(),
		uuid.New(),
		"UTC",
		time.UTC,
		employees,
		shifts,
		2,
		2026,
	)
	if err != nil {
		t.Fatalf("generateSchedulesWithORTools returned error: %v", err)
	}

	totalAssignedMinutes := int64(0)
	doubleStaffedSlots := 0
	for _, slot := range resp.Slots {
		totalAssignedMinutes += int64(len(slot.EmployeeIDs)) * 8 * 60
		if len(slot.EmployeeIDs) == 2 {
			doubleStaffedSlots++
		}
	}

	const expectedAssignedMinutes = int64(56 * 60)
	if totalAssignedMinutes != expectedAssignedMinutes {
		t.Fatalf("expected %d assigned minutes, got %d", expectedAssignedMinutes, totalAssignedMinutes)
	}
	if doubleStaffedSlots != 0 {
		t.Fatalf("expected no double-staffed slots once targets are exhausted, got %d", doubleStaffedSlots)
	}

	assertGeneratedScheduleRespectsDailyAndRestRules(t, resp, shifts)
}

func assertGeneratedScheduleRespectsDailyAndRestRules(
	t *testing.T,
	resp *domain.AutoGenerateSchedulesResponse,
	shifts []domain.ScheduleLocationShift,
) {
	t.Helper()

	shiftByID := make(map[uuid.UUID]domain.ScheduleLocationShift, len(shifts))
	for _, shift := range shifts {
		shiftByID[shift.ID] = shift
	}

	assignByEmployee := make(map[uuid.UUID]map[string]uuid.UUID)
	for _, slot := range resp.Slots {
		for _, employeeID := range slot.EmployeeIDs {
			byDate, ok := assignByEmployee[employeeID]
			if !ok {
				byDate = make(map[string]uuid.UUID)
				assignByEmployee[employeeID] = byDate
			}
			if _, exists := byDate[slot.Date]; exists {
				t.Fatalf("employee %s assigned more than once on %s", employeeID, slot.Date)
			}
			byDate[slot.Date] = slot.ShiftID
		}
	}

	locationTZ := time.UTC
	weekStart, err := time.ParseInLocation("2006-01-02", resp.WeekStartDate, locationTZ)
	if err != nil {
		t.Fatalf("parse week start: %v", err)
	}

	dateList := make([]string, 0, 7)
	for d := 0; d < 7; d++ {
		dateList = append(dateList, weekStart.AddDate(0, 0, d).Format("2006-01-02"))
	}

	for employeeID, byDate := range assignByEmployee {
		for d := 0; d < len(dateList)-1; d++ {
			curDate := dateList[d]
			nextDate := dateList[d+1]
			curShiftID, okA := byDate[curDate]
			nextShiftID, okB := byDate[nextDate]
			if !okA || !okB {
				continue
			}

			curShift := shiftByID[curShiftID]
			nextShift := shiftByID[nextShiftID]
			curStartMin := curShift.StartMicroseconds / (60 * 1_000_000)
			curEndMin := curShift.EndMicroseconds / (60 * 1_000_000)
			curEndAbs := int64(d)*1440 + curEndMin
			if curEndMin < curStartMin {
				curEndAbs += 1440
			}

			nextStartMin := nextShift.StartMicroseconds / (60 * 1_000_000)
			nextStartAbs := int64(d+1)*1440 + nextStartMin
			if nextStartAbs-curEndAbs < int64(8*60) {
				t.Fatalf(
					"employee %s violates minimum rest between %s and %s",
					employeeID,
					curDate,
					nextDate,
				)
			}
		}
	}
}

func newAutoGenEmployee(firstName string, hours float64) domain.ScheduleEmployeeContractHours {
	return domain.ScheduleEmployeeContractHours{
		ID:            uuid.New(),
		FirstName:     firstName,
		LastName:      "Test",
		ContractHours: &hours,
	}
}

func newAutoGenShift(name string, startHour int64, endHour int64) domain.ScheduleLocationShift {
	return domain.ScheduleLocationShift{
		ID:                uuid.New(),
		LocationID:        uuid.New(),
		ShiftName:         name,
		StartMicroseconds: startHour * 60 * 60 * 1_000_000,
		EndMicroseconds:   endHour * 60 * 60 * 1_000_000,
	}
}
