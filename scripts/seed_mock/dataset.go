package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hrbackend/internal/seed"

	"github.com/brianvoe/gofakeit/v7"
)

type generatedDataset struct {
	Organizations               []seed.OrganizationSeed
	SalaryTables                []seed.SalaryTableSeed
	Locations                   []seed.LocationSeed
	Departments                 []seed.DepartmentSeed
	Employees                   []seed.EmployeeSeed
	DepartmentHeads             []seed.DepartmentHeadSeed
	LeaveRequests               []seed.LeaveRequestSeed
	PayoutRequests              []seed.PayoutRequestSeed
	Schedules                   []seed.ScheduleSeed
	ShiftSwapRequests           []seed.ShiftSwapRequestSeed
	LateArrivals                []seed.LateArrivalSeed
	PayPeriods                  []seed.PayPeriodSeed
	EmployeeHandbookAssignments []seed.EmployeeHandbookAssignmentSeed
	PerformanceAssessments      []seed.PerformanceSeed
}

type departmentTemplate struct {
	Alias         string
	Name          string
	HeadPosition  string
	StaffPosition string
	Description   string
}

func buildGeneratedDataset(runLabel string, fakeSeed int64) generatedDataset {
	orgCount := intEnvOrDefault("SEED_ORGANIZATION_COUNT", 2)
	locationsPerOrg := intEnvOrDefault("SEED_LOCATIONS_PER_ORG", 3)
	employeesPerDepartment := intEnvOrDefault("SEED_EMPLOYEES_PER_DEPARTMENT", 6)
	passwordValue := envOrDefault("SEED_EMP_DEFAULT_PASSWORD", "ChangeMe123!")

	emailSuffix := strings.TrimSpace(runLabel)
	if emailSuffix == "" {
		emailSuffix = fmt.Sprintf("r%d", fakeSeed%1000000)
	}

	departments := []departmentTemplate{
		{Alias: "care", Name: "Care", HeadPosition: "Care Team Lead", StaffPosition: "Care Worker", Description: "Primary care and resident support."},
		{Alias: "operations", Name: "Operations", HeadPosition: "Operations Lead", StaffPosition: "Operations Coordinator", Description: "Daily operational coordination."},
		{Alias: "planning", Name: "Planning", HeadPosition: "Planning Lead", StaffPosition: "Scheduler", Description: "Roster and planning management."},
		{Alias: "finance", Name: "Finance", HeadPosition: "Finance Lead", StaffPosition: "Finance Officer", Description: "Payroll and financial administration."},
		{Alias: "hr", Name: "HR", HeadPosition: "HR Lead", StaffPosition: "HR Officer", Description: "Employee administration and onboarding."},
	}

	result := generatedDataset{
		Organizations: make([]seed.OrganizationSeed, 0, orgCount),
		SalaryTables:  buildCaoJeugdzorgSalaryTables(),
		Locations:     make([]seed.LocationSeed, 0, orgCount*locationsPerOrg),
		Departments:   make([]seed.DepartmentSeed, 0, len(departments)),
	}

	locationAliases := make([]string, 0, orgCount*locationsPerOrg)
	for orgIdx := 0; orgIdx < orgCount; orgIdx++ {
		orgAlias := fmt.Sprintf("org_%02d", orgIdx+1)
		orgCity := gofakeit.City()
		orgName := fmt.Sprintf("%s Care Group", gofakeit.Company())
		if runLabel != "" {
			orgName = fmt.Sprintf("%s (%s)", orgName, runLabel)
		}

		result.Organizations = append(result.Organizations, seed.OrganizationSeed{
			Alias:       orgAlias,
			Name:        orgName,
			Street:      gofakeit.StreetName(),
			HouseNumber: fmt.Sprintf("%d", gofakeit.Number(1, 400)),
			PostalCode:  gofakeit.Zip(),
			City:        orgCity,
			PhoneNumber: strPtr(gofakeit.Phone()),
			Email:       strPtr(fmt.Sprintf("contact+%s@%s.example", orgAlias, emailSuffix)),
		})

		for locIdx := 0; locIdx < locationsPerOrg; locIdx++ {
			locationAlias := fmt.Sprintf("%s_location_%02d", orgAlias, locIdx+1)
			locationAliases = append(locationAliases, locationAlias)
			locationType := "care_home"
			locationName := fmt.Sprintf("%s Residence", gofakeit.LastName())
			if locIdx == 0 {
				locationType = "office"
				locationName = fmt.Sprintf("%s Office", orgCity)
			}

			result.Locations = append(result.Locations, seed.LocationSeed{
				Alias:             locationAlias,
				OrganizationAlias: orgAlias,
				Name:              locationName,
				Street:            gofakeit.StreetName(),
				HouseNumber:       fmt.Sprintf("%d", gofakeit.Number(1, 500)),
				PostalCode:        gofakeit.Zip(),
				City:              orgCity,
				Timezone:          "Europe/Amsterdam",
				LocationType:      locationType,
			})
		}
	}

	for _, department := range departments {
		description := department.Description
		result.Departments = append(result.Departments, seed.DepartmentSeed{
			Alias:       department.Alias,
			Name:        department.Name,
			Description: &description,
		})
	}

	result.Employees = make([]seed.EmployeeSeed, 0, len(departments)*(employeesPerDepartment+1))
	result.DepartmentHeads = make([]seed.DepartmentHeadSeed, 0, len(departments))
	result.LeaveRequests = make([]seed.LeaveRequestSeed, 0, len(departments)*3)
	result.PayoutRequests = make([]seed.PayoutRequestSeed, 0, 3)
	result.Schedules = make([]seed.ScheduleSeed, 0, len(departments)*18)
	result.ShiftSwapRequests = make([]seed.ShiftSwapRequestSeed, 0, 4)
	result.LateArrivals = make([]seed.LateArrivalSeed, 0, len(departments)+2)
	result.PayPeriods = make([]seed.PayPeriodSeed, 0, 2)
	result.EmployeeHandbookAssignments = make([]seed.EmployeeHandbookAssignmentSeed, 0, len(departments)*3)
	result.PerformanceAssessments = make([]seed.PerformanceSeed, 0, 6)

	for deptIdx, department := range departments {
		headAlias := fmt.Sprintf("%s_head", department.Alias)
		headLocationAlias := locationAliases[deptIdx%len(locationAliases)]
		headSeed := generateEmployeeSeed(
			headAlias,
			emailSuffix,
			passwordValue,
			headLocationAlias,
			department.Alias,
			department.HeadPosition,
			true,
			deptIdx,
			nil,
		)
		result.Employees = append(result.Employees, headSeed)
		result.DepartmentHeads = append(result.DepartmentHeads, seed.DepartmentHeadSeed{
			DepartmentAlias: department.Alias,
			EmployeeAlias:   headAlias,
		})
		templateAlias := fmt.Sprintf("%s_baseline", department.Alias)
		result.EmployeeHandbookAssignments = append(result.EmployeeHandbookAssignments, seed.EmployeeHandbookAssignmentSeed{
			EmployeeAlias:      headAlias,
			TemplateAlias:      templateAlias,
			ActorEmployeeAlias: &headAlias,
		})
		result.LeaveRequests = append(result.LeaveRequests, buildApprovedVacationLeaveSeed(
			fmt.Sprintf("%s_head_vacation", department.Alias),
			headAlias,
			headAlias,
			"hr_head",
			deptIdx,
		))
		result.Schedules = append(result.Schedules,
			buildPresetScheduleSeed(fmt.Sprintf("%s_head_mon", headAlias), headAlias, headLocationAlias, headAlias, 1, time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)),
			buildPresetScheduleSeed(fmt.Sprintf("%s_head_wed", headAlias), headAlias, headLocationAlias, headAlias, 2, time.Date(2026, time.July, 8, 0, 0, 0, 0, time.UTC)),
			buildPresetScheduleSeed(fmt.Sprintf("%s_head_fri", headAlias), headAlias, headLocationAlias, headAlias, 1, time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)),
		)
		result.LateArrivals = append(result.LateArrivals, seed.LateArrivalSeed{
			Alias:                  fmt.Sprintf("%s_late_mon", headAlias),
			EmployeeAlias:          headAlias,
			CreatedByEmployeeAlias: strPtr(headAlias),
			ArrivalDate:            time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC),
			ArrivalTime:            "08:05",
			Reason:                 "Seeded late arrival after traffic delay",
		})

		for empIdx := 0; empIdx < employeesPerDepartment; empIdx++ {
			employeeAlias := fmt.Sprintf("%s_staff_%02d", department.Alias, empIdx+1)
			managerAlias := headAlias
			locationAlias := locationAliases[(deptIdx+empIdx)%len(locationAliases)]
			employeeSeed := generateEmployeeSeed(
				employeeAlias,
				emailSuffix,
				passwordValue,
				locationAlias,
				department.Alias,
				department.StaffPosition,
				false,
				empIdx,
				&managerAlias,
			)
			result.Employees = append(result.Employees, employeeSeed)
			if empIdx < 2 {
				result.EmployeeHandbookAssignments = append(result.EmployeeHandbookAssignments, seed.EmployeeHandbookAssignmentSeed{
					EmployeeAlias:      employeeAlias,
					TemplateAlias:      templateAlias,
					ActorEmployeeAlias: &headAlias,
				})
			}
			if empIdx == 0 {
				result.LeaveRequests = append(result.LeaveRequests, buildPendingPersonalLeaveSeed(
					fmt.Sprintf("%s_pending_personal", employeeAlias),
					employeeAlias,
					deptIdx,
				))
			}
			if empIdx == 1 {
				result.LeaveRequests = append(result.LeaveRequests, buildRejectedUnpaidLeaveSeed(
					fmt.Sprintf("%s_rejected_unpaid", employeeAlias),
					employeeAlias,
					employeeAlias,
					headAlias,
					deptIdx,
				))
			}
			if empIdx < 2 {
				shiftAAlias := fmt.Sprintf("%s_shift_a", employeeAlias)
				shiftBAlias := fmt.Sprintf("%s_shift_b", employeeAlias)
				swapShiftAlias := fmt.Sprintf("%s_swap_shift", employeeAlias)
				swapShiftExtraAlias := fmt.Sprintf("%s_swap_shift_extra", employeeAlias)
				baseDay := 7 + (empIdx * 2)
				shiftASlot := int16(1 + ((deptIdx + empIdx) % 2))
				shiftBSlot := int16(2 + ((deptIdx + empIdx) % 2))
				result.Schedules = append(result.Schedules,
					buildPresetScheduleSeed(
						shiftAAlias,
						employeeAlias,
						locationAlias,
						headAlias,
						shiftASlot,
						time.Date(2026, time.July, baseDay, 0, 0, 0, 0, time.UTC),
					),
					buildPresetScheduleSeed(
						shiftBAlias,
						employeeAlias,
						locationAlias,
						headAlias,
						shiftBSlot,
						time.Date(2026, time.July, baseDay+2, 0, 0, 0, 0, time.UTC),
					),
					buildPresetScheduleSeed(
						swapShiftAlias,
						employeeAlias,
						locationAlias,
						headAlias,
						1,
						time.Date(2026, time.July, 20+deptIdx, 0, 0, 0, 0, time.UTC),
					),
					buildPresetScheduleSeed(
						swapShiftExtraAlias,
						employeeAlias,
						locationAlias,
						headAlias,
						2,
						time.Date(2026, time.July, 25+deptIdx, 0, 0, 0, 0, time.UTC),
					),
				)
				switch employeeAlias {
				case "care_staff_01":
					ortShiftAlias := fmt.Sprintf("%s_ort_weekday", employeeAlias)
					result.Schedules = append(result.Schedules,
						buildPresetScheduleSeed(
							ortShiftAlias,
							employeeAlias,
							locationAlias,
							headAlias,
							2,
							time.Date(2026, time.July, 13, 0, 0, 0, 0, time.UTC),
						),
					)
				case "operations_staff_01":
					ortShiftAlias := fmt.Sprintf("%s_ort_saturday", employeeAlias)
					result.Schedules = append(result.Schedules,
						buildPresetScheduleSeed(
							ortShiftAlias,
							employeeAlias,
							locationAlias,
							headAlias,
							3,
							time.Date(2026, time.July, 18, 0, 0, 0, 0, time.UTC),
						),
					)
				case "planning_staff_02":
					ortShiftAlias := fmt.Sprintf("%s_ort_sunday", employeeAlias)
					result.Schedules = append(result.Schedules,
						buildPresetScheduleSeed(
							ortShiftAlias,
							employeeAlias,
							locationAlias,
							headAlias,
							2,
							time.Date(2026, time.July, 19, 0, 0, 0, 0, time.UTC),
						),
					)
				}
				if empIdx == 0 {
					result.LateArrivals = append(result.LateArrivals, seed.LateArrivalSeed{
						Alias:                  fmt.Sprintf("%s_late_shift_a", employeeAlias),
						EmployeeAlias:          employeeAlias,
						CreatedByEmployeeAlias: strPtr(headAlias),
						ArrivalDate:            time.Date(2026, time.July, baseDay, 0, 0, 0, 0, time.UTC),
						ArrivalTime:            lateArrivalTimeForShiftSlot(shiftASlot),
						Reason:                 "Seeded late arrival reported by department lead",
					})
				}
			}
		}
	}

	result.ShiftSwapRequests = append(result.ShiftSwapRequests,
		buildShiftSwapRequestSeed(
			"care_swap_confirmed",
			"care_staff_01",
			"care_staff_02",
			"care_staff_01_swap_shift",
			"care_staff_02_swap_shift",
			"confirmed",
			timePtr(time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)),
			strPtr("Works better with my family schedule"),
			strPtr("hr_head"),
			strPtr("Approved after recipient confirmation"),
		),
		buildShiftSwapRequestSeed(
			"operations_swap_admin_rejected",
			"operations_staff_01",
			"operations_staff_02",
			"operations_staff_01_swap_shift",
			"operations_staff_02_swap_shift",
			"admin_rejected",
			timePtr(time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)),
			strPtr("Please take my morning slot"),
			strPtr("hr_head"),
			strPtr("Rejected to preserve role coverage"),
		),
		buildShiftSwapRequestSeed(
			"planning_swap_pending_admin",
			"planning_staff_01",
			"planning_staff_02",
			"planning_staff_01_swap_shift",
			"planning_staff_02_swap_shift",
			"pending_admin",
			timePtr(time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)),
			strPtr("I can cover the early shift instead"),
			nil,
			nil,
		),
		buildShiftSwapRequestSeed(
			"planning_swap_waiting_admin_approval",
			"planning_staff_02",
			"planning_staff_01",
			"planning_staff_02_swap_shift_extra",
			"planning_staff_01_swap_shift_extra",
			"pending_admin",
			timePtr(time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)),
			strPtr("Agreed on both sides, waiting for admin approval"),
			nil,
			nil,
		),
		buildShiftSwapRequestSeed(
			"finance_swap_recipient_rejected",
			"finance_staff_01",
			"finance_staff_02",
			"finance_staff_01_swap_shift",
			"finance_staff_02_swap_shift",
			"recipient_rejected",
			timePtr(time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)),
			strPtr("I need to keep my current assignment"),
			nil,
			nil,
		),
		buildShiftSwapRequestSeed(
			"finance_swap_fully_approved",
			"finance_staff_02",
			"finance_staff_01",
			"finance_staff_02_swap_shift_extra",
			"finance_staff_01_swap_shift_extra",
			"confirmed",
			timePtr(time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)),
			strPtr("Both parties agreed to the swap"),
			strPtr("hr_head"),
			strPtr("Approved after both employees confirmed the change"),
		),
		buildShiftSwapRequestSeed(
			"hr_swap_pending_recipient",
			"hr_staff_01",
			"hr_staff_02",
			"hr_staff_01_swap_shift",
			"hr_staff_02_swap_shift",
			"pending_recipient",
			timePtr(time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)),
			nil,
			nil,
			nil,
		),
	)



	result.PayPeriods = append(result.PayPeriods,
		seed.PayPeriodSeed{
			Alias:                  "finance_head_july_paid",
			EmployeeAlias:          "finance_head",
			CreatedByEmployeeAlias: "hr_head",
			PaidByEmployeeAlias:    strPtr("hr_head"),
			Status:                 "paid",
			PeriodStart:            time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:              time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC),
		},
		seed.PayPeriodSeed{
			Alias:                  "care_staff_01_july_draft",
			EmployeeAlias:          "care_staff_01",
			CreatedByEmployeeAlias: "hr_head",
			Status:                 "draft",
			PeriodStart:            time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:              time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC),
		},
	)

	// Add ZZP (Freelance/Self-employed) employees
	zzpLocationAlias := locationAliases[0]
	for zzpIdx := 0; zzpIdx < 3; zzpIdx++ {
		zzpAlias := fmt.Sprintf("zzp_contractor_%02d", zzpIdx+1)
		zzpSeed := seed.EmployeeSeed{
			Alias:              zzpAlias,
			FirstName:          gofakeit.FirstName(),
			LastName:           gofakeit.LastName(),
			UserEmail:          fmt.Sprintf("%s+%s@example.com", sanitizeEmailPart(zzpAlias), emailSuffix),
			UserPassword:       passwordValue,
			Bsn:                gofakeit.Numerify("#########"),
			Street:             gofakeit.StreetName(),
			HouseNumber:        fmt.Sprintf("%d", gofakeit.Number(1, 300)),
			PostalCode:         gofakeit.Zip(),
			City:               gofakeit.City(),
			Gender:             randomGender(),
			ManagerAlias:       nil,
			EmployeeNumber:     strPtr(fmt.Sprintf("ZZP-%05d", 1000+zzpIdx)),
			EmploymentNumber:   strPtr(fmt.Sprintf("FLC-%05d", 5000+zzpIdx)),
			RoleName:           strPtr("employee"),
			PrivatePhoneNumber: strPtr(gofakeit.Phone()),
			WorkPhoneNumber:    strPtr(gofakeit.Phone()),
			Contract: &seed.EmployeeContractSeed{
				JobTitle:        "administrative_employee",
				DepartmentAlias: "hr",
				LocationAlias:   zzpLocationAlias,
				ContractType:    "on_call",
				StartDate:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				RosterFreeDay:   "saturday",
				WageTaxTable:    strPtr("white_table"),
			},
		}
		result.Employees = append(result.Employees, zzpSeed)

		// Add schedules for ZZP employees
		if zzpIdx < 2 {
			scheduleAlias := fmt.Sprintf("%s_schedule", zzpAlias)
			result.Schedules = append(result.Schedules,
				buildPresetScheduleSeed(
					scheduleAlias,
					zzpAlias,
					zzpLocationAlias,
					"hr_head",
					1,
					time.Date(2026, time.July, 8+zzpIdx, 0, 0, 0, 0, time.UTC),
				),
			)
		}
	}

	for i, employee := range result.Employees {
		if i >= 6 {
			break
		}
		result.PerformanceAssessments = append(result.PerformanceAssessments, seed.PerformanceSeed{
			EmployeeAlias: employee.Alias,
		})
	}

	return result
}

