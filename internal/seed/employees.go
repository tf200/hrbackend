package seed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository"
	dbrepo "hrbackend/internal/repository/db"
	"hrbackend/internal/service"
	"hrbackend/pkg/password"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EmployeeSeed struct {
	Alias              string
	FirstName          string
	LastName           string
	UserEmail          string
	UserPassword       string
	Bsn                string
	Street             string
	HouseNumber        string
	PostalCode         string
	City               string
	Gender             string
	ManagerAlias       *string
	EmployeeNumber     *string
	EmploymentNumber   *string
	RoleName           *string
	PrivatePhoneNumber *string
	WorkPhoneNumber    *string
	Contract           *EmployeeContractSeed
	SalaryAssignment   *EmployeeSalaryAssignmentSeed
}

type EmployeeContractSeed struct {
	JobTitle               string
	DepartmentAlias        string
	LocationAlias          string
	OrganizationalRoleName *string
	ContractType           string
	ContractHoursType      string
	StartDate              time.Time
	ContractEndDate        *time.Time
	HoursPerWeek           *float64
	MinHoursPerWeek        *float64
	MaxHoursPerWeek        *float64
	RosterFreeDay          string
	WageTaxTable           *string
}

type EmployeeSalaryAssignmentSeed struct {
	CAOCode       string
	Scale         int
	Step          string
	EffectiveFrom *time.Time
	EffectiveTo   *time.Time
}

type EmployeesSeeder struct {
	Employees []EmployeeSeed
}

func (s EmployeesSeeder) Name() string {
	return "employees"
}

func (s EmployeesSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Employees) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed employees: state is required")
	}
	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed employees: env DB must be pgx.Tx")
	}

	store := dbrepo.NewStoreWithTx(tx)
	employeeRepo := repository.NewEmployeeRepository(store)
	employeeService := service.NewEmployeeService(employeeRepo, nil, nil)
	roleIDs := make(map[string]uuid.UUID)
	seedCtx := context.WithValue(ctx, "employee_id", uuid.Nil)

	for _, item := range s.Employees {
		if strings.TrimSpace(item.Alias) == "" {
			return fmt.Errorf("seed employees: alias is required")
		}
		if strings.TrimSpace(item.UserEmail) == "" {
			return fmt.Errorf("seed employees[%s]: user email is required", item.Alias)
		}
		if strings.TrimSpace(item.UserPassword) == "" {
			return fmt.Errorf("seed employees[%s]: user password is required", item.Alias)
		}
		if strings.TrimSpace(item.FirstName) == "" || strings.TrimSpace(item.LastName) == "" {
			return fmt.Errorf("seed employees[%s]: first and last name are required", item.Alias)
		}
		if strings.TrimSpace(item.Gender) == "" {
			return fmt.Errorf("seed employees[%s]: gender is required", item.Alias)
		}

		roleID, err := resolveOptionalRoleID(ctx, env, roleIDs, item.RoleName)
		if err != nil {
			return fmt.Errorf("seed employees[%s]: %w", item.Alias, err)
		}

		userID, employeeID, err := ensureEmployee(
			seedCtx,
			store,
			employeeService,
			item,
			roleID,
		)
		if err != nil {
			return fmt.Errorf("seed employees[%s]: %w", item.Alias, err)
		}

		if roleID != nil {
			if err := store.AssignRoleToUser(seedCtx, dbrepo.AssignRoleToUserParams{
				UserID: userID,
				RoleID: *roleID,
			}); err != nil {
				return fmt.Errorf("seed employees[%s]: assign role: %w", item.Alias, err)
			}
		}

		if err := seedEmployeeDetails(seedCtx, env, employeeID, item); err != nil {
			return fmt.Errorf("seed employees[%s]: %w", item.Alias, err)
		}
		if err := ensureEmployeeContractAndSalary(seedCtx, env, employeeID, item); err != nil {
			return fmt.Errorf("seed employees[%s]: %w", item.Alias, err)
		}

		env.State.PutEmployee(item.Alias, employeeID)
	}

	for _, item := range s.Employees {
		if item.ManagerAlias == nil || strings.TrimSpace(*item.ManagerAlias) == "" {
			continue
		}

		employeeID, ok := env.State.EmployeeID(item.Alias)
		if !ok {
			return fmt.Errorf("seed employees[%s]: employee alias not found in state", item.Alias)
		}

		managerAlias := strings.TrimSpace(*item.ManagerAlias)
		if managerAlias == item.Alias {
			return fmt.Errorf("seed employees[%s]: manager alias cannot equal self", item.Alias)
		}

		managerID, ok := env.State.EmployeeID(managerAlias)
		if !ok {
			return fmt.Errorf("seed employees[%s]: manager alias %q not found in state", item.Alias, managerAlias)
		}

		managerIDCopy := managerID
		if _, err := employeeService.UpdateEmployee(seedCtx, employeeID, domain.UpdateEmployeeParams{
			ManagerEmployeeID: &managerIDCopy,
		}); err != nil {
			return fmt.Errorf("seed employees[%s]: set manager: %w", item.Alias, err)
		}
	}

	return nil
}

