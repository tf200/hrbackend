package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hrbackend/internal/seed"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seedConfig struct {
	DBSource                    string
	Table                       string
	RunLabel                    string
	Profile                     seed.AppOrganizationProfileDefaults
	NationalHolidays            []seed.NationalHolidaySeed
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
	Handbooks                   []seed.HandbookTemplateSeed
	EmployeeHandbookAssignments []seed.EmployeeHandbookAssignmentSeed
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := loadConfigFromEnv()
	if err != nil {
		exitErr(err)
	}

	pool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		exitErr(fmt.Errorf("connect db: %w", err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		exitErr(fmt.Errorf("ping db: %w", err))
	}

	seeders := map[string]seed.Seeder{
		"app_organization_profile": seed.AppOrganizationProfileSeeder{
			Defaults: cfg.Profile,
		},
		"national_holidays": seed.NationalHolidaysSeeder{
			Holidays: cfg.NationalHolidays,
		},
		"organisations": seed.OrganisationsSeeder{
			Organizations: cfg.Organizations,
		},
		"location": seed.LocationSeeder{
			Locations: cfg.Locations,
		},
		"departments": seed.DepartmentsSeeder{
			Departments: cfg.Departments,
		},
		"employees": seed.EmployeesSeeder{
			Employees: cfg.Employees,
		},
		"department_heads": seed.DepartmentHeadsSeeder{
			Assignments: cfg.DepartmentHeads,
		},
		"employee_contract_changes": seed.EmployeeContractChangesSeeder{
			Employees: cfg.Employees,
			Changes:   cfg.EmployeeContractChanges,
		},
		"leave_requests": seed.LeaveRequestsSeeder{
			Requests: cfg.LeaveRequests,
		},
		"payout_requests": seed.PayoutRequestsSeeder{
			Requests: cfg.PayoutRequests,
		},
		"schedules": seed.SchedulesSeeder{
			Schedules: cfg.Schedules,
		},
		"shift_swap_requests": seed.ShiftSwapRequestsSeeder{
			Requests: cfg.ShiftSwapRequests,
		},
		"late_arrivals": seed.LateArrivalsSeeder{
			Arrivals: cfg.LateArrivals,
		},
		"time_entries": seed.TimeEntriesSeeder{
			Entries: cfg.TimeEntries,
		},
		"pay_periods": seed.PayPeriodsSeeder{
			Periods: cfg.PayPeriods,
		},
		"handbooks": seed.HandbooksSeeder{
			Templates: cfg.Handbooks,
		},
		"employee_handbook_assignments": seed.EmployeeHandbookAssignmentsSeeder{
			Assignments: cfg.EmployeeHandbookAssignments,
		},
	}
	dependencies := map[string][]string{
		"app_organization_profile":      {},
		"national_holidays":             {},
		"organisations":                 {},
		"location":                      {"organisations"},
		"departments":                   {},
		"employees":                     {"location", "departments"},
		"department_heads":              {"departments", "employees"},
		"employee_contract_changes":     {"employees"},
		"leave_requests":                {"employees", "employee_contract_changes"},
		"payout_requests":               {"employees", "employee_contract_changes"},
		"schedules":                     {"location", "employees"},
		"shift_swap_requests":           {"employees", "schedules"},
		"late_arrivals":                 {"employees", "schedules"},
		"time_entries":                  {"employees", "schedules"},
		"pay_periods":                   {"employees", "time_entries"},
		"handbooks":                     {"departments", "employees"},
		"employee_handbook_assignments": {"employees", "handbooks"},
	}
	runOrder := []string{
		"app_organization_profile",
		"national_holidays",
		"organisations",
		"location",
		"departments",
		"employees",
		"department_heads",
		"employee_contract_changes",
		"leave_requests",
		"payout_requests",
		"schedules",
		"time_entries",
		"pay_periods",
		"shift_swap_requests",
		"late_arrivals",
		"handbooks",
		"employee_handbook_assignments",
	}
	state := seed.NewState()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		exitErr(fmt.Errorf("begin tx: %w", err))
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	env := seed.Env{
		DB:    tx,
		State: state,
	}

	if cfg.Table == "all" {
		for _, key := range runOrder {
			seeder := seeders[key]
			if err := seeder.Seed(ctx, env); err != nil {
				exitErr(err)
			}
			fmt.Printf("Seed completed for table: %s\n", seeder.Name())
		}
		if err := tx.Commit(ctx); err != nil {
			exitErr(fmt.Errorf("commit tx: %w", err))
		}
		return
	}

	_, ok := seeders[cfg.Table]
	if !ok {
		exitErr(fmt.Errorf("unknown --table %q", cfg.Table))
	}

	executed := make(map[string]bool)
	var runWithDependencies func(key string) error
	runWithDependencies = func(key string) error {
		if executed[key] {
			return nil
		}

		for _, dep := range dependencies[key] {
			if err := runWithDependencies(dep); err != nil {
				return err
			}
		}

		currentSeeder, ok := seeders[key]
		if !ok {
			return fmt.Errorf("unknown seeder dependency %q", key)
		}
		if err := currentSeeder.Seed(ctx, env); err != nil {
			return err
		}
		executed[key] = true
		fmt.Printf("Seed completed for table: %s\n", currentSeeder.Name())
		return nil
	}

	if err := runWithDependencies(cfg.Table); err != nil {
		exitErr(err)
	}
	if err := tx.Commit(ctx); err != nil {
		exitErr(fmt.Errorf("commit tx: %w", err))
	}
}