func generateEmployeeSeed(
	alias, emailSuffix, passwordValue string,
	locationAlias, departmentAlias, position string,
	isHead bool,
	index int,
	managerAlias *string,
) seed.EmployeeSeed {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	contract := buildEmployeeContractSeed(locationAlias, departmentAlias, position, isHead)
	return seed.EmployeeSeed{
		Alias:              alias,
		FirstName:          firstName,
		LastName:           lastName,
		UserEmail:          fmt.Sprintf("%s+%s@example.com", sanitizeEmailPart(alias), emailSuffix),
		UserPassword:       passwordValue,
		Bsn:                gofakeit.Numerify("#########"),
		Street:             gofakeit.StreetName(),
		HouseNumber:        fmt.Sprintf("%d", gofakeit.Number(1, 300)),
		PostalCode:         gofakeit.Zip(),
		City:               gofakeit.City(),
		Gender:             randomGender(),
		ManagerAlias:       managerAlias,
		EmployeeNumber:     strPtr(gofakeit.Numerify("EMP-#####")),
		EmploymentNumber:   strPtr(gofakeit.Numerify("JOB-#####")),
		RoleName:           strPtr("employee"),
		PrivatePhoneNumber: strPtr(gofakeit.Phone()),
		WorkPhoneNumber:    strPtr(gofakeit.Phone()),
		Contract:           &contract,
		SalaryAssignment:   buildEmployeeSalaryAssignmentSeed(departmentAlias, isHead, index),
	}
}

