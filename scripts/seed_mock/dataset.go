package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"hrbackend/internal/seed"

	"github.com/brianvoe/gofakeit/v7"
)

type generatedDataset struct {
	Organizations               []seed.OrganizationSeed
	Locations                   []seed.LocationSeed
	Departments                 []seed.DepartmentSeed
	Employees                   []seed.EmployeeSeed
	DepartmentHeads             []seed.DepartmentHeadSeed
	EmployeeContractChanges     []seed.EmployeeContractChangeSeed
	LeaveRequests               []seed.LeaveRequestSeed
	PayoutRequests              []seed.PayoutRequestSeed
	Schedules                   []seed.ScheduleSeed
	ShiftSwapRequests           []seed.ShiftSwapRequestSeed
	LateArrivals                []seed.LateArrivalSeed
	TimeEntries                 []seed.TimeEntrySeed
	PayPeriods                  []seed.PayPeriodSeed
	EmployeeHandbookAssignments []seed.EmployeeHandbookAssignmentSeed
}

type departmentTemplate struct {
	Alias         string
	Name          string
	HeadPosition  string
	StaffPosition string
	Description   string
}

type ortSampleOverride struct {
	IrregularHoursProfile string
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
	result.EmployeeContractChanges = make([]seed.EmployeeContractChangeSeed, 0, len(departments)*2)
	result.LeaveRequests = make([]seed.LeaveRequestSeed, 0, len(departments)*3)
	result.PayoutRequests = make([]seed.PayoutRequestSeed, 0, 3)
	result.Schedules = make([]seed.ScheduleSeed, 0, len(departments)*18)
	result.ShiftSwapRequests = make([]seed.ShiftSwapRequestSeed, 0, 4)
	result.LateArrivals = make([]seed.LateArrivalSeed, 0, len(departments)+2)
	result.TimeEntries = make([]seed.TimeEntrySeed, 0, len(departments)*6)
	result.PayPeriods = make([]seed.PayPeriodSeed, 0, 2)
	result.EmployeeHandbookAssignments = make([]seed.EmployeeHandbookAssignmentSeed, 0, len(departments)*3)
	ortOverrides := map[string]ortSampleOverride{
		"finance_head":        {IrregularHoursProfile: "roster"},
		"care_staff_01":       {IrregularHoursProfile: "non_roster"},
		"operations_staff_01": {IrregularHoursProfile: "none"},
		"planning_staff_02":   {IrregularHoursProfile: "none"},
	}

	for deptIdx, department := range departments {
		headAlias := fmt.Sprintf("%s_head", department.Alias)
		headLocationAlias := locationAliases[deptIdx%len(locationAliases)]
		headSeed := generateEmployeeSeed(
			headAlias,
			emailSuffix,
			passwordValue,
			headLocationAlias,
			department.Alias,
			nil,
			department.HeadPosition,
		)
		headSeed = applyORTOverride(headSeed, ortOverrides)
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
		result.EmployeeContractChanges = append(result.EmployeeContractChanges,
			buildContractChangeSeed(headAlias, "hr_head", result.Employees[len(result.Employees)-1], 4, 2),
		)
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
		headTimeEntries := []seed.TimeEntrySeed{
			buildApprovedTimeEntrySeed(fmt.Sprintf("%s_head_mon_entry", headAlias), fmt.Sprintf("%s_head_mon", headAlias), headAlias, headAlias, "07:30", "15:30", 30, "normal", nil),
			buildSubmittedTimeEntrySeed(fmt.Sprintf("%s_head_wed_entry", headAlias), fmt.Sprintf("%s_head_wed", headAlias), headAlias, headAlias, "15:00", "23:00", 30, "normal"),
		}
		if headAlias == "finance_head" {
			headTimeEntries[1] = buildApprovedTimeEntrySeed(
				fmt.Sprintf("%s_head_wed_entry", headAlias),
				fmt.Sprintf("%s_head_wed", headAlias),
				headAlias,
				headAlias,
				"19:00",
				"23:00",
				30,
				"normal",
				strPtr("Approved roster evening sample for ORT payroll seeding"),
			)
		}
		result.TimeEntries = append(result.TimeEntries, headTimeEntries...)
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
				&managerAlias,
				department.StaffPosition,
			)
			employeeSeed = applyORTOverride(employeeSeed, ortOverrides)
			result.Employees = append(result.Employees, employeeSeed)
			if empIdx < 2 {
				result.EmployeeHandbookAssignments = append(result.EmployeeHandbookAssignments, seed.EmployeeHandbookAssignmentSeed{
					EmployeeAlias:      employeeAlias,
					TemplateAlias:      templateAlias,
					ActorEmployeeAlias: &headAlias,
				})
			}
			if empIdx == 0 {
				result.EmployeeContractChanges = append(result.EmployeeContractChanges,
					buildContractChangeSeed(employeeAlias, headAlias, employeeSeed, 8, 3),
				)
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
				)
				result.TimeEntries = append(result.TimeEntries,
					buildApprovedTimeEntrySeed(
						fmt.Sprintf("%s_entry_a", employeeAlias),
						shiftAAlias,
						employeeAlias,
						headAlias,
						"07:30",
						"15:30",
						30,
						"normal",
						nil,
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
					result.TimeEntries = append(result.TimeEntries,
						buildApprovedTimeEntrySeed(
							fmt.Sprintf("%s_ort_weekday_entry", employeeAlias),
							ortShiftAlias,
							employeeAlias,
							headAlias,
							"20:00",
							"23:00",
							15,
							"normal",
							strPtr("Approved non-roster weekday evening sample for ORT payroll seeding"),
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
					result.TimeEntries = append(result.TimeEntries,
						buildApprovedTimeEntrySeed(
							fmt.Sprintf("%s_ort_saturday_entry", employeeAlias),
							ortShiftAlias,
							employeeAlias,
							headAlias,
							"21:00",
							"23:30",
							15,
							"overtime",
							strPtr("Approved Saturday evening and night sample for ORT payroll seeding"),
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
					result.TimeEntries = append(result.TimeEntries,
						buildApprovedTimeEntrySeed(
							fmt.Sprintf("%s_ort_sunday_entry", employeeAlias),
							ortShiftAlias,
							employeeAlias,
							headAlias,
							"12:00",
							"18:00",
							30,
							"travel",
							strPtr("Approved Sunday sample for ORT payroll seeding"),
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
				if empIdx == 0 {
					result.TimeEntries = append(result.TimeEntries,
						buildRejectedTimeEntrySeed(
							fmt.Sprintf("%s_entry_b", employeeAlias),
							shiftBAlias,
							employeeAlias,
							headAlias,
							"15:00",
							"22:45",
							15,
							"training",
							"Training time needs corrected classification",
						),
					)
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

	result.PayoutRequests = append(result.PayoutRequests,
		seed.PayoutRequestSeed{
			Alias:                          "finance_head_paid_payout",
			EmployeeAlias:                  "finance_head",
			BalanceAdjustedByEmployeeAlias: "hr_head",
			RequestedHours:                 6,
			BalanceYear:                    2026,
			Status:                         "paid",
			RequestNote:                    strPtr("Seeded extra-hours payout for payroll visibility"),
			DecisionByEmployeeAlias:        strPtr("hr_head"),
			PaidByEmployeeAlias:            strPtr("hr_head"),
			SalaryMonth:                    timePtr(time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)),
			DecisionNote:                   strPtr("Approved for July payroll"),
		},
		seed.PayoutRequestSeed{
			Alias:                          "hr_head_approved_payout",
			EmployeeAlias:                  "hr_head",
			BalanceAdjustedByEmployeeAlias: "hr_head",
			RequestedHours:                 4,
			BalanceYear:                    2026,
			Status:                         "approved",
			RequestNote:                    strPtr("Seeded approved payout request"),
			DecisionByEmployeeAlias:        strPtr("hr_head"),
			SalaryMonth:                    timePtr(time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)),
			DecisionNote:                   strPtr("Approved for August payroll"),
		},
		seed.PayoutRequestSeed{
			Alias:                          "operations_head_rejected_payout",
			EmployeeAlias:                  "operations_head",
			BalanceAdjustedByEmployeeAlias: "hr_head",
			RequestedHours:                 5,
			BalanceYear:                    2026,
			Status:                         "rejected",
			RequestNote:                    strPtr("Seeded rejected payout request"),
			DecisionByEmployeeAlias:        strPtr("hr_head"),
			DecisionNote:                   strPtr("Rejected pending staffing review"),
		},
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
			Alias:                 zzpAlias,
			FirstName:             gofakeit.FirstName(),
			LastName:              gofakeit.LastName(),
			UserEmail:             fmt.Sprintf("%s+%s@example.com", sanitizeEmailPart(zzpAlias), emailSuffix),
			UserPassword:          passwordValue,
			Bsn:                   gofakeit.Numerify("#########"),
			Street:                gofakeit.StreetName(),
			HouseNumber:           fmt.Sprintf("%d", gofakeit.Number(1, 300)),
			PostalCode:            gofakeit.Zip(),
			City:                  gofakeit.City(),
			Position:              strPtr(fmt.Sprintf("Freelance Consultant %d", zzpIdx+1)),
			Gender:                randomGender(),
			LocationAlias:         strPtr(zzpLocationAlias),
			DepartmentAlias:       nil,
			ManagerAlias:          nil,
			EmployeeNumber:        strPtr(fmt.Sprintf("ZZP-%05d", 1000+zzpIdx)),
			EmploymentNumber:      strPtr(fmt.Sprintf("FLC-%05d", 5000+zzpIdx)),
			RoleName:              strPtr("admin"),
			PrivatePhoneNumber:    strPtr(gofakeit.Phone()),
			WorkPhoneNumber:       strPtr(gofakeit.Phone()),
			ContractStartDate:     timePtr(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
			ContractEndDate:       nil,
			ContractHours:         float64Ptr(float64([]int{20, 25, 30}[gofakeit.Number(0, 2)])),
			ContractType:          "ZZP",
			ContractRate:          float64Ptr(float64(gofakeit.Number(25, 55))),
			IrregularHoursProfile: "none",
		}
		result.Employees = append(result.Employees, zzpSeed)

		// Add time entries for ZZP employees
		if zzpIdx < 2 {
			scheduleAlias := fmt.Sprintf("%s_schedule", zzpAlias)
			timeEntryAlias := fmt.Sprintf("%s_entry", zzpAlias)
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
			result.TimeEntries = append(result.TimeEntries,
				buildApprovedTimeEntrySeed(
					timeEntryAlias,
					scheduleAlias,
					zzpAlias,
					"hr_head",
					"09:00",
					"17:00",
					30,
					"normal",
					strPtr("Freelance project work"),
				),
			)
		}
	}

	return result
}

func generateEmployeeSeed(
	alias, emailSuffix, passwordValue, locationAlias, departmentAlias string,
	managerAlias *string,
	position string,
) seed.EmployeeSeed {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	contractStartDate := time.Now().AddDate(-gofakeit.Number(1, 8), -gofakeit.Number(0, 11), 0)
	return seed.EmployeeSeed{
		Alias:                 alias,
		FirstName:             firstName,
		LastName:              lastName,
		UserEmail:             fmt.Sprintf("%s+%s@example.com", sanitizeEmailPart(alias), emailSuffix),
		UserPassword:          passwordValue,
		Bsn:                   gofakeit.Numerify("#########"),
		Street:                gofakeit.StreetName(),
		HouseNumber:           fmt.Sprintf("%d", gofakeit.Number(1, 300)),
		PostalCode:            gofakeit.Zip(),
		City:                  gofakeit.City(),
		Position:              strPtr(position),
		Gender:                randomGender(),
		LocationAlias:         strPtr(locationAlias),
		DepartmentAlias:       strPtr(departmentAlias),
		ManagerAlias:          managerAlias,
		EmployeeNumber:        strPtr(gofakeit.Numerify("EMP-#####")),
		EmploymentNumber:      strPtr(gofakeit.Numerify("JOB-#####")),
		RoleName:              strPtr("admin"),
		PrivatePhoneNumber:    strPtr(gofakeit.Phone()),
		WorkPhoneNumber:       strPtr(gofakeit.Phone()),
		ContractStartDate:     &contractStartDate,
		ContractEndDate:       nil,
		ContractHours:         float64Ptr(float64([]int{24, 28, 32, 36}[gofakeit.Number(0, 3)])),
		ContractType:          "loondienst",
		ContractRate:          float64Ptr(float64(gofakeit.Number(18, 34))),
		IrregularHoursProfile: "none",
	}
}

func applyORTOverride(item seed.EmployeeSeed, overrides map[string]ortSampleOverride) seed.EmployeeSeed {
	override, ok := overrides[item.Alias]
	if !ok {
		return item
	}
	item.IrregularHoursProfile = override.IrregularHoursProfile
	return item
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

func buildContractChangeSeed(
	employeeAlias, actorAlias string,
	employee seed.EmployeeSeed,
	monthsAfterStart int,
	rateIncrease float64,
) seed.EmployeeContractChangeSeed {
	effectiveFrom := time.Now().UTC().AddDate(0, -monthsAfterStart, 0)
	if employee.ContractStartDate != nil {
		candidate := employee.ContractStartDate.AddDate(0, monthsAfterStart, 0)
		if candidate.Before(effectiveFrom) {
			effectiveFrom = candidate
		}
	}

	contractHours := 36.0
	if employee.ContractHours != nil {
		contractHours = math.Min(40, *employee.ContractHours+4)
	}

	var contractRate *float64
	if employee.ContractRate != nil {
		updatedRate := *employee.ContractRate + rateIncrease
		contractRate = &updatedRate
	}

	return seed.EmployeeContractChangeSeed{
		EmployeeAlias:         employeeAlias,
		ActorEmployeeAlias:    actorAlias,
		EffectiveFrom:         effectiveFrom,
		ContractHours:         contractHours,
		ContractType:          employee.ContractType,
		ContractRate:          contractRate,
		IrregularHoursProfile: employee.IrregularHoursProfile,
		ContractEndDate:       employee.ContractEndDate,
	}
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

func buildApprovedTimeEntrySeed(
	alias, scheduleAlias, employeeAlias, adminAlias, startTime, endTime string,
	breakMinutes int32,
	hourType string,
	notes *string,
) seed.TimeEntrySeed {
	return seed.TimeEntrySeed{
		Alias:                   alias,
		ScheduleAlias:           scheduleAlias,
		EmployeeAlias:           employeeAlias,
		Status:                  "approved",
		StartTime:               startTime,
		EndTime:                 endTime,
		BreakMinutes:            breakMinutes,
		HourType:                hourType,
		Notes:                   notes,
		ApprovedByEmployeeAlias: strPtr(adminAlias),
	}
}

func buildSubmittedTimeEntrySeed(
	alias, scheduleAlias, employeeAlias, adminAlias, startTime, endTime string,
	breakMinutes int32,
	hourType string,
) seed.TimeEntrySeed {
	return seed.TimeEntrySeed{
		Alias:                    alias,
		ScheduleAlias:            scheduleAlias,
		EmployeeAlias:            employeeAlias,
		Status:                   "submitted",
		StartTime:                startTime,
		EndTime:                  endTime,
		BreakMinutes:             breakMinutes,
		HourType:                 hourType,
		SubmittedByEmployeeAlias: strPtr(employeeAlias),
		ApprovedByEmployeeAlias:  strPtr(adminAlias),
	}
}

func buildRejectedTimeEntrySeed(
	alias, scheduleAlias, employeeAlias, adminAlias, startTime, endTime string,
	breakMinutes int32,
	hourType string,
	rejectionReason string,
) seed.TimeEntrySeed {
	return seed.TimeEntrySeed{
		Alias:                    alias,
		ScheduleAlias:            scheduleAlias,
		EmployeeAlias:            employeeAlias,
		Status:                   "rejected",
		StartTime:                startTime,
		EndTime:                  endTime,
		BreakMinutes:             breakMinutes,
		HourType:                 hourType,
		SubmittedByEmployeeAlias: strPtr(employeeAlias),
		ApprovedByEmployeeAlias:  strPtr(adminAlias),
		RejectionReason:          strPtr(rejectionReason),
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