func ensureEmployee(
	ctx context.Context,
	store *dbrepo.Store,
	employeeService domain.EmployeeService,
	item EmployeeSeed,
	roleID *uuid.UUID,
) (uuid.UUID, uuid.UUID, error) {
	existingUser, err := store.GetUserByEmail(ctx, item.UserEmail)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, uuid.Nil, fmt.Errorf("lookup existing user by email: %w", err)
		}

		resolvedRoleID := uuid.Nil
		if roleID != nil {
			resolvedRoleID = *roleID
		}

		createdEmployee, err := employeeService.CreateEmployee(ctx, domain.CreateEmployeeParams{
			FirstName:           item.FirstName,
			LastName:            item.LastName,
			Bsn:                 item.Bsn,
			Street:              item.Street,
			HouseNumber:         item.HouseNumber,
			PostalCode:          item.PostalCode,
			City:                item.City,
			ManagerEmployeeID:   nil,
			EmployeeNumber:      item.EmployeeNumber,
			EmploymentNumber:    item.EmploymentNumber,
			PrivateEmailAddress: &item.UserEmail,
			WorkEmailAddress:    &item.UserEmail,
			PrivatePhoneNumber:  item.PrivatePhoneNumber,
			WorkPhoneNumber:     item.WorkPhoneNumber,
			Gender:              item.Gender,
			RoleID:              resolvedRoleID,
			UserEmail:           item.UserEmail,
			UserPassword:        item.UserPassword,
		})
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("create employee via service: %w", err)
		}

		return createdEmployee.UserID, createdEmployee.ID, nil
	}

	hashedPassword, err := password.HashPassword(item.UserPassword)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("hash password: %w", err)
	}
	if err := store.UpdatePassword(ctx, dbrepo.UpdatePasswordParams{
		ID:       existingUser.ID,
		Password: hashedPassword,
	}); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("update existing user password: %w", err)
	}

	if _, err := employeeService.UpdateEmployee(ctx, existingUser.EmployeeID, domain.UpdateEmployeeParams{
		FirstName:           &item.FirstName,
		LastName:            &item.LastName,
		EmployeeNumber:      item.EmployeeNumber,
		EmploymentNumber:    item.EmploymentNumber,
		PrivateEmailAddress: &item.UserEmail,
		PrivatePhoneNumber:  item.PrivatePhoneNumber,
		WorkPhoneNumber:     item.WorkPhoneNumber,
		Gender:              &item.Gender,
	}); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("update existing employee via service: %w", err)
	}

	return existingUser.ID, existingUser.EmployeeID, nil
}

func resolveOptionalRoleID(
	ctx context.Context,
	env Env,
	cache map[string]uuid.UUID,
	roleName *string,
) (*uuid.UUID, error) {
	if roleName == nil || strings.TrimSpace(*roleName) == "" {
		return nil, nil
	}

	name := strings.TrimSpace(*roleName)
	if id, ok := cache[name]; ok {
		return &id, nil
	}

	var roleID uuid.UUID
	if err := env.DB.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, name).Scan(&roleID); err != nil {
		return nil, fmt.Errorf("resolve role %q: %w", name, err)
	}
	cache[name] = roleID
	return &roleID, nil
}