func buildEmployeeContractSeed(locationAlias, departmentAlias, position string, isHead bool) seed.EmployeeContractSeed {
	jobTitle := employeeJobTitleForDepartment(departmentAlias, isHead)
	if strings.Contains(strings.ToLower(position), "lead") {
		jobTitle = "team_lead"
	}
	hoursPerWeek := 36.0
	if !isHead && departmentAlias == "care" {
		hoursPerWeek = 32.0
	}

	return seed.EmployeeContractSeed{
		JobTitle:        jobTitle,
		DepartmentAlias: departmentAlias,
		LocationAlias:   locationAlias,
		ContractType:    "permanent",
		StartDate:       time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		HoursPerWeek:    float64Ptr(hoursPerWeek),
		RosterFreeDay:   "saturday",
		WageTaxTable:    strPtr("white_table"),
	}
}

func employeeJobTitleForDepartment(departmentAlias string, isHead bool) string {
	if isHead {
		return "team_lead"
	}
	switch departmentAlias {
	case "care":
		return "youth_worker_d"
	case "operations":
		return "care_coordinator"
	case "planning", "finance", "hr":
		return "administrative_employee"
	default:
		return "pedagogical_worker"
	}
}

func buildEmployeeSalaryAssignmentSeed(departmentAlias string, isHead bool, index int) *seed.EmployeeSalaryAssignmentSeed {
	scale := 8
	step := fmt.Sprintf("%d", 3+(index%5))
	if isHead {
		scale = 11
		step = "5"
	} else {
		switch departmentAlias {
		case "care":
			scale = 7
		case "operations", "planning", "finance", "hr":
			scale = 8
		}
	}
	effectiveFrom := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	return &seed.EmployeeSalaryAssignmentSeed{
		CAOCode:       "CAO_JEUGDZORG",
		Scale:         scale,
		Step:          step,
		EffectiveFrom: &effectiveFrom,
	}
}

