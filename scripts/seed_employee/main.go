package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"hrbackend/pkg/password"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seedConfig struct {
	DBSource         string
	EmployeeEmail    string
	EmployeePassword string
	Profile          employeeProfileDefaults
}

type employeeProfileDefaults struct {
	FirstName           string
	LastName            string
	BSN                 string
	Street              string
	HouseNumber         string
	HouseNumberAddition string
	PostalCode          string
	City                string
	Position            string
	PrivateEmail        string
	WorkEmail           string
	PrivatePhone        string
	WorkPhone           string
	HomeTelephone       string
	Gender              string
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

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		exitErr(fmt.Errorf("begin tx: %w", err))
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	employeeRoleID, err := ensureEmployeeRole(ctx, tx)
	if err != nil {
		exitErr(err)
	}

	grantedCount, err := syncEmployeeRolePermissions(ctx, tx, employeeRoleID)
	if err != nil {
		exitErr(err)
	}

	userID, createdUser, err := ensureEmployeeUser(ctx, tx, cfg.EmployeeEmail, cfg.EmployeePassword)
	if err != nil {
		exitErr(err)
	}

	createdProfile, err := ensureEmployeeProfile(ctx, tx, userID, cfg)
	if err != nil {
		exitErr(err)
	}

	if err := ensureUserRole(ctx, tx, userID, employeeRoleID); err != nil {
		exitErr(err)
	}

	if err := tx.Commit(ctx); err != nil {
		exitErr(fmt.Errorf("commit tx: %w", err))
	}

	fmt.Println("Employee seed completed.")
	fmt.Printf("Employee email: %s\n", cfg.EmployeeEmail)
	fmt.Printf("User created: %t\n", createdUser)
	fmt.Printf("Profile created: %t\n", createdProfile)
	fmt.Printf("Employee permission links ensured: %d\n", grantedCount)
}

func ensureEmployeeRole(ctx context.Context, tx pgx.Tx) (uuid.UUID, error) {
	if _, err := tx.Exec(ctx, `
		INSERT INTO roles (name, description)
		VALUES ('employee', 'Standard employee with self-service access')
		ON CONFLICT (name) DO NOTHING
	`); err != nil {
		return uuid.Nil, fmt.Errorf("ensure employee role: %w", err)
	}

	var roleID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'employee'`).
		Scan(&roleID); err != nil {
		return uuid.Nil, fmt.Errorf("read employee role id: %w", err)
	}
	return roleID, nil
}

// employeePermissions defines the set of self-service permissions granted to
// the standard employee role.
var employeePermissions = []string{
	"PORTAL.EMPLOYEE.ACCESS",
	"EMPLOYEE.VIEW",
	"HANDBOOK.SELF.VIEW",
	"HANDBOOK.SELF.UPDATE",
	"LATE_ARRIVAL.CREATE",
	"LEAVE.REQUEST.CREATE",
	"LEAVE.REQUEST.VIEW",
	"PAYOUT.REQUEST.CREATE",
	"PAYOUT.REQUEST.VIEW",
	"EXPENSE.REQUEST.CREATE",
	"EXPENSE.REQUEST.VIEW",
	"SCHEDULE.VIEW",
	"SCHEDULE_SWAP.REQUEST",
	"SCHEDULE_SWAP.RESPOND",
	"SCHEDULE_SWAP.VIEW",
	"SHIFT.VIEW",
	"TIME_ENTRY.CREATE",
	"TIME_ENTRY.VIEW",
	"TRAINING.CATALOG.VIEW",
	"TRAINING.ASSIGNMENTS.VIEW",
	"PERFORMANCE.ASSESSMENT.VIEW",
}

func syncEmployeeRolePermissions(ctx context.Context, tx pgx.Tx, roleID uuid.UUID) (int64, error) {
	var count int64
	for _, name := range employeePermissions {
		tag, err := tx.Exec(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT $1, id FROM permissions WHERE name = $2
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`, roleID, name)
		if err != nil {
			return 0, fmt.Errorf("grant permission %q to employee role: %w", name, err)
		}
		count += tag.RowsAffected()
	}
	return count, nil
}

