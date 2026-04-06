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
	Alias                 string
	FirstName             string
	LastName              string
	UserEmail             string
	UserPassword          string
	Bsn                   string
	Street                string
	HouseNumber           string
	PostalCode            string
	City                  string
	Position              *string
	Gender                string
	LocationAlias         *string
	DepartmentAlias       *string
	ManagerAlias          *string
	EmployeeNumber        *string
	EmploymentNumber      *string
	RoleName              *string
	PrivatePhoneNumber    *string
	WorkPhoneNumber       *string
	ContractStartDate     *time.Time
	ContractEndDate       *time.Time
	ContractHours         *float64
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
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
	employeeService := service.NewEmployeeService(employeeRepo, nil)
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

		locationID, err := resolveOptionalLocationAlias(env, item.LocationAlias, item.Alias)
		if err != nil {
			return err
		}
		departmentID, err := resolveOptionalDepartmentAlias(env, item.DepartmentAlias, item.Alias)
		if err != nil {
			return err
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
			locationID,
			departmentID,
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
	locationID, departmentID *uuid.UUID,
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
			FirstName:             item.FirstName,
			LastName:              item.LastName,
			Bsn:                   item.Bsn,
			Street:                item.Street,
			HouseNumber:           item.HouseNumber,
			PostalCode:            item.PostalCode,
			City:                  item.City,
			Position:              item.Position,
			DepartmentID:          departmentID,
			ManagerEmployeeID:     nil,
			EmployeeNumber:        item.EmployeeNumber,
			EmploymentNumber:      item.EmploymentNumber,
			PrivateEmailAddress:   &item.UserEmail,
			WorkEmailAddress:      &item.UserEmail,
			PrivatePhoneNumber:    item.PrivatePhoneNumber,
			WorkPhoneNumber:       item.WorkPhoneNumber,
			Gender:                item.Gender,
			LocationID:            locationID,
			ContractHours:         item.ContractHours,
			ContractType:          item.ContractType,
			ContractStartDate:     item.ContractStartDate,
			ContractEndDate:       item.ContractEndDate,
			ContractRate:          item.ContractRate,
			IrregularHoursProfile: item.IrregularHoursProfile,
			RoleID:                resolvedRoleID,
			UserEmail:             item.UserEmail,
			UserPassword:          item.UserPassword,
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
		Position:            item.Position,
		DepartmentID:        departmentID,
		EmployeeNumber:      item.EmployeeNumber,
		EmploymentNumber:    item.EmploymentNumber,
		PrivateEmailAddress: &item.UserEmail,
		PrivatePhoneNumber:  item.PrivatePhoneNumber,
		WorkPhoneNumber:     item.WorkPhoneNumber,
		Gender:              &item.Gender,
		LocationID:          locationID,
	}); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("update existing employee via service: %w", err)
	}

	if err := syncExistingEmployeeContract(ctx, store, employeeService, existingUser.EmployeeID, item); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("sync existing employee contract via service: %w", err)
	}

	return existingUser.ID, existingUser.EmployeeID, nil
}

func syncExistingEmployeeContract(
	ctx context.Context,
	store *dbrepo.Store,
	employeeService domain.EmployeeService,
	employeeID uuid.UUID,
	item EmployeeSeed,
) error {
	changeCount, err := store.CountEmployeeContractChanges(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("count contract changes: %w", err)
	}
	if changeCount > 0 {
		return nil
	}

	if item.ContractStartDate != nil {
		var contractEndDate time.Time
		if item.ContractEndDate != nil {
			contractEndDate = *item.ContractEndDate
		}

		if _, err := employeeService.AddContractDetails(ctx, employeeID, domain.AddContractDetailsParams{
			ContractHours:         item.ContractHours,
			ContractStartDate:     *item.ContractStartDate,
			ContractEndDate:       contractEndDate,
			ContractRate:          item.ContractRate,
			IrregularHoursProfile: item.IrregularHoursProfile,
		}); err != nil {
			return fmt.Errorf("add contract details: %w", err)
		}
	}

	switch item.ContractType {
	case "loondienst":
		if _, err := employeeService.UpdateIsSubcontractor(ctx, employeeID, domain.UpdateIsSubcontractorParams{
			IsSubcontractor: false,
		}); err != nil {
			return fmt.Errorf("set contract type loondienst: %w", err)
		}
	case "ZZP":
		if _, err := employeeService.UpdateIsSubcontractor(ctx, employeeID, domain.UpdateIsSubcontractorParams{
			IsSubcontractor: true,
		}); err != nil {
			return fmt.Errorf("set contract type ZZP: %w", err)
		}
	}

	return nil
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
	if _, err := env.DB.Exec(ctx, `DELETE FROM certification WHERE employee_id = $1`, employeeID); err != nil {
		return fmt.Errorf("reset certification: %w", err)
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

	certificationCount := gofakeit.Number(0, 2)
	for i := 0; i < certificationCount; i++ {
		if _, err := env.DB.Exec(ctx, `
			INSERT INTO certification (
				employee_id,
				name,
				issued_by,
				date_issued
			) VALUES ($1, $2, $3, $4)
		`, employeeID,
			randomFrom([]string{"BHV", "Medication Safety", "First Aid", "Care Compliance"}),
			gofakeit.Company(),
			time.Now().AddDate(-gofakeit.Number(1, 5), 0, 0),
		); err != nil {
			return fmt.Errorf("seed employee certification: %w", err)
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

func randomFrom(values []string) string {
	return values[gofakeit.Number(0, len(values)-1)]
}

func strPtr(value string) *string {
	return &value
}

func resolveOptionalLocationAlias(env Env, alias *string, employeeAlias string) (*uuid.UUID, error) {
	if alias == nil || strings.TrimSpace(*alias) == "" {
		return nil, nil
	}
	id, ok := env.State.LocationID(strings.TrimSpace(*alias))
	if !ok {
		return nil, fmt.Errorf(
			"seed employees[%s]: missing location alias %q in seed state",
			employeeAlias,
			strings.TrimSpace(*alias),
		)
	}
	return &id, nil
}

func resolveOptionalDepartmentAlias(env Env, alias *string, employeeAlias string) (*uuid.UUID, error) {
	if alias == nil || strings.TrimSpace(*alias) == "" {
		return nil, nil
	}
	id, ok := env.State.DepartmentID(strings.TrimSpace(*alias))
	if !ok {
		return nil, fmt.Errorf(
			"seed employees[%s]: missing department alias %q in seed state",
			employeeAlias,
			strings.TrimSpace(*alias),
		)
	}
	return &id, nil
}