func seedEmployeeDetails(ctx context.Context, env Env, employeeID uuid.UUID, item EmployeeSeed) error {
	if _, err := env.DB.Exec(ctx, `DELETE FROM employee_education WHERE employee_id = $1`, employeeID); err != nil {
		return fmt.Errorf("reset employee_education: %w", err)
	}
	if _, err := env.DB.Exec(ctx, `DELETE FROM employee_qualifications WHERE employee_id = $1`, employeeID); err != nil {
		return fmt.Errorf("reset employee_qualifications: %w", err)
	}
	if _, err := env.DB.Exec(ctx, `DELETE FROM employee_experience WHERE employee_id = $1`, employeeID); err != nil {
		return fmt.Errorf("reset employee_experience: %w", err)
	}

	educationCount := gofakeit.Number(0, 2)
	for i := 0; i < educationCount; i++ {
		startDate := time.Now().AddDate(-gofakeit.Number(4, 10), 0, 0)
		endDate := startDate.AddDate(gofakeit.Number(1, 4), 0, 0)
		if _, err := env.DB.Exec(ctx, `
			INSERT INTO employee_education (
				employee_id,
				institution_name,
				degree,
				field_of_study,
				start_date,
				end_date
			) VALUES ($1, $2, $3, $4, $5, $6)
		`, employeeID,
			fmt.Sprintf("%s Institute", gofakeit.Company()),
			randomFrom([]string{"MBO", "HBO", "Bachelor", "Associate Degree"}),
			randomFrom([]string{"Healthcare", "Management", "Administration", "Social Work"}),
			startDate, endDate,
		); err != nil {
			return fmt.Errorf("seed employee education: %w", err)
		}
	}

	qualificationCodes := []string{
		"company_emergency_response_certificate",
		"first_aid_diploma",
		"cpr_resuscitation_certificate",
		"big_registration_nurse",
	}
	qualificationCount := gofakeit.Number(0, 2)
	seeded := map[string]bool{}
	for i := 0; i < qualificationCount; i++ {
		code := qualificationCodes[gofakeit.Number(0, len(qualificationCodes)-1)]
		if seeded[code] {
			continue
		}
		seeded[code] = true
		achievedOn := time.Now().AddDate(-gofakeit.Number(1, 5), 0, 0)
		var expirationDate *time.Time
		if code != "big_registration_nurse" {
			d := achievedOn.AddDate(gofakeit.Number(1, 2), 0, 0)
			expirationDate = &d
		}
		certNumber := fmt.Sprintf("CERT-%d", gofakeit.Number(10000, 99999))
		var qualificationID uuid.UUID
		if err := env.DB.QueryRow(ctx, `SELECT id FROM qualifications WHERE code = $1`, code).Scan(&qualificationID); err != nil {
			return fmt.Errorf("lookup qualification code %q: %w", code, err)
		}
		if _, err := env.DB.Exec(ctx, `
		INSERT INTO employee_qualifications (
			employee_id,
			qualification_id,
			achieved_on,
			expiration_date,
			certificate_number
		) VALUES ($1, $2, $3, $4, $5)
	`, employeeID,
			qualificationID,
			achievedOn,
			expirationDate,
			strPtr(certNumber),
		); err != nil {
			return fmt.Errorf("seed employee qualification: %w", err)
		}
	}

	experienceCount := gofakeit.Number(1, 3)
	for i := 0; i < experienceCount; i++ {
		startDate := time.Now().AddDate(-gofakeit.Number(3, 12), 0, 0)
		endDate := startDate.AddDate(gofakeit.Number(1, 4), 0, 0)
		if _, err := env.DB.Exec(ctx, `
			INSERT INTO employee_experience (
				employee_id,
				job_title,
				company_name,
				start_date,
				end_date,
				description
			) VALUES ($1, $2, $3, $4, $5, $6)
		`, employeeID,
			gofakeit.JobTitle(),
			gofakeit.Company(),
			startDate,
			endDate,
			strPtr(gofakeit.Sentence(10)),
		); err != nil {
			return fmt.Errorf("seed employee experience: %w", err)
		}
	}

	return nil
}