func loadConfigFromEnv() (seedConfig, error) {
	table := flag.String("table", "all", "table seeder to run (or all)")
	flag.Parse()

	dbSource := firstNonEmpty(os.Getenv("MIGRATION_DB_SOURCE"), os.Getenv("DB_SOURCE"))
	if dbSource == "" {
		return seedConfig{}, errors.New("MIGRATION_DB_SOURCE or DB_SOURCE is required")
	}

	fakeSeed := time.Now().UnixNano()
	if seedFromEnv := strings.TrimSpace(os.Getenv("SEED_FAKE_SEED")); seedFromEnv != "" {
		parsedSeed, err := strconv.ParseInt(seedFromEnv, 10, 64)
		if err != nil {
			return seedConfig{}, fmt.Errorf("SEED_FAKE_SEED must be int64: %w", err)
		}
		fakeSeed = parsedSeed
	}
	gofakeit.Seed(fakeSeed)

	runLabel := os.Getenv("SEED_RUN_LABEL")
	dataset := buildGeneratedDataset(runLabel, fakeSeed)

	return seedConfig{
		DBSource: dbSource,
		Table:    *table,
		RunLabel: runLabel,
		Profile: seed.AppOrganizationProfileDefaults{
			Name:            envOrDefault("SEED_ORG_NAME", fmt.Sprintf("%s Profile", gofakeit.Company())),
			DefaultTimezone: envOrDefault("SEED_ORG_TIMEZONE", "Europe/Amsterdam"),
			Email:           optionalEnv("SEED_ORG_EMAIL"),
			PhoneNumber:     optionalEnv("SEED_ORG_PHONE"),
			Website:         optionalEnv("SEED_ORG_WEBSITE"),
			HQStreet:        optionalEnv("SEED_ORG_HQ_STREET"),
			HQHouseNumber:   optionalEnv("SEED_ORG_HQ_HOUSE_NUMBER"),
			HQHouseNumberAddition: optionalEnv(
				"SEED_ORG_HQ_HOUSE_NUMBER_ADDITION",
			),
			HQPostalCode: optionalEnv("SEED_ORG_HQ_POSTAL_CODE"),
			HQCity:       optionalEnv("SEED_ORG_HQ_CITY"),
		},
		NationalHolidays: []seed.NationalHolidaySeed{
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), Name: "Nieuwjaarsdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC), Name: "Goede vrijdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.April, 5, 0, 0, 0, 0, time.UTC), Name: "Eerste paasdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.April, 6, 0, 0, 0, 0, time.UTC), Name: "Tweede paasdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.April, 27, 0, 0, 0, 0, time.UTC), Name: "Koningsdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC), Name: "Bevrijdingsdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.May, 14, 0, 0, 0, 0, time.UTC), Name: "Hemelvaartsdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.May, 24, 0, 0, 0, 0, time.UTC), Name: "Eerste pinksterdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC), Name: "Tweede pinksterdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC), Name: "Eerste kerstdag", IsNational: true},
			{CountryCode: "NL", HolidayDate: time.Date(2026, time.December, 26, 0, 0, 0, 0, time.UTC), Name: "Tweede kerstdag", IsNational: true},
		},
		Organizations:           dataset.Organizations,
		Locations:               dataset.Locations,
		Departments:             dataset.Departments,
		Employees:               dataset.Employees,
		DepartmentHeads:         dataset.DepartmentHeads,
		EmployeeContractChanges: dataset.EmployeeContractChanges,
		LeaveRequests:           dataset.LeaveRequests,
		PayoutRequests:          dataset.PayoutRequests,
		Schedules:               dataset.Schedules,
		ShiftSwapRequests:       dataset.ShiftSwapRequests,
		LateArrivals:            dataset.LateArrivals,
		TimeEntries:             dataset.TimeEntries,
		PayPeriods:              dataset.PayPeriods,
		Handbooks: []seed.HandbookTemplateSeed{
			{
				Alias:              "care_baseline",
				DepartmentAlias:    "care",
				ActorEmployeeAlias: strPtr("care_head"),
				Title:              "Care Department Onboarding",
				Description:        strPtr("Baseline onboarding handbook for care employees."),
				Steps: []seed.HandbookStepSeed{
					{SortOrder: 1, Kind: "content", Title: "Welcome to Care", Body: strPtr("This handbook explains resident care expectations, escalation paths, and safe handover basics."), IsRequired: boolPtr(true)},
					{SortOrder: 2, Kind: "ack", Title: "Acknowledge Resident Safety Rules", Body: strPtr("Confirm that you understand resident identification, medication escalation, and incident reporting rules."), IsRequired: boolPtr(true)},
					{SortOrder: 3, Kind: "link", Title: "Read the Medication Protocol", Content: []byte(`{"url":"https://www.rijksoverheid.nl/"}`), IsRequired: boolPtr(true)},
				},
			},
			{
				Alias:              "operations_baseline",
				DepartmentAlias:    "operations",
				ActorEmployeeAlias: strPtr("operations_head"),
				Title:              "Operations Department Onboarding",
				Description:        strPtr("Baseline onboarding handbook for operations employees."),
				Steps: []seed.HandbookStepSeed{
					{SortOrder: 1, Kind: "content", Title: "Operations Workflow", Body: strPtr("This handbook covers opening checks, facility issues, and daily coordination responsibilities."), IsRequired: boolPtr(true)},
					{SortOrder: 2, Kind: "ack", Title: "Acknowledge Escalation Process", Body: strPtr("Confirm that you understand how to escalate urgent facility and staffing incidents."), IsRequired: boolPtr(true)},
					{SortOrder: 3, Kind: "quiz", Title: "Operations Basics Check", Content: []byte(`{"question":"Who should be notified first for an urgent building safety issue?","options":["A resident family member","The on-duty operations lead","The payroll team"],"correct_option_index":1}`), IsRequired: boolPtr(true)},
				},
			},
			{
				Alias:              "planning_baseline",
				DepartmentAlias:    "planning",
				ActorEmployeeAlias: strPtr("planning_head"),
				Title:              "Planning Department Onboarding",
				Description:        strPtr("Baseline onboarding handbook for planning employees."),
				Steps: []seed.HandbookStepSeed{
					{SortOrder: 1, Kind: "content", Title: "Roster Planning Standards", Body: strPtr("This handbook explains roster coverage, shift balance, and absence follow-up expectations."), IsRequired: boolPtr(true)},
					{SortOrder: 2, Kind: "ack", Title: "Acknowledge Coverage Rules", Body: strPtr("Confirm that you understand minimum coverage and handover timing requirements."), IsRequired: boolPtr(true)},
					{SortOrder: 3, Kind: "link", Title: "Review Scheduling Guidance", Content: []byte(`{"url":"https://www.rijksoverheid.nl/"}`), IsRequired: boolPtr(true)},
				},
			},
			{
				Alias:              "finance_baseline",
				DepartmentAlias:    "finance",
				ActorEmployeeAlias: strPtr("finance_head"),
				Title:              "Finance Department Onboarding",
				Description:        strPtr("Baseline onboarding handbook for finance employees."),
				Steps: []seed.HandbookStepSeed{
					{SortOrder: 1, Kind: "content", Title: "Payroll and Controls", Body: strPtr("This handbook covers payroll deadlines, approval checks, and payout control responsibilities."), IsRequired: boolPtr(true)},
					{SortOrder: 2, Kind: "ack", Title: "Acknowledge Payroll Controls", Body: strPtr("Confirm that you understand separation of duties and payment approval controls."), IsRequired: boolPtr(true)},
					{SortOrder: 3, Kind: "quiz", Title: "Finance Basics Check", Content: []byte(`{"question":"Which action best supports payroll control?","options":["Approving your own payout change","Reviewing source data before approval","Skipping exception review"],"correct_option_index":1}`), IsRequired: boolPtr(true)},
				},
			},
			{
				Alias:              "hr_baseline",
				DepartmentAlias:    "hr",
				ActorEmployeeAlias: strPtr("hr_head"),
				Title:              "HR Department Onboarding",
				Description:        strPtr("Baseline onboarding handbook for HR employees."),
				Steps: []seed.HandbookStepSeed{
					{SortOrder: 1, Kind: "content", Title: "HR Administration Standards", Body: strPtr("This handbook explains employee record handling, privacy expectations, and onboarding responsibilities."), IsRequired: boolPtr(true)},
					{SortOrder: 2, Kind: "ack", Title: "Acknowledge Privacy Rules", Body: strPtr("Confirm that you understand confidentiality and employee record access rules."), IsRequired: boolPtr(true)},
					{SortOrder: 3, Kind: "link", Title: "Review Government Leave Guidance", Content: []byte(`{"url":"https://www.rijksoverheid.nl/"}`), IsRequired: boolPtr(true)},
				},
			},
		},
		EmployeeHandbookAssignments: dataset.EmployeeHandbookAssignments,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func optionalEnv(key string) *string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func strPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "seed-mock: %v\n", err)
	os.Exit(1)
}