func ensureEmployeeUser(
	ctx context.Context,
	tx pgx.Tx,
	email, plainPassword string,
) (uuid.UUID, bool, error) {
	var userID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM custom_user WHERE email = $1`, email).Scan(&userID)
	if err == nil {
		return userID, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, fmt.Errorf("lookup employee user: %w", err)
	}

	hashedPassword, err := password.HashPassword(plainPassword)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("hash employee password: %w", err)
	}

	if err := tx.QueryRow(ctx, `
		INSERT INTO custom_user (password, email, is_active, profile_picture)
		VALUES ($1, $2, true, NULL)
		RETURNING id
	`, hashedPassword, email).Scan(&userID); err != nil {
		return uuid.Nil, false, fmt.Errorf("create employee user: %w", err)
	}

	return userID, true, nil
}

func ensureEmployeeProfile(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	cfg seedConfig,
) (bool, error) {
	var profileID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM employee_profile WHERE user_id = $1`, userID).
		Scan(&profileID)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("lookup employee profile: %w", err)
	}

	profile := cfg.Profile
	if profile.PrivateEmail == "" {
		profile.PrivateEmail = cfg.EmployeeEmail
	}
	if profile.WorkEmail == "" {
		profile.WorkEmail = cfg.EmployeeEmail
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO employee_profile (
			user_id,
			first_name,
			last_name,
			bsn,
			street,
			house_number,
			house_number_addition,
			postal_code,
			city,
			position,
			private_email_address,
			work_email_address,
			private_phone_number,
			work_phone_number,
			home_telephone_number,
			gender
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, NULLIF($10, ''),
			NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''), NULLIF($14, ''), NULLIF($15, ''), $16::gender_enum
		)
	`, userID, profile.FirstName, profile.LastName, profile.BSN, profile.Street, profile.HouseNumber,
		profile.HouseNumberAddition, profile.PostalCode, profile.City, profile.Position,
		profile.PrivateEmail, profile.WorkEmail, profile.PrivatePhone, profile.WorkPhone,
		profile.HomeTelephone, profile.Gender); err != nil {
		return false, fmt.Errorf("create employee profile: %w", err)
	}

	return true, nil
}

func ensureUserRole(ctx context.Context, tx pgx.Tx, userID, roleID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET role_id = EXCLUDED.role_id
	`, userID, roleID); err != nil {
		return fmt.Errorf("assign employee role to user: %w", err)
	}
	return nil
}

func loadConfigFromEnv() (seedConfig, error) {
	dbSource := strings.TrimSpace(os.Getenv("MIGRATION_DB_SOURCE"))
	employeeEmail := strings.TrimSpace(os.Getenv("EMPLOYEE_EMAIL"))
	employeePassword := strings.TrimSpace(os.Getenv("EMPLOYEE_PASSWORD"))

	if dbSource == "" {
		return seedConfig{}, errors.New("MIGRATION_DB_SOURCE is required")
	}
	if employeeEmail == "" {
		return seedConfig{}, errors.New("EMPLOYEE_EMAIL is required")
	}
	if employeePassword == "" {
		return seedConfig{}, errors.New("EMPLOYEE_PASSWORD is required")
	}

	profile := employeeProfileDefaults{
		FirstName:           envOrDefault("EMPLOYEE_FIRST_NAME", "Test"),
		LastName:            envOrDefault("EMPLOYEE_LAST_NAME", "Employee"),
		BSN:                 envOrDefault("EMPLOYEE_BSN", "111111111"),
		Street:              envOrDefault("EMPLOYEE_STREET", "Werkstraat"),
		HouseNumber:         envOrDefault("EMPLOYEE_HOUSE_NUMBER", "42"),
		HouseNumberAddition: envOrDefault("EMPLOYEE_HOUSE_NUMBER_ADDITION", ""),
		PostalCode:          envOrDefault("EMPLOYEE_POSTAL_CODE", "2000BB"),
		City:                envOrDefault("EMPLOYEE_CITY", "Rotterdam"),
		Position:            envOrDefault("EMPLOYEE_POSITION", "Medewerker"),
		PrivateEmail:        strings.TrimSpace(os.Getenv("EMPLOYEE_PRIVATE_EMAIL")),
		WorkEmail:           strings.TrimSpace(os.Getenv("EMPLOYEE_WORK_EMAIL")),
		PrivatePhone:        strings.TrimSpace(os.Getenv("EMPLOYEE_PRIVATE_PHONE")),
		WorkPhone:           strings.TrimSpace(os.Getenv("EMPLOYEE_WORK_PHONE")),
		HomeTelephone:       strings.TrimSpace(os.Getenv("EMPLOYEE_HOME_TELEPHONE")),
		Gender:              normalizeGender(envOrDefault("EMPLOYEE_GENDER", "unknown")),
	}

	return seedConfig{
		DBSource:         dbSource,
		EmployeeEmail:    employeeEmail,
		EmployeePassword: employeePassword,
		Profile:          profile,
	}, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func normalizeGender(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "male", "female", "other", "unknown":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "unknown"
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "seed-employee: %v\n", err)
	os.Exit(1)
}