func ensureEmployeeContractAndSalary(ctx context.Context, env Env, employeeID uuid.UUID, item EmployeeSeed) error {
	if item.Contract == nil {
		return nil
	}

	departmentID, ok := env.State.DepartmentID(item.Contract.DepartmentAlias)
	if !ok {
		return fmt.Errorf("contract department alias %q not found in seed state", item.Contract.DepartmentAlias)
	}
	locationID, ok := env.State.LocationID(item.Contract.LocationAlias)
	if !ok {
		return fmt.Errorf("contract location alias %q not found in seed state", item.Contract.LocationAlias)
	}

	var organizationalRoleID *uuid.UUID
	if item.Contract.OrganizationalRoleName != nil && strings.TrimSpace(*item.Contract.OrganizationalRoleName) != "" {
		var id uuid.UUID
		if err := env.DB.QueryRow(ctx, `SELECT id FROM organizational_roles WHERE name = $1`, strings.TrimSpace(*item.Contract.OrganizationalRoleName)).Scan(&id); err != nil {
			return fmt.Errorf("resolve organizational role %q: %w", *item.Contract.OrganizationalRoleName, err)
		}
		organizationalRoleID = &id
	}

	if _, err := env.DB.Exec(ctx, `
		DELETE FROM employee_salary_assignments WHERE employee_id = $1
	`, employeeID); err != nil {
		return fmt.Errorf("delete salary assignments: %w", err)
	}
	if _, err := env.DB.Exec(ctx, `
		DELETE FROM employee_contracts WHERE employee_id = $1
	`, employeeID); err != nil {
		return fmt.Errorf("delete contracts: %w", err)
	}

	var contractID uuid.UUID
	err := env.DB.QueryRow(ctx, `
		INSERT INTO employee_contracts (
			employee_id,
			job_title,
			department_id,
			location_id,
			organizational_role_id,
			contract_type,
			contract_hours_type,
			start_date,
			contract_end_date,
			hours_per_week,
			min_hours_per_week,
			max_hours_per_week,
			roster_free_day,
			wage_tax_table,
			contract_event_type
		)
		VALUES ($1, $2::employee_job_title_enum, $3, $4, $5, $6::employee_contract_type_enum, $7::contract_hours_type_enum, $8, $9, $10, $11, $12, $13, $14::wage_tax_table_enum, 'initial'::employee_contract_event_type_enum)
		RETURNING id
	`, employeeID, item.Contract.JobTitle, departmentID, locationID, organizationalRoleID,
		item.Contract.ContractType, item.Contract.ContractHoursType, item.Contract.StartDate,
		item.Contract.ContractEndDate, item.Contract.HoursPerWeek, item.Contract.MinHoursPerWeek,
		item.Contract.MaxHoursPerWeek, item.Contract.RosterFreeDay, item.Contract.WageTaxTable).Scan(&contractID)
	if err != nil {
		return fmt.Errorf("insert employee contract: %w", err)
	}

	if item.SalaryAssignment == nil {
		return nil
	}

	caoCode := strings.TrimSpace(item.SalaryAssignment.CAOCode)
	if caoCode == "" {
		caoCode = "CAO_JEUGDZORG"
	}
	effectiveFrom := item.SalaryAssignment.EffectiveFrom
	if effectiveFrom == nil {
		effectiveFrom = &item.Contract.StartDate
	}

	var salaryScaleStepID uuid.UUID
	if err := env.DB.QueryRow(ctx, `
		SELECT css.id
		FROM cao_salary_scale_steps css
		JOIN cao_salary_tables cst ON cst.id = css.salary_table_id
		WHERE cst.cao_code = $1
		  AND css.scale = $2
		  AND css.step = $3
		  AND cst.effective_from <= $4
		  AND (cst.effective_to IS NULL OR cst.effective_to > $4)
		ORDER BY cst.effective_from DESC
		LIMIT 1
	`, caoCode, item.SalaryAssignment.Scale, item.SalaryAssignment.Step, *effectiveFrom).Scan(&salaryScaleStepID); err != nil {
		return fmt.Errorf("resolve salary scale step %s scale %d step %s: %w", caoCode, item.SalaryAssignment.Scale, item.SalaryAssignment.Step, err)
	}

	_, err = env.DB.Exec(ctx, `
		INSERT INTO employee_salary_assignments (
			employee_id,
			contract_id,
			salary_scale_step_id,
			effective_from,
			effective_to
		)
		VALUES ($1, $2, $3, $4, $5)
	`, employeeID, contractID, salaryScaleStepID, effectiveFrom, item.SalaryAssignment.EffectiveTo)
	if err != nil {
		return fmt.Errorf("insert salary assignment: %w", err)
	}

	return nil
}

func randomFrom(values []string) string {
	return values[gofakeit.Number(0, len(values)-1)]
}

func strPtr(value string) *string {
	return &value
}
