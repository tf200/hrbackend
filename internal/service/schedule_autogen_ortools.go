//go:build ortools

package service

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/or-tools/ortools/sat/go/cpmodel"
	cmpb "github.com/google/or-tools/ortools/sat/proto/cpmodel"
	sppb "github.com/google/or-tools/ortools/sat/proto/satparameters"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type autoGenEmployee struct {
	ID            uuid.UUID
	FirstName     string
	LastName      string
	TargetMinutes int64
}

type autoGenShift struct {
	ID              uuid.UUID
	Name            string
	StartMinutes    int
	EndMinutes      int
	DurationMinutes int64
}

func (s *ScheduleService) AutoGenerateSchedules(
	ctx context.Context,
	req *domain.AutoGenerateSchedulesRequest,
) (*domain.AutoGenerateSchedulesResponse, error) {
	if err := s.validateAutoGenerateRequest(req); err != nil {
		return nil, err
	}

	employees, locationShifts, timezone, locationTZ, err := s.loadAutoGenerateInputs(ctx, req)
	if err != nil {
		return nil, err
	}

	return s.generateSchedulesWithORTools(
		ctx,
		req.LocationID,
		timezone,
		locationTZ,
		employees,
		locationShifts,
		req.Week,
		req.Year,
	)
}

func (s *ScheduleService) generateSchedulesWithORTools(
	ctx context.Context,
	locationID uuid.UUID,
	timezone string,
	locationTZ *time.Location,
	employees []domain.ScheduleEmployeeContractHours,
	locationShifts []domain.ScheduleLocationShift,
	week int32,
	year int32,
) (*domain.AutoGenerateSchedulesResponse, error) {
	const (
		minStaffPerShift = int64(1)
		maxStaffPerShift = int64(2)
		maxSolveSeconds  = 90.0
		minRestMinutes   = int64(8 * 60)
	)

	weekStart, err := isoWeekStartDate(int(year), int(week), locationTZ)
	if err != nil {
		return nil, err
	}
	weekStartStr := weekStart.Format("2006-01-02")

	inputsEmp := make([]autoGenEmployee, 0, len(employees))
	for _, e := range employees {
		if e.ContractHours == nil || *e.ContractHours <= 0 {
			continue
		}
		targetMinutes := int64(math.Round(*e.ContractHours * 60.0))
		inputsEmp = append(inputsEmp, autoGenEmployee{
			ID:            e.ID,
			FirstName:     e.FirstName,
			LastName:      e.LastName,
			TargetMinutes: targetMinutes,
		})
	}
	if len(inputsEmp) == 0 {
		return nil, fmt.Errorf("no employees with positive contract hours")
	}

	inputsShifts := make([]autoGenShift, 0, len(locationShifts))
	var maxShiftMinutes int64
	for _, ls := range locationShifts {
		startMin := int(ls.StartMicroseconds / (60 * 1_000_000))
		endMin := int(ls.EndMicroseconds / (60 * 1_000_000))
		dur := int64(endMin - startMin)
		if endMin < startMin {
			dur = int64(endMin + 1440 - startMin)
		}
		if dur <= 0 {
			continue
		}
		if dur > maxShiftMinutes {
			maxShiftMinutes = dur
		}
		inputsShifts = append(inputsShifts, autoGenShift{
			ID:              ls.ID,
			Name:            ls.ShiftName,
			StartMinutes:    startMin,
			EndMinutes:      endMin,
			DurationMinutes: dur,
		})
	}
	if len(inputsShifts) == 0 {
		return nil, fmt.Errorf("no valid shifts found for location")
	}

	dCount := 7
	planEmployees := make([]domain.SchedulePlanEmployee, 0, len(inputsEmp))
	for _, emp := range inputsEmp {
		planEmployees = append(planEmployees, domain.SchedulePlanEmployee{
			ID:            emp.ID,
			FirstName:     emp.FirstName,
			LastName:      emp.LastName,
			TargetMinutes: emp.TargetMinutes,
		})
	}

	shiftTemplates := make([]domain.ScheduleShiftTemplate, 0, len(inputsShifts))
	for _, sh := range inputsShifts {
		overnight := sh.EndMinutes < sh.StartMinutes
		shiftTemplates = append(shiftTemplates, domain.ScheduleShiftTemplate{
			ShiftID:         sh.ID,
			Name:            sh.Name,
			StartMinute:     int32(sh.StartMinutes),
			EndMinute:       int32(sh.EndMinutes),
			DurationMinutes: sh.DurationMinutes,
			Overnight:       overnight,
		})
	}

	emptySlots := make([]domain.SchedulePlanSlot, 0, dCount*len(inputsShifts))
	for dIdx := 0; dIdx < dCount; dIdx++ {
		dateStr := weekStart.AddDate(0, 0, dIdx).Format("2006-01-02")
		for _, sh := range inputsShifts {
			emptySlots = append(
				emptySlots,
				domain.SchedulePlanSlot{Date: dateStr, ShiftID: sh.ID, EmployeeIDs: []uuid.UUID{}},
			)
		}
	}

	minEmployeesPerDay := int(minStaffPerShift) * len(inputsShifts)
	if len(inputsEmp) < minEmployeesPerDay {
		return &domain.AutoGenerateSchedulesResponse{
			Status:        "infeasible",
			PlanID:        uuid.New(),
			LocationID:    locationID,
			Timezone:      timezone,
			Week:          week,
			Year:          year,
			WeekStartDate: weekStartStr,
			Constraints: domain.SchedulePlanConstraints{
				MaxStaffPerShift: int32(maxStaffPerShift),
				AllowEmptyShift:  true,
			},
			Employees:      planEmployees,
			ShiftTemplates: shiftTemplates,
			Slots:          emptySlots,
			Summary:        []domain.ScheduleEmployeeSummary{},
			Warnings: []domain.SchedulePlanWarning{
				{
					Code:    "INFEASIBLE",
					Message: "Not enough employees to staff all shifts (1 shift/day per employee).",
				},
			},
		}, nil
	}

	model := cpmodel.NewCpModelBuilder()
	eCount := len(inputsEmp)
	sCount := len(inputsShifts)
	assign := make([][][]cpmodel.BoolVar, eCount)
	for eIdx := 0; eIdx < eCount; eIdx++ {
		assign[eIdx] = make([][]cpmodel.BoolVar, dCount)
		for dIdx := 0; dIdx < dCount; dIdx++ {
			assign[eIdx][dIdx] = make([]cpmodel.BoolVar, sCount)
			for shIdx := 0; shIdx < sCount; shIdx++ {
				assign[eIdx][dIdx][shIdx] = model.NewBoolVar().
					WithName(fmt.Sprintf("a_%d_%d_%d", eIdx, dIdx, shIdx))
			}
		}
	}

	for dIdx := 0; dIdx < dCount; dIdx++ {
		for shIdx := 0; shIdx < sCount; shIdx++ {
			expr := cpmodel.NewLinearExpr()
			for eIdx := 0; eIdx < eCount; eIdx++ {
				expr.Add(assign[eIdx][dIdx][shIdx])
			}
			model.AddLinearConstraint(expr, minStaffPerShift, maxStaffPerShift)
		}
	}

	for eIdx := 0; eIdx < eCount; eIdx++ {
		for dIdx := 0; dIdx < dCount; dIdx++ {
			expr := cpmodel.NewLinearExpr()
			for shIdx := 0; shIdx < sCount; shIdx++ {
				expr.Add(assign[eIdx][dIdx][shIdx])
			}
			model.AddLinearConstraint(expr, 0, 1)
		}
	}

	for eIdx := 0; eIdx < eCount; eIdx++ {
		for dIdx := 0; dIdx < dCount-1; dIdx++ {
			for shA := 0; shA < sCount; shA++ {
				endA := inputsShifts[shA].EndMinutes
				startA := inputsShifts[shA].StartMinutes
				endAbs := int64(dIdx*1440 + endA)
				if endA < startA {
					endAbs = int64(dIdx*1440 + endA + 1440)
				}
				for shB := 0; shB < sCount; shB++ {
					startAbs := int64((dIdx+1)*1440 + inputsShifts[shB].StartMinutes)
					rest := startAbs - endAbs
					if rest < minRestMinutes {
						model.AddLinearConstraint(
							cpmodel.NewLinearExpr().
								Add(assign[eIdx][dIdx][shA]).
								Add(assign[eIdx][dIdx+1][shB]),
							0,
							1,
						)
					}
				}
			}
		}
	}

	maxTotalMinutes := int64(7) * maxShiftMinutes
	overtimeWeight := maxTotalMinutes*int64(eCount) + 1
	objective := cpmodel.NewLinearExpr()
	for eIdx, emp := range inputsEmp {
		total := model.NewIntVar(0, maxTotalMinutes).WithName(fmt.Sprintf("total_%d", eIdx))
		totalExpr := cpmodel.NewLinearExpr()
		for dIdx := 0; dIdx < dCount; dIdx++ {
			for shIdx := 0; shIdx < sCount; shIdx++ {
				totalExpr.AddTerm(assign[eIdx][dIdx][shIdx], inputsShifts[shIdx].DurationMinutes)
			}
		}
		model.AddEquality(total, totalExpr)
		overtime := model.NewIntVar(0, maxTotalMinutes).WithName(fmt.Sprintf("ot_%d", eIdx))
		model.AddGreaterOrEqual(
			cpmodel.NewLinearExpr().Add(overtime),
			cpmodel.NewLinearExpr().Add(total).AddConstant(-emp.TargetMinutes),
		)
		objective.AddTerm(overtime, overtimeWeight)
		objective.AddTerm(total, 1)
	}
	model.Minimize(objective)

	modelProto, err := model.Model()
	if err != nil {
		return nil, err
	}
	params := &sppb.SatParameters{
		MaxTimeInSeconds: proto.Float64(maxSolveSeconds),
		NumSearchWorkers: proto.Int32(int32(runtime.NumCPU())),
	}
	res, err := cpmodel.SolveCpModelWithParameters(modelProto, params)
	if err != nil {
		return nil, err
	}

	status := res.GetStatus()
	statusStr := "infeasible"
	if status == cmpb.CpSolverStatus_OPTIMAL {
		statusStr = "optimal"
	} else if status == cmpb.CpSolverStatus_FEASIBLE {
		statusStr = "feasible"
	}
	if statusStr == "infeasible" {
		return &domain.AutoGenerateSchedulesResponse{
			Status:        "infeasible",
			PlanID:        uuid.New(),
			LocationID:    locationID,
			Timezone:      timezone,
			Week:          week,
			Year:          year,
			WeekStartDate: weekStartStr,
			Constraints: domain.SchedulePlanConstraints{
				MaxStaffPerShift: int32(maxStaffPerShift),
				AllowEmptyShift:  true,
			},
			Employees:      planEmployees,
			ShiftTemplates: shiftTemplates,
			Slots:          emptySlots,
			Summary:        []domain.ScheduleEmployeeSummary{},
			Warnings: []domain.SchedulePlanWarning{
				{
					Code:    "INFEASIBLE",
					Message: "Solver could not find a feasible schedule within the time limit.",
				},
			},
		}, nil
	}

	slots := make([]domain.SchedulePlanSlot, 0, dCount*sCount)
	assignedMinutesByEmp := make(map[uuid.UUID]int64, eCount)
	shiftCountsByEmp := make(map[uuid.UUID]map[uuid.UUID]int, eCount)
	for _, emp := range inputsEmp {
		shiftCountsByEmp[emp.ID] = make(map[uuid.UUID]int)
	}
	for dIdx := 0; dIdx < dCount; dIdx++ {
		dateStr := weekStart.AddDate(0, 0, dIdx).Format("2006-01-02")
		for shIdx, sh := range inputsShifts {
			employeeIDs := make([]uuid.UUID, 0, int(maxStaffPerShift))
			for eIdx, emp := range inputsEmp {
				if cpmodel.SolutionBooleanValue(res, assign[eIdx][dIdx][shIdx]) {
					employeeIDs = append(employeeIDs, emp.ID)
					assignedMinutesByEmp[emp.ID] += sh.DurationMinutes
					shiftCountsByEmp[emp.ID][sh.ID]++
				}
			}
			slots = append(
				slots,
				domain.SchedulePlanSlot{Date: dateStr, ShiftID: sh.ID, EmployeeIDs: employeeIDs},
			)
		}
	}

	summaries := make([]domain.ScheduleEmployeeSummary, 0, eCount)
	for _, emp := range inputsEmp {
		assigned := assignedMinutesByEmp[emp.ID]
		overtime := int64(0)
		if assigned > emp.TargetMinutes {
			overtime = assigned - emp.TargetMinutes
		}
		summaries = append(summaries, domain.ScheduleEmployeeSummary{
			EmployeeID:      emp.ID,
			TargetMinutes:   emp.TargetMinutes,
			AssignedMinutes: assigned,
			OvertimeMinutes: overtime,
			ShiftCounts:     shiftCountsByEmp[emp.ID],
		})
	}

	return &domain.AutoGenerateSchedulesResponse{
		Status:        statusStr,
		PlanID:        uuid.New(),
		LocationID:    locationID,
		Timezone:      timezone,
		Week:          week,
		Year:          year,
		WeekStartDate: weekStartStr,
		Constraints: domain.SchedulePlanConstraints{
			MaxStaffPerShift: int32(maxStaffPerShift),
			AllowEmptyShift:  true,
		},
		Employees:      planEmployees,
		ShiftTemplates: shiftTemplates,
		Slots:          slots,
		Summary:        summaries,
	}, nil
}