func intEnvOrDefault(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func randomGender() string {
	options := []string{"male", "female", "other"}
	return options[gofakeit.Number(0, len(options)-1)]
}

func float64Ptr(value float64) *float64 {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func buildApprovedVacationLeaveSeed(
	alias, employeeAlias, createdByAlias, decisionByAlias string,
	departmentIndex int,
) seed.LeaveRequestSeed {
	startDate := time.Date(2026, time.June, 9+(departmentIndex*7), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)
	return seed.LeaveRequestSeed{
		Alias:                   alias,
		EmployeeAlias:           employeeAlias,
		CreatedByEmployeeAlias:  strPtr(createdByAlias),
		DecisionByEmployeeAlias: strPtr(decisionByAlias),
		LeaveType:               "vacation",
		Status:                  "approved",
		StartDate:               startDate,
		EndDate:                 endDate,
		Reason:                  strPtr("Planned summer leave"),
		DecisionNote:            strPtr("Approved for baseline staffing coverage"),
	}
}

func buildPendingPersonalLeaveSeed(alias, employeeAlias string, departmentIndex int) seed.LeaveRequestSeed {
	startDate := time.Date(2026, time.September, 8+(departmentIndex*3), 0, 0, 0, 0, time.UTC)
	return seed.LeaveRequestSeed{
		Alias:         alias,
		EmployeeAlias: employeeAlias,
		LeaveType:     "personal",
		Status:        "pending",
		StartDate:     startDate,
		EndDate:       startDate,
		Reason:        strPtr("Personal appointment"),
	}
}

func buildRejectedUnpaidLeaveSeed(
	alias, employeeAlias, createdByAlias, decisionByAlias string,
	departmentIndex int,
) seed.LeaveRequestSeed {
	startDate := time.Date(2026, time.November, 3+(departmentIndex*4), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)
	return seed.LeaveRequestSeed{
		Alias:                   alias,
		EmployeeAlias:           employeeAlias,
		CreatedByEmployeeAlias:  strPtr(createdByAlias),
		DecisionByEmployeeAlias: strPtr(decisionByAlias),
		LeaveType:               "unpaid",
		Status:                  "rejected",
		StartDate:               startDate,
		EndDate:                 endDate,
		Reason:                  strPtr("Extended personal travel request"),
		DecisionNote:            strPtr("Rejected to preserve staffing coverage"),
	}
}

func buildPresetScheduleSeed(
	alias, employeeAlias, locationAlias, createdByAlias string,
	shiftSlot int16,
	shiftDate time.Time,
) seed.ScheduleSeed {
	return seed.ScheduleSeed{
		Alias:                  alias,
		EmployeeAlias:          employeeAlias,
		LocationAlias:          locationAlias,
		CreatedByEmployeeAlias: createdByAlias,
		IsCustom:               false,
		ShiftSlot:              shiftSlot,
		ShiftDate:              shiftDate,
	}
}

func buildShiftSwapRequestSeed(
	alias, requesterEmployeeAlias, recipientEmployeeAlias, requesterScheduleAlias, recipientScheduleAlias, status string,
	expiresAt *time.Time,
	recipientResponseNote *string,
	adminEmployeeAlias *string,
	adminDecisionNote *string,
) seed.ShiftSwapRequestSeed {
	return seed.ShiftSwapRequestSeed{
		Alias:                  alias,
		RequesterEmployeeAlias: requesterEmployeeAlias,
		RecipientEmployeeAlias: recipientEmployeeAlias,
		RequesterScheduleAlias: requesterScheduleAlias,
		RecipientScheduleAlias: recipientScheduleAlias,
		Status:                 status,
		ExpiresAt:              expiresAt,
		RecipientResponseNote:  recipientResponseNote,
		AdminEmployeeAlias:     adminEmployeeAlias,
		AdminDecisionNote:      adminDecisionNote,
	}
}

func lateArrivalTimeForShiftSlot(slot int16) string {
	switch slot {
	case 1:
		return "08:10"
	case 2:
		return "15:20"
	case 3:
		return "20:20"
	default:
		return "08:10"
	}
}

func sanitizeEmailPart(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", ".")
	normalized = strings.ReplaceAll(normalized, "'", "")
	normalized = strings.ReplaceAll(normalized, "\"", "")
	normalized = strings.ReplaceAll(normalized, "..", ".")
	if normalized == "" {
		return "user"
	}
	return normalized
}