func (s *ScheduleService) SaveGeneratedSchedules(
	ctx context.Context,
	creatorID uuid.UUID,
	req *domain.SaveGeneratedSchedulesRequest,
) error {
	if len(req.Slots) == 0 {
		return nil
	}
	if req.PlanID == uuid.Nil {
		return fmt.Errorf("plan_id is required")
	}
	if req.LocationID == uuid.Nil {
		return fmt.Errorf("location_id is required")
	}
	if req.Week < 1 || req.Week > 53 {
		return fmt.Errorf("invalid week")
	}
	if req.Year <= 0 {
		return fmt.Errorf("invalid year")
	}

	location, err := s.repository.GetLocationByID(ctx, req.LocationID)
	if err != nil {
		s.logError(
			ctx,
			"SaveGeneratedSchedules",
			"failed to fetch location",
			err,
			zap.String("location_id", req.LocationID.String()),
		)
		return err
	}
	locationTZ, err := time.LoadLocation(location.Timezone)
	if err != nil {
		s.logError(
			ctx,
			"SaveGeneratedSchedules",
			"invalid location timezone",
			err,
			zap.String("timezone", location.Timezone),
		)
		return fmt.Errorf("invalid location timezone: %w", err)
	}
	if err := s.ensureWeekEmpty(ctx, req.LocationID, req.Week, req.Year, locationTZ); err != nil {
		return err
	}

	weekStart, err := isoWeekStartDate(int(req.Year), int(req.Week), locationTZ)
	if err != nil {
		return err
	}
	weekEnd := weekStart.AddDate(0, 0, 6)
	_ = weekEnd

	locationShifts, err := s.repository.GetShiftsByLocationID(ctx, req.LocationID)
	if err != nil {
		return err
	}
	shiftIDSet := make(map[uuid.UUID]struct{}, len(locationShifts))
	shiftByID := make(map[uuid.UUID]domain.ScheduleLocationShift, len(locationShifts))
	for _, sh := range locationShifts {
		shiftIDSet[sh.ID] = struct{}{}
		shiftByID[sh.ID] = sh
	}

	maxStaffPerShift := 2
	seenSlot := make(map[string]struct{}, len(req.Slots))
	for _, slot := range req.Slots {
		if slot.Date == "" {
			return fmt.Errorf("slot.date is required")
		}
		if slot.ShiftID == uuid.Nil {
			return fmt.Errorf("slot.shift_id is required")
		}
		if _, ok := shiftIDSet[slot.ShiftID]; !ok {
			return fmt.Errorf("slot.shift_id does not belong to location")
		}
		slotKey := slot.Date + ":" + slot.ShiftID.String()
		if _, ok := seenSlot[slotKey]; ok {
			return fmt.Errorf("duplicate slot: %s", slotKey)
		}
		seenSlot[slotKey] = struct{}{}

		date, err := time.ParseInLocation("2006-01-02", slot.Date, locationTZ)
		if err != nil {
			return fmt.Errorf("invalid slot.date format, expected YYYY-MM-DD")
		}
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, locationTZ)
		if dayStart.Before(weekStart) || dayStart.After(weekEnd) {
			return fmt.Errorf("slot.date is outside requested ISO week")
		}

		if len(slot.EmployeeIDs) > maxStaffPerShift {
			return fmt.Errorf("too many employees for slot (max %d)", maxStaffPerShift)
		}
		empSeen := make(map[uuid.UUID]struct{}, len(slot.EmployeeIDs))
		for _, eid := range slot.EmployeeIDs {
			if eid == uuid.Nil {
				return fmt.Errorf("slot.employee_ids contains invalid uuid")
			}
			if _, ok := empSeen[eid]; ok {
				return fmt.Errorf("slot.employee_ids must not contain duplicates")
			}
			empSeen[eid] = struct{}{}
		}
	}

	assignByEmpByDate := make(map[uuid.UUID]map[string]uuid.UUID)
	for _, slot := range req.Slots {
		for _, eid := range slot.EmployeeIDs {
			m, ok := assignByEmpByDate[eid]
			if !ok {
				m = make(map[string]uuid.UUID)
				assignByEmpByDate[eid] = m
			}
			if _, exists := m[slot.Date]; exists {
				return fmt.Errorf("employee has more than one shift on %s", slot.Date)
			}
			m[slot.Date] = slot.ShiftID
		}
	}

	dateList := make([]string, 0, 7)
	for d := 0; d < 7; d++ {
		dateList = append(dateList, weekStart.AddDate(0, 0, d).Format("2006-01-02"))
	}
	for _, byDate := range assignByEmpByDate {
		for d := 0; d < 6; d++ {
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
				return fmt.Errorf("minimum rest violation between %s and %s", curDate, nextDate)
			}
		}
	}

	for _, slot := range req.Slots {
		if len(slot.EmployeeIDs) == 0 {
			continue
		}
		shiftID := slot.ShiftID
		shiftDate := slot.Date
		if _, err := s.CreateSchedule(ctx, creatorID, &domain.CreateScheduleRequest{
			EmployeeIDs:     slot.EmployeeIDs,
			LocationID:      req.LocationID,
			IsCustom:        false,
			LocationShiftID: &shiftID,
			ShiftDate:       &shiftDate,
		}); err != nil {
			return err
		}
	}
	return nil
}
